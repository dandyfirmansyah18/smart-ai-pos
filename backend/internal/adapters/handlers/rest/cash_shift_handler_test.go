package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
)

type dummyCashShiftUC struct {
	shiftID  string
	shift    *domain.CashShift
	closeRes *dto.CloseCashShiftResponse
	err      error
}

func (d *dummyCashShiftUC) Open(ctx context.Context, cashierID string, openingCash float64, notes string) (string, error) {
	return d.shiftID, d.err
}

func (d *dummyCashShiftUC) GetCurrent(ctx context.Context, cashierID string) (*domain.CashShift, error) {
	return d.shift, d.err
}

func (d *dummyCashShiftUC) Close(ctx context.Context, cashierID string, closingCash float64, notes string) (*dto.CloseCashShiftResponse, error) {
	return d.closeRes, d.err
}

func TestCashShiftHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &dummyCashShiftUC{
		shiftID: uuid.New().String(),
		shift: &domain.CashShift{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			OpeningCash: 100000,
			Status:      domain.CashShiftStatusOpen,
			OpenedAt:    time.Now(),
		},
		closeRes: &dto.CloseCashShiftResponse{
			Message:     "closed",
			ClosingCash: 150000,
		},
	}
	h := rest.NewCashShiftHandler(uc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Next()
	})

	r.POST("/api/cash-shifts/open", h.OpenShift)
	r.GET("/api/cash-shifts/current", h.GetCurrentShift)
	r.POST("/api/cash-shifts/close", h.CloseShift)

	body, _ := json.Marshal(dto.OpenCashShiftRequest{OpeningCash: 100000})
	req, _ := http.NewRequest(http.MethodPost, "/api/cash-shifts/open", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", w.Code)
	}

	reqGet, _ := http.NewRequest(http.MethodGet, "/api/cash-shifts/current", nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", wGet.Code)
	}

	bodyClose, _ := json.Marshal(dto.CloseCashShiftRequest{ClosingCash: 150000})
	reqClose, _ := http.NewRequest(http.MethodPost, "/api/cash-shifts/close", bytes.NewBuffer(bodyClose))
	reqClose.Header.Set("Content-Type", "application/json")
	wClose := httptest.NewRecorder()
	r.ServeHTTP(wClose, reqClose)
	if wClose.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", wClose.Code)
	}
}
