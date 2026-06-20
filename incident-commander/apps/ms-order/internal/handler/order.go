package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/model"
)

type OrderHandler struct {
	cfg    config.Config
	db     *sql.DB
	rdb    *redis.Client
	writer *kafka.Writer
}

func NewOrderHandler(cfg config.Config, db *sql.DB, rdb *redis.Client, writer *kafka.Writer) *OrderHandler {
	return &OrderHandler{cfg: cfg, db: db, rdb: rdb, writer: writer}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	span, ctx := tracer.StartSpanFromContext(r.Context(), "order.create")
	defer span.Finish()

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" || req.ProductID == "" || req.Quantity <= 0 {
		http.Error(w, `{"error":"customer_id, product_id, and quantity are required"}`, http.StatusBadRequest)
		return
	}

	span.SetTag("customer_id", req.CustomerID)
	span.SetTag("product_id", req.ProductID)
	span.SetTag("quantity", req.Quantity)

	order := model.Order{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	// Save to PostgreSQL
	if err := h.saveOrder(ctx, order); err != nil {
		log.Printf("[handler] save order failed: %v", err)
		http.Error(w, `{"error":"failed to create order"}`, http.StatusInternalServerError)
		return
	}

	// Cache to Redis
	if err := h.cacheOrder(ctx, order); err != nil {
		log.Printf("[handler] WARN redis cache failed (non-fatal): %v", err)
	}

	// Publish to Kafka
	if err := h.publishOrderCreated(ctx, order); err != nil {
		log.Printf("[handler] publish event failed: %v", err)
		http.Error(w, `{"error":"failed to publish order"}`, http.StatusInternalServerError)
		return
	}

	resp := model.CreateOrderResponse{
		OrderID: order.ID,
		Status:  order.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) saveOrder(ctx context.Context, order model.Order) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "postgres.query")
	defer span.Finish()

	span.SetTag("order.id", order.ID)
	span.SetTag("db.type", "postgres")

	_, err := h.db.ExecContext(ctx,
		`INSERT INTO orders (id, customer_id, product_id, quantity, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		order.ID, order.CustomerID, order.ProductID, order.Quantity, order.Status, order.CreatedAt,
	)
	return err
}

func (h *OrderHandler) cacheOrder(ctx context.Context, order model.Order) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "redis.cache")
	defer span.Finish()

	span.SetTag("redis.key", "order:"+order.ID)

	cached := model.CachedOrder{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		ProductID:  order.ProductID,
		Quantity:   order.Quantity,
		Status:     order.Status,
	}
	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return h.rdb.Set(ctx, "order:"+order.ID, data, 5*time.Minute).Err()
}

func (h *OrderHandler) publishOrderCreated(ctx context.Context, order model.Order) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.produce")
	defer span.Finish()

	event := model.OrderCreatedEvent{
		OrderID:   order.ID,
		ProductID: order.ProductID,
		Quantity:  order.Quantity,
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	span.SetTag("kafka.topic", "order-created")
	span.SetTag("kafka.key", order.ID)

	err = h.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order.ID),
		Value: msg,
	})
	if err != nil {
		return err
	}

	log.Printf("[handler] order-created published: order_id=%s", order.ID)
	return nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, event model.PaymentCompletedEvent) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "order.update_status")
	defer span.Finish()

	span.SetTag("order.id", event.OrderID)
	span.SetTag("payment.status", event.Status)

	newStatus := "COMPLETED"
	if event.Status == "FAILED" {
		newStatus = "FAILED"
	}

	_, err := h.db.ExecContext(ctx,
		`UPDATE orders SET status=$1 WHERE id=$2`,
		newStatus, event.OrderID,
	)
	if err != nil {
		return err
	}

	// Update Redis cache
	cached, err := h.rdb.Get(ctx, "order:"+event.OrderID).Bytes()
	if err == nil {
		var order model.CachedOrder
		if json.Unmarshal(cached, &order) == nil {
			order.Status = newStatus
			data, _ := json.Marshal(order)
			h.rdb.Set(ctx, "order:"+event.OrderID, data, 5*time.Minute)
		}
	}

	log.Printf("[handler] order status updated: order_id=%s status=%s", event.OrderID, newStatus)
	return nil
}

// ServeHTTP delegates to the default mux or can be used as a proper handler
func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.CreateOrder(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
