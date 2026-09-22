package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerOrderRoutes(api *gin.RouterGroup) {
	orderHandler := NewOrderHandler(s.orderUseCase)
	api.POST("/orders/checkout", orderHandler.Checkout)
}
