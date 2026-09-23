package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/ports/inbound"
)

type FinanceHandler struct {
	useCase inbound.FinanceUseCase
}

func NewFinanceHandler(useCase inbound.FinanceUseCase) *FinanceHandler {
	return &FinanceHandler{useCase: useCase}
}

// GetOrderHistory GET /api/orders/history
// @Summary Get order history
// @Description Get historical orders for financial reporting
// @Tags Finance
// @Produce json
// @Success 200 {array} domain.Order
// @Router /orders/history [get]
func (h *FinanceHandler) GetOrderHistory(c *gin.Context) {
	orders, err := h.useCase.GetOrderHistory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order history", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetProfitLoss GET /api/finance/profit-loss
// @Summary Get profit & loss summary
// @Description Calculate total revenue, profit, and loss
// @Tags Finance
// @Produce json
// @Success 200 {object} object
// @Router /finance/profit-loss [get]
func (h *FinanceHandler) GetProfitLoss(c *gin.Context) {
	pl, err := h.useCase.GetProfitLoss(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate profit & loss", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pl)
}
