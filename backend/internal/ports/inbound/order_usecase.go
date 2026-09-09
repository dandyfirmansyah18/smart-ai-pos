package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type CheckoutItem struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

type CheckoutRequest struct {
	IdempotencyKey string         `json:"idempotency_key" binding:"required"`
	Items          []CheckoutItem `json:"items" binding:"required,dive"`
}

type OrderUseCase interface {
	Checkout(ctx context.Context, req CheckoutRequest) (*domain.Order, error)
}
