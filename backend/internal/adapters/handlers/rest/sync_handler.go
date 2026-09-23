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

// GetSyncStatus GET /api/sync/status
// @Summary Get sync status
// @Description Check offline sync status and pending records count
// @Tags Sync
// @Produce json
// @Success 200 {object} inbound.SyncStatusResponse
// @Router /sync/status [get]
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	status, err := h.syncUseCase.GetSyncStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// TriggerSync POST /api/sync/trigger
// @Summary Trigger synchronization
// @Description Trigger data synchronization from local SQLite to central PostgreSQL
// @Tags Sync
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Router /sync/trigger [post]
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
