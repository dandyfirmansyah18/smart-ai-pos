package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestKitchenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := outbound.NewMockOrderRepository()
	hub := ws.NewHub()
	go hub.Run()

	ordID := uuid.New()
	mockRepo.Orders[ordID.String()] = &domain.Order{
		ID:     ordID,
		Status: domain.OrderStatusPending,
	}

	h := rest.NewKitchenHandler(mockRepo, hub)

	r := gin.New()
	r.GET("/api/kitchen/orders", h.ListActiveOrders)
	r.PATCH("/api/kitchen/orders/:id/status", h.UpdateOrderStatus)

	// Test ListActiveOrders
	reqList, _ := http.NewRequest(http.MethodGet, "/api/kitchen/orders", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wList.Code)
	}

	// Test UpdateOrderStatus
	body, _ := json.Marshal(dto.UpdateOrderStatusRequest{Status: domain.OrderStatusPreparing})
	reqUpdate, _ := http.NewRequest(http.MethodPatch, "/api/kitchen/orders/"+ordID.String()+"/status", bytes.NewBuffer(body))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	r.ServeHTTP(wUpdate, reqUpdate)
	if wUpdate.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wUpdate.Code)
	}
}
