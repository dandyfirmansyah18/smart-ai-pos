package inbound

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
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
		Status:      domain.CashShiftStatusOpen,
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

func (u *CashShiftUseCaseImpl) Close(ctx context.Context, userID string, closingCash float64, notes string) (*dto.CloseCashShiftResponse, error) {
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

	return &dto.CloseCashShiftResponse{
		Message:      "cash shift closed successfully",
		OpeningCash:  shift.OpeningCash,
		TotalSales:   totalSales,
		ExpectedCash: expectedCash,
		ClosingCash:  closingCash,
		Difference:   closingCash - expectedCash,
	}, nil
}

var _ CashShiftUseCase = (*CashShiftUseCaseImpl)(nil)
