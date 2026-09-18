package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
)

type FinanceHandler struct {
	db *sql.DB
}

func NewFinanceHandler(db *sql.DB) *FinanceHandler {
	return &FinanceHandler{db: db}
}

// GetOrderHistory GET /api/orders/history
func (h *FinanceHandler) GetOrderHistory(c *gin.Context) {
	query := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	          FROM orders ORDER BY created_at DESC LIMIT 100`

	rows, err := h.db.QueryContext(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order history", "details": err.Error()})
		return
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan order"})
			return
		}
		
		// Fetch items for order
		itemRows, err := h.db.QueryContext(c.Request.Context(), `SELECT oi.id, oi.order_id, oi.product_id, COALESCE(p.sku, ''), COALESCE(p.name, 'Product'), oi.quantity, oi.unit_price FROM order_items oi LEFT JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`, o.ID)
		if err == nil {
			var items []domain.OrderItem
			for itemRows.Next() {
				var item domain.OrderItem
				if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.SKU, &item.Name, &item.Quantity, &item.UnitPrice); err == nil {
					items = append(items, item)
				}
			}
			itemRows.Close()
			o.Items = items
		}

		orders = append(orders, o)
	}

	if orders == nil {
		orders = []domain.Order{}
	}

	c.JSON(http.StatusOK, orders)
}

// GetProfitLoss GET /api/finance/profit-loss
func (h *FinanceHandler) GetProfitLoss(c *gin.Context) {
	var totalRevenue float64
	_ = h.db.QueryRowContext(c.Request.Context(), "SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status != 'CANCELLED'").Scan(&totalRevenue)

	// Estimated COGS (55% of revenue)
	totalCOGS := totalRevenue * 0.55
	grossProfit := totalRevenue - totalCOGS
	margin := 0.0
	if totalRevenue > 0 {
		margin = (grossProfit / totalRevenue) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_revenue":   totalRevenue,
		"total_cogs":      totalCOGS,
		"gross_profit":    grossProfit,
		"profit_margin":   margin,
	})
}
