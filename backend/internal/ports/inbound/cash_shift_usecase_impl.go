package inbound

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type CashShiftUseCaseImpl struct {
	repo outbound.CashShiftRepository
}

func NewCashShiftUseCaseImpl(repo outbound.CashShiftRepository) *CashShiftUseCaseImpl {
	return &CashShiftUseCaseImpl{repo: repo}
}

func (u *CashShiftUseCaseImpl) Open(ctx context.Context, userID string, openingCash float64, notes string) (string, error) {
	parsedUID, err := uuid.Parse(userID)
	if err != nil {
		return "", errors.New("invalid user ID format")
	}

	// Check if already open
	existing, _ := u.repo.GetCurrentShift(ctx, userID)
	if existing != nil {
		return "", errors.New("you already have an open cash shift")
	}

	shiftID := uuid.New()
	shift := &domain.CashShift{
		ID:          shiftID,
		UserID:      parsedUID,
		OpeningCash: openingCash,
		Notes:       notes,
	}

	if err := u.repo.OpenShift(ctx, shift); err != nil {
		return "", err
	}

	return shiftID.String(), nil
}

func (u *CashShiftUseCaseImpl) GetCurrent(ctx context.Context, userID string) (*domain.CashShift, error) {
	return u.repo.GetCurrentShift(ctx, userID)
}

func (u *CashShiftUseCaseImpl) Close(ctx context.Context, userID string, closingCash float64, notes string) (map[string]interface{}, error) {
	shift, err := u.repo.GetCurrentShift(ctx, userID)
	if err != nil {
		return nil, errors.New("no open cash shift found for user")
	}

	totalSales, err := u.repo.CalculateTotalSales(ctx, shift.OpenedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sales: %w", err)
	}

	expectedCash := shift.OpeningCash + totalSales

	if err := u.repo.CloseShift(ctx, shift.ID.String(), closingCash, expectedCash, totalSales, notes); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":       "cash shift closed successfully",
		"opening_cash":  shift.OpeningCash,
		"total_sales":   totalSales,
		"expected_cash": expectedCash,
		"closing_cash":  closingCash,
		"difference":    closingCash - expectedCash,
	}, nil
}

var _ CashShiftUseCase = (*CashShiftUseCaseImpl)(nil)
