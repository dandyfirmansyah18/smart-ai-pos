package outbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type FinanceRepository interface {
	ListOrderHistory(ctx context.Context) ([]domain.Order, error)
	CalculateProfitLoss(ctx context.Context) (totalRevenue, totalCOGS, grossProfit, margin float64, err error)
}
