package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
)

type CashShiftHandler struct {
	db *sql.DB
}

func NewCashShiftHandler(db *sql.DB) *CashShiftHandler {
	return &CashShiftHandler{db: db}
}

// OpenShift POST /api/cash-shifts/open
func (h *CashShiftHandler) OpenShift(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
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

	// Check if user already has an open shift
	var activeID string
	err = h.db.QueryRowContext(c.Request.Context(), "SELECT id FROM cash_shifts WHERE user_id = $1 AND status = 'OPEN'", userID).Scan(&activeID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you already have an open cash shift"})
		return
	}

	shiftID := uuid.New()
	query := `INSERT INTO cash_shifts (id, user_id, status, opening_cash, notes, opened_at) 
	          VALUES ($1, $2, 'OPEN', $3, $4, CURRENT_TIMESTAMP)`

	_, err = h.db.ExecContext(c.Request.Context(), query, shiftID, userID, req.OpeningCash, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open cash shift", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "cash shift opened successfully", "shift_id": shiftID})
}

// GetCurrentShift GET /api/cash-shifts/current
func (h *CashShiftHandler) GetCurrentShift(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(string)

	query := `SELECT id, user_id, status, opening_cash, closing_cash, expected_cash, total_cash_sales, total_qris_sales, total_debit_sales, notes, opened_at, closed_at 
	          FROM cash_shifts WHERE user_id = $1 AND status = 'OPEN' LIMIT 1`

	var s domain.CashShift
	err := h.db.QueryRowContext(c.Request.Context(), query, userID).Scan(
		&s.ID, &s.UserID, &s.Status, &s.OpeningCash, &s.ClosingCash, &s.ExpectedCash,
		&s.TotalCashSales, &s.TotalQrisSales, &s.TotalDebitSales, &s.Notes, &s.OpenedAt, &s.ClosedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "no open cash shift found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current shift"})
		return
	}

	c.JSON(http.StatusOK, s)
}

// CloseShift POST /api/cash-shifts/close
func (h *CashShiftHandler) CloseShift(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(string)

	var req struct {
		ClosingCash float64 `json:"closing_cash" binding:"required,gte=0"`
		Notes       string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	var shiftID uuid.UUID
	var openingCash float64
	err := h.db.QueryRowContext(c.Request.Context(), "SELECT id, opening_cash FROM cash_shifts WHERE user_id = $1 AND status = 'OPEN' LIMIT 1", userID).Scan(&shiftID, &openingCash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no open cash shift found for user"})
		return
	}

	// Calculate sales during shift from orders
	var totalSales float64
	_ = h.db.QueryRowContext(c.Request.Context(), "SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE created_at >= (SELECT opened_at FROM cash_shifts WHERE id = $1) AND status != 'CANCELLED'", shiftID).Scan(&totalSales)

	expectedCash := openingCash + totalSales

	updateQuery := `UPDATE cash_shifts SET status = 'CLOSED', closing_cash = $1, expected_cash = $2, total_cash_sales = $3, notes = COALESCE(NULLIF($4, ''), notes), closed_at = CURRENT_TIMESTAMP WHERE id = $5`

	_, err = h.db.ExecContext(c.Request.Context(), updateQuery, req.ClosingCash, expectedCash, totalSales, req.Notes, shiftID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close shift", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "cash shift closed successfully",
		"opening_cash":   openingCash,
		"total_sales":    totalSales,
		"expected_cash":  expectedCash,
		"closing_cash":   req.ClosingCash,
		"difference":     req.ClosingCash - expectedCash,
	})
}
