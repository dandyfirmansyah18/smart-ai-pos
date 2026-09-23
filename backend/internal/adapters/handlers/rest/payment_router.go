package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerPaymentRoutes(api *gin.RouterGroup) {
	if s.paymentUseCase == nil {
		return
	}
	paymentHandler := NewPaymentHandler(s.paymentUseCase)
	api.POST("/payments/charge", paymentHandler.CreatePaymentCharge)
	api.GET("/payments/order/:order_id", paymentHandler.GetPaymentByOrderID)
	api.POST("/payments/webhook", paymentHandler.HandleWebhook)
}
