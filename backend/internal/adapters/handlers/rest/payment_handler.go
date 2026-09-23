package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/inbound"
)

type PaymentHandler struct {
	useCase inbound.PaymentUseCase
}

func NewPaymentHandler(useCase inbound.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: useCase}
}

// CreatePaymentCharge POST /api/payments/charge
// @Summary Create payment charge
// @Description Initiate payment charge (Midtrans QRIS / Snap)
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.CreatePaymentChargeRequest true "Payment charge details"
// @Success 201 {object} domain.OrderPayment
// @Router /payments/charge [post]
func (h *PaymentHandler) CreatePaymentCharge(c *gin.Context) {
	var req dto.CreatePaymentChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	method := req.PaymentMethod
	if method == "" {
		method = domain.PaymentMethodMidtrans
	}

	payment, err := h.useCase.CreatePayment(c.Request.Context(), req.OrderID, req.Amount, method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPaymentByOrderID GET /api/payments/order/:order_id
// @Summary Get payment by order ID
// @Description Get payment status and details by order ID
// @Tags Payments
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {object} domain.OrderPayment
// @Router /payments/order/{order_id} [get]
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id parameter is required"})
		return
	}

	payment, err := h.useCase.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment record not found"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// HandleWebhook POST /api/payments/webhook
// @Summary Midtrans payment webhook
// @Description Handle asynchronous payment status notification from Midtrans
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.MidtransWebhookRequest true "Webhook payload"
// @Success 200 {object} object
// @Router /payments/webhook [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	var req dto.MidtransWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	txID := req.OrderID
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required in webhook"})
		return
	}

	err := h.useCase.HandleWebhook(c.Request.Context(), txID, req.TransactionStatus, "received webhook")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
