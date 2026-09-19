package outbound

import (
	"context"
	"time"

	"github.com/pos-backend/internal/domain"
)

type CashShiftRepository interface {
	OpenShift(ctx context.Context, shift *domain.CashShift) error
	GetCurrentShift(ctx context.Context, userID string) (*domain.CashShift, error)
	CloseShift(ctx context.Context, shiftID string, closingCash float64, expectedCash float64, totalSales float64, notes string) error
	CalculateTotalSales(ctx context.Context, openedAt time.Time) (float64, error)
}
