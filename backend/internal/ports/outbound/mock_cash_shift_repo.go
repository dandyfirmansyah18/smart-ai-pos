package outbound

import (
	"context"
	"time"

	"github.com/pos-backend/internal/domain"
)

type MockCashShiftRepository struct {
	Shifts map[string]*domain.CashShift
}

func NewMockCashShiftRepository() *MockCashShiftRepository {
	return &MockCashShiftRepository{
		Shifts: make(map[string]*domain.CashShift),
	}
}

func (m *MockCashShiftRepository) OpenShift(ctx context.Context, shift *domain.CashShift) error {
	m.Shifts[shift.UserID.String()] = shift
	return nil
}

func (m *MockCashShiftRepository) GetCurrentShift(ctx context.Context, userID string) (*domain.CashShift, error) {
	s, ok := m.Shifts[userID]
	if !ok || s.Status == "CLOSED" {
		return nil, domain.ErrOrderNotFound
	}
	return s, nil
}

func (m *MockCashShiftRepository) CloseShift(ctx context.Context, shiftID string, closingCash float64, expectedCash float64, totalSales float64, notes string) error {
	for _, s := range m.Shifts {
		if s.ID.String() == shiftID {
			s.Status = "CLOSED"
			s.ClosingCash = closingCash
			s.ExpectedCash = expectedCash
			s.TotalCashSales = totalSales
			s.Notes = notes
			now := time.Now()
			s.ClosedAt = &now
			return nil
		}
	}
	return domain.ErrOrderNotFound
}

func (m *MockCashShiftRepository) CalculateTotalSales(ctx context.Context, openedAt time.Time) (float64, error) {
	return 150000.0, nil
}

var _ CashShiftRepository = (*MockCashShiftRepository)(nil)
