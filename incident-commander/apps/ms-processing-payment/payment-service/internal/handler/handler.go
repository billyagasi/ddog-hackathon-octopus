package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/model"
)

type Handler struct {
	cfg    config.Config
	db     *sql.DB
	es     *elasticsearch.Client
	writer *kafka.Writer
}

func New(cfg config.Config, db *sql.DB, es *elasticsearch.Client, writer *kafka.Writer) *Handler {
	return &Handler{
		cfg:    cfg,
		db:     db,
		es:     es,
		writer: writer,
	}
}

func (h *Handler) ProcessPayment(ctx context.Context, event model.InventoryReservedEvent) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "payment.process")
	defer span.Finish()

	span.SetTag("order_id", event.OrderID)
	span.SetTag("product_id", event.ProductID)

	log.Printf("[handler] processing payment for order=%s amount=%.2f", event.OrderID, event.Amount)

	gatewayResp, err := h.callPaymentGateway(ctx, event)
	if err != nil {
		return fmt.Errorf("gateway call failed: %w", err)
	}

	payment, err := h.savePayment(ctx, event, gatewayResp)
	if err != nil {
		return fmt.Errorf("save payment failed: %w", err)
	}

	if err := h.indexToElasticsearch(ctx, event, payment); err != nil {
		log.Printf("[handler] WARN elasticsearch index failed (non-fatal): %v", err)
	}

	if err := h.publishEvent(ctx, payment); err != nil {
		return fmt.Errorf("publish event failed: %w", err)
	}

	log.Printf("[handler] payment completed: payment_id=%s status=%s", payment.ID, payment.Status)
	return nil
}

func (h *Handler) callPaymentGateway(ctx context.Context, event model.InventoryReservedEvent) (model.PaymentGatewayResponse, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "http.client")
	defer span.Finish()

	span.SetTag("http.url", h.cfg.PaymentGatewayURL+"/api/payment")

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":   event.OrderID,
		"product_id": event.ProductID,
		"quantity":   event.Quantity,
		"amount":     event.Amount,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.cfg.PaymentGatewayURL+"/api/payment",
		bytes.NewReader(payload))
	if err != nil {
		return model.PaymentGatewayResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.PaymentGatewayResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.PaymentGatewayResponse{}, err
	}

	var gatewayResp model.PaymentGatewayResponse
	if err := json.Unmarshal(body, &gatewayResp); err != nil {
		return model.PaymentGatewayResponse{}, err
	}

	span.SetTag("http.status_code", resp.StatusCode)
	span.SetTag("payment.gateway_status", gatewayResp.Status)
	span.SetTag("payment.gateway_txn_id", gatewayResp.TransactionID)

	return gatewayResp, nil
}

func (h *Handler) savePayment(ctx context.Context, event model.InventoryReservedEvent, gwResp model.PaymentGatewayResponse) (model.Payment, error) {
	span, ctx := tracer.StartSpanFromContext(ctx, "postgres.query")
	defer span.Finish()

	payment := model.Payment{
		ID:        uuid.New().String(),
		OrderID:   event.OrderID,
		Amount:    event.Amount,
		Status:    gwResp.Status,
		GatewayTxnID: gwResp.TransactionID,
		CreatedAt: time.Now(),
	}

	span.SetTag("payment.id", payment.ID)
	span.SetTag("db.type", "postgres")

	_, err := h.db.ExecContext(ctx,
		`INSERT INTO payments (id, order_id, amount, status, gateway_txn_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		payment.ID, payment.OrderID, payment.Amount, payment.Status,
		payment.GatewayTxnID, payment.CreatedAt,
	)
	if err != nil {
		return payment, err
	}

	return payment, nil
}

func (h *Handler) indexToElasticsearch(ctx context.Context, event model.InventoryReservedEvent, payment model.Payment) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "elasticsearch.index")
	defer span.Finish()

	doc := model.ESPaymentDocument{
		PaymentID:    payment.ID,
		OrderID:      payment.OrderID,
		ProductID:    event.ProductID,
		Quantity:     event.Quantity,
		Amount:       payment.Amount,
		Status:       payment.Status,
		GatewayTxnID: payment.GatewayTxnID,
		Timestamp:    payment.CreatedAt.Format(time.RFC3339),
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	span.SetTag("elasticsearch.index", "payment-transactions")
	span.SetTag("payment.id", payment.ID)

	res, err := h.es.Index(
		"payment-transactions",
		strings.NewReader(string(body)),
		h.es.Index.WithContext(ctx),
		h.es.Index.WithDocumentID(payment.ID),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch index error: %s", string(respBody))
	}

	return nil
}

func (h *Handler) publishEvent(ctx context.Context, payment model.Payment) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "kafka.produce")
	defer span.Finish()

	event := model.PaymentCompletedEvent{
		OrderID:   payment.OrderID,
		PaymentID: payment.ID,
		Amount:    payment.Amount,
		Status:    payment.Status,
		GatewayTxnID: payment.GatewayTxnID,
		Timestamp: payment.CreatedAt.Format(time.RFC3339),
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	span.SetTag("kafka.topic", "payment-completed")
	span.SetTag("kafka.key", payment.OrderID)
	span.SetTag("payment.status", payment.Status)

	err = h.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(payment.OrderID),
		Value: msg,
	})
	if err != nil {
		return err
	}

	return nil
}
