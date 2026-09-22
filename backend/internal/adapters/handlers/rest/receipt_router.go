package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerReceiptRoutes(api *gin.RouterGroup) {
	if s.receiptUseCase == nil {
		return
	}
	receiptHandler := NewReceiptHandler(s.receiptUseCase)
	api.POST("/receipts/scan", receiptHandler.ScanReceipt)
	api.GET("/receipts/audits", receiptHandler.ListAudits)
}
