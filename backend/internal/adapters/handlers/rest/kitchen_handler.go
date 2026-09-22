package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/outbound"
)

type KitchenHandler struct {
	orderRepo outbound.OrderRepository
	hub       *ws.Hub
}

func NewKitchenHandler(orderRepo outbound.OrderRepository, hub *ws.Hub) *KitchenHandler {
	return &KitchenHandler{
		orderRepo: orderRepo,
		hub:       hub,
	}
}

// ListActiveOrders GET /api/kitchen/orders
func (h *KitchenHandler) ListActiveOrders(c *gin.Context) {
	orders, err := h.orderRepo.ListActiveOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list kitchen orders", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus PATCH /api/kitchen/orders/:id/status
func (h *KitchenHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body format", "details": err.Error()})
		return
	}

	err := h.orderRepo.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status", "details": err.Error()})
		return
	}

	// Broadcast order status update via WebSocket hub if available
	if h.hub != nil {
		h.hub.BroadcastOrderStatusUpdate(id, req.Status)
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully", "status": req.Status})
}
