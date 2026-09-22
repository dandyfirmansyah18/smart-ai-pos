package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerSyncRoutes(api *gin.RouterGroup) {
	if s.syncUseCase == nil {
		return
	}
	syncHandler := NewSyncHandler(s.syncUseCase)
	api.GET("/sync/status", syncHandler.GetSyncStatus)
	api.POST("/sync/trigger", syncHandler.TriggerSync)
}
