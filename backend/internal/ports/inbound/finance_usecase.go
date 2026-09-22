package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type FinanceUseCase interface {
	GetOrderHistory(ctx context.Context) ([]domain.Order, error)
	GetProfitLoss(ctx context.Context) (map[string]float64, error)
}
