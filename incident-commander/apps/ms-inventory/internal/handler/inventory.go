package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/model"
)

type InventoryHandler struct {
	cfg    config.Config
	db     *sql.DB
	rdb    *redis.Client
	writer *kafka.Writer
}

func NewInventoryHandler(cfg config.Config, db *sql.DB, rdb *redis.Client, writer *kafka.Writer) *InventoryHandler {
	return &InventoryHandler{cfg: cfg, db: db, rdb: rdb, writer: writer}
}

func (h *InventoryHandler) ProcessOrder(ctx context.Context, event model.OrderCreatedEvent) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "inventory.reserve")
	defer span.Finish()

	span.SetTag("order_id", event.OrderID)
	span.SetTag("product_id", event.ProductID)

	log.Printf("[handler] processing reservation: order=%s product=%s qty=%d",
		event.OrderID, event.ProductID, event.Quantity)

	// 1. Check Redis cache
	stock, found := h.checkCache(ctx, event.ProductID)
	if found {
		span.SetTag("redis.hit", true)
		log.Printf("[handler] redis cache hit: product=%s stock=%d", event.ProductID, stock)
	} else {
		span.SetTag("redis.hit", false)
	}

	// 2. Check PostgreSQL inventory
	inv, err := h.checkInventory(ctx, event.ProductID)
	if err != nil {
		return err
	}

	if inv.Stock < event.Quantity {
		log.Printf("[handler] insufficient stock: product=%s stock=%d needed=%d",
			event.ProductID, inv.Stock, event.Quantity)

		return h.publishResponse(ctx, event, "FAILED", 0)
	}

	// 3. Reserve (decrement stock)
	if err := h.reserveStock(ctx, event.ProductID, event.Quantity); err != nil {
		return err
	}

	// 4. Update Redis cache
	h.updateCache(ctx, event.ProductID, inv.Stock-event.Quantity)

	// 5. Publish inventory-reserved
	amount := float64(event.Quantity) * (10000 + float64(rand.Intn(50000)))
	return h.publishResponse(ctx, event, "RESERVED", amount)
}

func (h *InventoryHandler) checkCache(ctx context.Context, productID string) (int, bool) {
	span, _ := tracer.StartSpanFromContext(ctx, "redis.cache")
	defer span.Finish()

	span.SetTag("redis.key", "inventory:"+productID)

	data, err := h.rdb.Get(ctx, "inventory:"+productID).Bytes()
	if err != nil {
		return 0, false
	}

	var cached model.CachedInventory
	if err := json.Unmarshal(data, &cached); err != nil {
		return 0, false
	}

	return cached.Stock, true
}

func (h *InventoryHandler) checkInventory(ctx context.Context, productID string) (model.Inventory, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "postgres.query")
	defer span.Finish()

	span.SetTag("db.type", "postgres")
	span.SetTag("product_id", productID)

	var inv model.Inventory
	err := h.db.QueryRowContext(ctx,
		`SELECT product_id, stock, updated_at FROM inventory WHERE product_id=$1`,
		productID,
	).Scan(&inv.ProductID, &inv.Stock, &inv.UpdatedAt)
	if err != nil {
		return inv, err
	}

	return inv, nil
}

func (h *InventoryHandler) reserveStock(ctx context.Context, productID string, qty int) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "postgres.update")
	defer span.Finish()

	_, err := h.db.ExecContext(ctx,
		`UPDATE inventory SET stock=stock-$1, updated_at=NOW() WHERE product_id=$2 AND stock>=$1`,
		qty, productID,
	)
	return err
}

func (h *InventoryHandler) updateCache(ctx context.Context, productID string, newStock int) {
	span, ctx := tracer.StartSpanFromContext(ctx, "redis.cache")
	defer span.Finish()

	cached := model.CachedInventory{
		ProductID: productID,
		Stock:     newStock,
	}
	data, _ := json.Marshal(cached)
	h.rdb.Set(ctx, "inventory:"+productID, data, 1*time.Minute)
}

func (h *InventoryHandler) publishResponse(ctx context.Context, event model.OrderCreatedEvent, status string, amount float64) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.produce")
	defer span.Finish()

	resp := model.InventoryReservedEvent{
		OrderID:   event.OrderID,
		ProductID: event.ProductID,
		Quantity:  event.Quantity,
		Amount:    amount,
		Status:    status,
	}

	msg, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	span.SetTag("kafka.topic", "inventory-reserved")
	span.SetTag("kafka.key", event.OrderID)
	span.SetTag("inventory.status", status)

	err = h.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.OrderID),
		Value: msg,
	})
	if err != nil {
		return err
	}

	log.Printf("[handler] inventory-reserved published: order=%s status=%s", event.OrderID, status)
	return nil
}
