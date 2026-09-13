package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
)

type OrderUseCase interface {
	Checkout(ctx context.Context, req dto.CheckoutRequest) (*domain.Order, error)
}

