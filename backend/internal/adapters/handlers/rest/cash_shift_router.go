package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
)

func (s *Server) registerCashShiftRoutes(api *gin.RouterGroup) {
	if s.cashShiftUseCase == nil {
		return
	}
	cashHandler := NewCashShiftHandler(s.cashShiftUseCase)
	cashGroup := api.Group("/cash-shifts")
	if s.authUseCase != nil {
		cashGroup.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
	}
	{
		cashGroup.POST("/open", cashHandler.OpenShift)
		cashGroup.GET("/current", cashHandler.GetCurrentShift)
		cashGroup.POST("/close", cashHandler.CloseShift)
	}
}
