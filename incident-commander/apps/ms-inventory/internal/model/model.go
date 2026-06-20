package model

import "time"

// OrderCreatedEvent consumed from Kafka topic order-created
type OrderCreatedEvent struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Inventory is the database model
type Inventory struct {
	ProductID string    `json:"product_id"`
	Stock     int       `json:"stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InventoryReservedEvent published to Kafka topic inventory-reserved
type InventoryReservedEvent struct {
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// CachedInventory stored in Redis
type CachedInventory struct {
	ProductID string `json:"product_id"`
	Stock     int    `json:"stock"`
}
