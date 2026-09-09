package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Order status constants
const (
	StatusPending   = "PENDING"
	StatusCompleted = "COMPLETED"
	StatusCancelled = "CANCELLED"
	StatusRefunded  = "REFUNDED"
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
	Status         string      `json:"status"` // PENDING, COMPLETED, CANCELLED, REFUNDED
	IdempotencyKey string      `json:"idempotency_key"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}
