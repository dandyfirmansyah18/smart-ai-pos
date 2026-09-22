package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
)

type dummyPaymentUC struct {
	payment *domain.OrderPayment
	err     error
}

func (d *dummyPaymentUC) CreatePayment(ctx context.Context, orderID string, amount float64, method domain.PaymentMethod) (*domain.OrderPayment, error) {
	return d.payment, d.err
}

func (d *dummyPaymentUC) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error) {
	return d.payment, d.err
}

func (d *dummyPaymentUC) HandleWebhook(ctx context.Context, transactionID, status, message string) error {
	return d.err
}

func TestPaymentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payID := uuid.New()
	ordUUID := uuid.New()
	ordIDStr := ordUUID.String()
	uc := &dummyPaymentUC{
		payment: &domain.OrderPayment{
			ID:      payID,
			OrderID: ordUUID,
			Amount:  50000,
			Status:  domain.PaymentStatusPending,
		},
	}
	h := rest.NewPaymentHandler(uc)

	r := gin.New()
	r.POST("/api/payments/charge", h.CreatePaymentCharge)
	r.GET("/api/payments/order/:order_id", h.GetPaymentByOrderID)
	r.POST("/api/payments/webhook", h.HandleWebhook)

	// Test Charge
	body, _ := json.Marshal(dto.CreatePaymentChargeRequest{OrderID: ordIDStr, Amount: 50000})
	req, _ := http.NewRequest(http.MethodPost, "/api/payments/charge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// Test GetPaymentByOrderID
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/payments/order/"+ordIDStr, nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wGet.Code)
	}

	// Test Webhook
	whBody, _ := json.Marshal(dto.MidtransWebhookRequest{OrderID: ordIDStr, TransactionStatus: "settlement"})
	reqWH, _ := http.NewRequest(http.MethodPost, "/api/payments/webhook", bytes.NewBuffer(whBody))
	reqWH.Header.Set("Content-Type", "application/json")
	wWH := httptest.NewRecorder()
	r.ServeHTTP(wWH, reqWH)
	if wWH.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wWH.Code)
	}
}
