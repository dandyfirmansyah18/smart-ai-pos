package rest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/ports/inbound"
)

type CashShiftHandler struct {
	useCase inbound.CashShiftUseCase
}

func NewCashShiftHandler(useCase inbound.CashShiftUseCase) *CashShiftHandler {
	return &CashShiftHandler{useCase: useCase}
}

func parseUserID(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	switch v := userIDVal.(type) {
	case string:
		return v, true
	case uuid.UUID:
		return v.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// OpenShift POST /api/cash-shifts/open
func (h *CashShiftHandler) OpenShift(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		OpeningCash float64 `json:"opening_cash" binding:"required,gte=0"`
		Notes       string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	shiftID, err := h.useCase.Open(c.Request.Context(), userID, req.OpeningCash, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "cash shift opened successfully", "shift_id": shiftID})
}

// GetCurrentShift GET /api/cash-shifts/current
func (h *CashShiftHandler) GetCurrentShift(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	shift, err := h.useCase.GetCurrent(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no open cash shift found"})
		return
	}

	c.JSON(http.StatusOK, shift)
}

// CloseShift POST /api/cash-shifts/close
func (h *CashShiftHandler) CloseShift(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ClosingCash float64 `json:"closing_cash" binding:"required,gte=0"`
		Notes       string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	result, err := h.useCase.Close(c.Request.Context(), userID, req.ClosingCash, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
