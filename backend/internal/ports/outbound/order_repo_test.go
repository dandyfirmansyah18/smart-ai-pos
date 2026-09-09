package outbound_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestMockOrderRepository(t *testing.T) {
	mockRepo := outbound.NewMockOrderRepository()
	ctx := context.Background()

	orderID := uuid.New()
	order := &domain.Order{
		ID:             orderID,
		TransactionID:  "TXN-12345",
		TotalAmount:    99.98,
		Status:         domain.StatusPending,
		IdempotencyKey: "IDEM-999",
		Items: []domain.OrderItem{
			{
				ID:        uuid.New(),
				OrderID:   orderID,
				ProductID: uuid.New(),
				Quantity:  2,
				UnitPrice: 49.99,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test CreateOrder
	err := mockRepo.CreateOrder(ctx, nil, order)
	if err != nil {
		t.Fatalf("expected no error creating order, got %v", err)
	}

	// Test GetByID
	found, err := mockRepo.GetByID(ctx, orderID.String())
	if err != nil {
		t.Fatalf("expected no error getting order by ID, got %v", err)
	}
	if found.TransactionID != "TXN-12345" {
		t.Errorf("expected TXN-12345, got %s", found.TransactionID)
	}

	// Test GetByIdempotencyKey
	foundIdem, err := mockRepo.GetByIdempotencyKey(ctx, "IDEM-999")
	if err != nil {
		t.Fatalf("expected no error getting order by idempotency key, got %v", err)
	}
	if foundIdem.ID != orderID {
		t.Errorf("expected order ID %s, got %s", orderID, foundIdem.ID)
	}

	// Test UpdateStatus
	err = mockRepo.UpdateStatus(ctx, orderID.String(), domain.StatusCompleted)
	if err != nil {
		t.Fatalf("expected no error updating status, got %v", err)
	}
	if mockRepo.Orders[orderID.String()].Status != domain.StatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", mockRepo.Orders[orderID.String()].Status)
	}
}
