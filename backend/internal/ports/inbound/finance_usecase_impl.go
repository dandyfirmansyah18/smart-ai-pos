package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type FinanceUseCaseImpl struct {
	repo outbound.FinanceRepository
}

func NewFinanceUseCaseImpl(repo outbound.FinanceRepository) *FinanceUseCaseImpl {
	return &FinanceUseCaseImpl{repo: repo}
}

func (u *FinanceUseCaseImpl) GetOrderHistory(ctx context.Context) ([]domain.Order, error) {
	return u.repo.ListOrderHistory(ctx)
}

func (u *FinanceUseCaseImpl) GetProfitLoss(ctx context.Context) (map[string]float64, error) {
	rev, cogs, profit, margin, err := u.repo.CalculateProfitLoss(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]float64{
		"total_revenue": rev,
		"total_cogs":    cogs,
		"gross_profit":  profit,
		"profit_margin": margin,
	}, nil
}

var _ FinanceUseCase = (*FinanceUseCaseImpl)(nil)
