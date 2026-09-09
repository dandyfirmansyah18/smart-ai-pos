package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestOrderHandler_Checkout_Success(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	mockProductRepo.Products["SKU-ITEM"] = &domain.Product{
		ID:            prodID,
		SKU:           "SKU-ITEM",
		Name:          "Checkout Item",
		Price:         25.00,
		StockQuantity: 10,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	cfg := &config.Config{Port: "8080", Env: "test"}
	server := rest.NewServer(cfg, mockProductRepo, orderUseCase)

	payload := map[string]interface{}{
		"idempotency_key": "IDEM-HEADER-TEST-999",
		"items": []map[string]interface{}{
			{
				"sku":      "SKU-ITEM",
				"quantity": 2,
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/orders/checkout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body: %s", w.Code, w.Body.String())
	}

	var createdOrder domain.Order
	if err := json.Unmarshal(w.Body.Bytes(), &createdOrder); err != nil {
		t.Fatalf("failed to unmarshal created order: %v", err)
	}

	if createdOrder.TotalAmount != 50.00 {
		t.Errorf("expected total amount 50.00, got %.2f", createdOrder.TotalAmount)
	}
}

func TestOrderHandler_Checkout_BadRequest(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	cfg := &config.Config{Port: "8080", Env: "test"}
	server := rest.NewServer(cfg, mockProductRepo, orderUseCase)

	// Missing idempotency key
	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"sku":      "SKU-ITEM",
				"quantity": 0,
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/orders/checkout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing idempotency key, got %d", w.Code)
	}
}
