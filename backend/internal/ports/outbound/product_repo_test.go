package outbound_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestMockProductRepository(t *testing.T) {
	mockRepo := outbound.NewMockProductRepository()
	ctx := context.Background()

	prodID := uuid.New()
	prod := &domain.Product{
		ID:            prodID,
		SKU:           "SKU-100",
		Name:          "Test Product",
		Description:   "Test Description",
		Price:         49.99,
		StockQuantity: 10,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockRepo.Products[prod.SKU] = prod

	// Test GetBySKU
	found, err := mockRepo.GetBySKU(ctx, "SKU-100")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.ID != prodID {
		t.Errorf("expected product ID %s, got %s", prodID, found.ID)
	}

	// Test GetBySKU Not Found
	_, err = mockRepo.GetBySKU(ctx, "NON-EXISTENT")
	if err != domain.ErrProductNotFound {
		t.Errorf("expected ErrProductNotFound, got %v", err)
	}

	// Test UpdateStock
	err = mockRepo.UpdateStock(ctx, nil, "SKU-100", 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mockRepo.Products["SKU-100"].StockQuantity != 5 {
		t.Errorf("expected stock 5, got %d", mockRepo.Products["SKU-100"].StockQuantity)
	}

	// Test ListAll
	all, err := mockRepo.ListAll(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 product, got %d", len(all))
	}
}
