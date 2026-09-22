package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/ports/inbound"
)

type SyncHandler struct {
	syncUseCase inbound.SyncUseCase
}

func NewSyncHandler(syncUseCase inbound.SyncUseCase) *SyncHandler {
	return &SyncHandler{syncUseCase: syncUseCase}
}

func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	status, err := h.syncUseCase.GetSyncStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *SyncHandler) TriggerSync(c *gin.Context) {
	var req struct {
		PostgresDSN string `json:"postgres_dsn"`
	}
	_ = c.ShouldBindJSON(&req)

	err := h.syncUseCase.TriggerSync(c.Request.Context(), req.PostgresDSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "synchronization completed successfully"})
}
