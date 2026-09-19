package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type CashShiftUseCase interface {
	Open(ctx context.Context, userID string, openingCash float64, notes string) (string, error)
	GetCurrent(ctx context.Context, userID string) (*domain.CashShift, error)
	Close(ctx context.Context, userID string, closingCash float64, notes string) (map[string]interface{}, error)
}
