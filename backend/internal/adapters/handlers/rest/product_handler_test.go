package rest_test

import (
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

func TestProductHandler_ListProducts(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	mockProductRepo.Products["SKU-001"] = &domain.Product{
		ID:            prodID,
		SKU:           "SKU-001",
		Name:          "Test Product",
		Price:         10.50,
		StockQuantity: 50,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)

	cfg := &config.Config{Port: "8080", Env: "test"}
	server := rest.NewServer(cfg, mockProductRepo, orderUseCase)

	req, _ := http.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	var products []domain.Product
	if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if len(products) != 1 || products[0].SKU != "SKU-001" {
		t.Errorf("unexpected products output: %+v", products)
	}
}

func TestProductHandler_GetProductBySKU(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	prodID := uuid.New()
	mockProductRepo.Products["SKU-LATTE"] = &domain.Product{
		ID:            prodID,
		SKU:           "SKU-LATTE",
		Name:          "Iced Latte",
		Price:         4.50,
		StockQuantity: 30,
	}

	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	cfg := &config.Config{Port: "8080", Env: "test"}
	server := rest.NewServer(cfg, mockProductRepo, orderUseCase)

	// Test found
	req, _ := http.NewRequest(http.MethodGet, "/api/products/SKU-LATTE", nil)
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// Test not found
	reqNotFound, _ := http.NewRequest(http.MethodGet, "/api/products/NON-EXISTENT", nil)
	wNotFound := httptest.NewRecorder()
	server.Router().ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", wNotFound.Code)
	}
}
