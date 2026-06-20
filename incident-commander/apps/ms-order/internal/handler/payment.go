package handler

import (
	"context"
	"encoding/json"
	"log"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/model"
)

type PaymentHandler struct {
	orderHandler *OrderHandler
}

func NewPaymentHandler(orderHandler *OrderHandler) *PaymentHandler {
	return &PaymentHandler{orderHandler: orderHandler}
}

func (h *PaymentHandler) HandlePaymentCompleted(ctx context.Context, msg []byte) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "order.payment_completed")
	defer span.Finish()

	var event model.PaymentCompletedEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	log.Printf("[handler] received payment-completed: order=%s status=%s",
		event.OrderID, event.Status)

	return h.orderHandler.UpdateOrderStatus(ctx, event)
}
