package rest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/dto"
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
// @Summary Open cash shift
// @Description Open a new cashier shift (buka kasir)
// @Tags CashShifts
// @Accept json
// @Produce json
// @Param request body dto.OpenCashShiftRequest true "Opening cash amount and notes"
// @Success 201 {object} dto.OpenCashShiftResponse
// @Router /cash-shifts/open [post]
func (h *CashShiftHandler) OpenShift(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.OpenCashShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	shiftID, err := h.useCase.Open(c.Request.Context(), userID, req.OpeningCash, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.OpenCashShiftResponse{
		Message: "cash shift opened successfully",
		ShiftID: shiftID,
	})
}

// GetCurrentShift GET /api/cash-shifts/current
// @Summary Get current cash shift
// @Description Get active open cash shift for logged-in cashier
// @Tags CashShifts
// @Produce json
// @Success 200 {object} domain.CashShift
// @Router /cash-shifts/current [get]
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
// @Summary Close cash shift
// @Description Close active cashier shift (tutup kasir)
// @Tags CashShifts
// @Accept json
// @Produce json
// @Param request body dto.CloseCashShiftRequest true "Closing cash amount and notes"
// @Success 200 {object} dto.CloseCashShiftResponse
// @Router /cash-shifts/close [post]
func (h *CashShiftHandler) CloseShift(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CloseCashShiftRequest
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
