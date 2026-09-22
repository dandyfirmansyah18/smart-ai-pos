package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerFinanceRoutes(api *gin.RouterGroup) {
	if s.financeUseCase == nil {
		return
	}
	financeHandler := NewFinanceHandler(s.financeUseCase)
	api.GET("/orders/history", financeHandler.GetOrderHistory)
	api.GET("/finance/profit-loss", financeHandler.GetProfitLoss)
}
