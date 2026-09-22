package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/ports/inbound"
)

type dummySyncUC struct {
	status *inbound.SyncStatusResponse
	err    error
}

func (d *dummySyncUC) GetSyncStatus(ctx context.Context) (*inbound.SyncStatusResponse, error) {
	return d.status, d.err
}

func (d *dummySyncUC) TriggerSync(ctx context.Context, dsn string) error {
	return d.err
}

func TestSyncHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &dummySyncUC{
		status: &inbound.SyncStatusResponse{IsOnline: true, PendingOrders: 0},
	}
	h := rest.NewSyncHandler(uc)

	r := gin.New()
	r.GET("/api/sync/status", h.GetSyncStatus)
	r.POST("/api/sync/trigger", h.TriggerSync)

	req1, _ := http.NewRequest(http.MethodGet, "/api/sync/status", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w1.Code)
	}

	req2, _ := http.NewRequest(http.MethodPost, "/api/sync/trigger", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}
