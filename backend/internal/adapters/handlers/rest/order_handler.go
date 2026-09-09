package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/inbound"
)

type OrderHandler struct {
	useCase inbound.OrderUseCase
}

func NewOrderHandler(useCase inbound.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: useCase}
}

// Checkout POST /api/orders/checkout
func (h *OrderHandler) Checkout(c *gin.Context) {
	var req inbound.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body format", "details": err.Error()})
		return
	}

	// Support idempotency key passed via custom HTTP header X-Idempotency-Key
	headerIdempotencyKey := c.GetHeader("X-Idempotency-Key")
	if req.IdempotencyKey == "" && headerIdempotencyKey != "" {
		req.IdempotencyKey = headerIdempotencyKey
	}

	if req.IdempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Idempotency-Key header or idempotency_key in body is required"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order items list cannot be empty"})
		return
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item quantity must be greater than zero"})
			return
		}
	}

	order, err := h.useCase.Checkout(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrInsufficientStock) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrDuplicateIdempotencyKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "checkout request with this idempotency key is currently being processed"})
			return
		}
		if errors.Is(err, domain.ErrInvalidOrder) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process checkout", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}
