package inbound_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestCheckout_Success(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	prod := &domain.Product{
		ID:            prodID,
		SKU:           "SKU-COFFEE",
		Name:          "Coffee Espresso",
		Description:   "Fresh Coffee",
		Price:         5.00,
		StockQuantity: 10,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	mockProductRepo.Products[prod.SKU] = prod

	useCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	ctx := context.Background()

	req := inbound.CheckoutRequest{
		IdempotencyKey: "IDEM-TEST-100",
		Items: []inbound.CheckoutItem{
			{
				SKU:      "SKU-COFFEE",
				Quantity: 2,
			},
		},
	}

	order, err := useCase.Checkout(ctx, req)
	if err != nil {
		t.Fatalf("expected no error during checkout, got %v", err)
	}

	if order.TotalAmount != 10.00 {
		t.Errorf("expected total amount 10.00, got %.2f", order.TotalAmount)
	}
	if order.Status != domain.StatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", order.Status)
	}
	if len(order.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(order.Items))
	}

	// Verify product stock was decremented to 8
	if mockProductRepo.Products["SKU-COFFEE"].StockQuantity != 8 {
		t.Errorf("expected stock 8, got %d", mockProductRepo.Products["SKU-COFFEE"].StockQuantity)
	}
}

func TestCheckout_InsufficientStock(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	prod := &domain.Product{
		ID:            prodID,
		SKU:           "SKU-LIMITED",
		Name:          "Limited Item",
		Price:         20.00,
		StockQuantity: 2,
	}
	mockProductRepo.Products[prod.SKU] = prod

	useCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	ctx := context.Background()

	req := inbound.CheckoutRequest{
		IdempotencyKey: "IDEM-TEST-101",
		Items: []inbound.CheckoutItem{
			{
				SKU:      "SKU-LIMITED",
				Quantity: 5, // Requesting more than available stock (2)
			},
		},
	}

	_, err := useCase.Checkout(ctx, req)
	if err == nil {
		t.Fatalf("expected error due to insufficient stock, got nil")
	}
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	// Stock must remain unchanged at 2
	if mockProductRepo.Products["SKU-LIMITED"].StockQuantity != 2 {
		t.Errorf("expected stock to remain 2, got %d", mockProductRepo.Products["SKU-LIMITED"].StockQuantity)
	}
}

func TestCheckout_IdempotencyRetry(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	prod := &domain.Product{
		ID:            prodID,
		SKU:           "SKU-IDEM",
		Name:          "Idempotency Test Item",
		Price:         15.00,
		StockQuantity: 10,
	}
	mockProductRepo.Products[prod.SKU] = prod

	useCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	ctx := context.Background()

	req := inbound.CheckoutRequest{
		IdempotencyKey: "IDEM-SAME-KEY-555",
		Items: []inbound.CheckoutItem{
			{
				SKU:      "SKU-IDEM",
				Quantity: 1,
			},
		},
	}

	// First checkout
	order1, err := useCase.Checkout(ctx, req)
	if err != nil {
		t.Fatalf("expected first checkout to succeed, got %v", err)
	}

	// Stock decremented to 9
	if mockProductRepo.Products["SKU-IDEM"].StockQuantity != 9 {
		t.Errorf("expected stock 9 after first checkout, got %d", mockProductRepo.Products["SKU-IDEM"].StockQuantity)
	}

	// Retry checkout with identical idempotency key
	order2, err := useCase.Checkout(ctx, req)
	if err != nil {
		t.Fatalf("expected retry checkout to return existing order without error, got %v", err)
	}

	// Assert return details match original order
	if order1.ID != order2.ID {
		t.Errorf("expected same order ID %s, got %s", order1.ID, order2.ID)
	}

	// Stock should still be 9 (not decremented twice)
	if mockProductRepo.Products["SKU-IDEM"].StockQuantity != 9 {
		t.Errorf("expected stock to remain 9 after retry, got %d", mockProductRepo.Products["SKU-IDEM"].StockQuantity)
	}
}

func TestCheckout_ConcurrentLockBlocked(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	useCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	ctx := context.Background()

	// Simulate Client A holding lock on key
	mockLockService.AcquireLock(ctx, "lock:checkout:LOCKED-KEY", 10*time.Second)

	req := inbound.CheckoutRequest{
		IdempotencyKey: "LOCKED-KEY",
		Items: []inbound.CheckoutItem{
			{
				SKU:      "SKU-TEST",
				Quantity: 1,
			},
		},
	}

	_, err := useCase.Checkout(ctx, req)
	if err == nil {
		t.Fatalf("expected error when lock is held by another process")
	}
	if !errors.Is(err, domain.ErrDuplicateIdempotencyKey) {
		t.Errorf("expected ErrDuplicateIdempotencyKey, got %v", err)
	}
}
