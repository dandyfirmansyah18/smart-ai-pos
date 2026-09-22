package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
)

type dummyFinanceUC struct {
	orders []domain.Order
	pl     map[string]float64
	err    error
}

func (d *dummyFinanceUC) GetOrderHistory(ctx context.Context) ([]domain.Order, error) {
	return d.orders, d.err
}

func (d *dummyFinanceUC) GetProfitLoss(ctx context.Context) (map[string]float64, error) {
	return d.pl, d.err
}

func TestFinanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &dummyFinanceUC{
		orders: []domain.Order{{ID: uuid.New()}},
		pl:     map[string]float64{"total_revenue": 500000},
	}
	h := rest.NewFinanceHandler(uc)

	r := gin.New()
	r.GET("/api/orders/history", h.GetOrderHistory)
	r.GET("/api/finance/profit-loss", h.GetProfitLoss)

	req1, _ := http.NewRequest(http.MethodGet, "/api/orders/history", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w1.Code)
	}

	req2, _ := http.NewRequest(http.MethodGet, "/api/finance/profit-loss", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}
