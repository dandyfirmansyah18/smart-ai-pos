package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type ProductUseCase interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error)
}
