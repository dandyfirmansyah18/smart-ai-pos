package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusUnpaid    OrderStatus = "UNPAID"
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusServed    OrderStatus = "SERVED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusExpired   OrderStatus = "EXPIRED"
	OrderStatusRefunded  OrderStatus = "REFUNDED"

	// Legacy aliases for backward compatibility
	StatusPending   = OrderStatusPending
	StatusCompleted = OrderStatusCompleted
	StatusCancelled = OrderStatusCancelled
	StatusRefunded  = OrderStatusRefunded
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidOrder            = errors.New("invalid order data")
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
)

type Order struct {
	ID             uuid.UUID   `json:"id"`
	TransactionID  string      `json:"transaction_id"`
	TotalAmount    float64     `json:"total_amount"`
	Status         OrderStatus `json:"status"` // PENDING, PREPARING, READY, SERVED, COMPLETED, CANCELLED, REFUNDED
	IdempotencyKey string      `json:"idempotency_key"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}
