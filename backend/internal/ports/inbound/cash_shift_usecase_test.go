package inbound_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestCashShiftUseCase_OpenAndClose(t *testing.T) {
	mockRepo := outbound.NewMockCashShiftRepository()
	uc := inbound.NewCashShiftUseCaseImpl(mockRepo)
	ctx := context.Background()

	userID := uuid.New().String()

	// 1. Open Shift
	shiftID, err := uc.Open(ctx, userID, 500000, "Starting modal")
	if err != nil {
		t.Fatalf("expected no error opening shift, got %v", err)
	}
	if shiftID == "" {
		t.Errorf("expected shift ID")
	}

	// 2. Get Current Shift
	curr, err := uc.GetCurrent(ctx, userID)
	if err != nil {
		t.Fatalf("expected current shift, got %v", err)
	}
	if curr.OpeningCash != 500000 {
		t.Errorf("expected opening cash 500000, got %.2f", curr.OpeningCash)
	}

	// 3. Close Shift
	res, err := uc.Close(ctx, userID, 650000, "Closing notes")
	if err != nil {
		t.Fatalf("expected no error closing shift, got %v", err)
	}
	if res["expected_cash"] != 650000.0 {
		t.Errorf("expected expected cash 650000, got %v", res["expected_cash"])
	}
}
