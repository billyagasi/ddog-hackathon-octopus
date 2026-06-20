package model

import "time"

type InventoryReservedEvent struct {
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type PaymentGatewayResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type Payment struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	GatewayTxnID string `json:"gateway_txn_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentCompletedEvent struct {
	OrderID   string  `json:"order_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	GatewayTxnID string `json:"gateway_txn_id"`
	Timestamp string  `json:"timestamp"`
}

type ESPaymentDocument struct {
	PaymentID   string  `json:"payment_id"`
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	GatewayTxnID string `json:"gateway_txn_id"`
	Timestamp   string  `json:"@timestamp"`
}
