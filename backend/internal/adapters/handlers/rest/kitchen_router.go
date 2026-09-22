package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
)

func (s *Server) registerKitchenRoutes(api *gin.RouterGroup) {
	if s.orderRepo == nil {
		return
	}
	kitchenHandler := NewKitchenHandler(s.orderRepo, s.hub)
	kitchenGroup := api.Group("/kitchen")
	if s.authUseCase != nil {
		kitchenGroup.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
		kitchenGroup.Use(middleware.RequireRole("ADMIN", "KITCHEN", "CASHIER"))
	}
	{
		kitchenGroup.GET("/orders", kitchenHandler.ListActiveOrders)
		kitchenGroup.PATCH("/orders/:id/status", kitchenHandler.UpdateOrderStatus)
	}
}
