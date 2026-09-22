package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type FinanceSQLiteRepository struct {
	db *sql.DB
}

func NewFinanceSQLiteRepository(db *sql.DB) *FinanceSQLiteRepository {
	return &FinanceSQLiteRepository{db: db}
}

func (r *FinanceSQLiteRepository) ListOrderHistory(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	          FROM orders ORDER BY created_at DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order history: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		var idStr string
		if err := rows.Scan(&idStr, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		o.ID, _ = uuid.Parse(idStr)

		itemRows, err := r.db.QueryContext(ctx, `SELECT oi.id, oi.order_id, oi.product_id, COALESCE(p.sku, ''), COALESCE(p.name, 'Product'), oi.quantity, oi.unit_price FROM order_items oi LEFT JOIN products p ON oi.product_id = p.id WHERE oi.order_id = ?`, o.ID.String())
		if err == nil {
			var items []domain.OrderItem
			for itemRows.Next() {
				var item domain.OrderItem
				var itemIdStr, orderIdStr, prodIdStr string
				if err := itemRows.Scan(&itemIdStr, &orderIdStr, &prodIdStr, &item.SKU, &item.Name, &item.Quantity, &item.UnitPrice); err == nil {
					item.ID, _ = uuid.Parse(itemIdStr)
					item.OrderID, _ = uuid.Parse(orderIdStr)
					item.ProductID, _ = uuid.Parse(prodIdStr)
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

	return orders, nil
}

func (r *FinanceSQLiteRepository) CalculateProfitLoss(ctx context.Context) (totalRevenue, totalCOGS, grossProfit, margin float64, err error) {
	_ = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status != 'CANCELLED'").Scan(&totalRevenue)

	totalCOGS = totalRevenue * 0.55
	grossProfit = totalRevenue - totalCOGS
	if totalRevenue > 0 {
		margin = (grossProfit / totalRevenue) * 100
	}

	return totalRevenue, totalCOGS, grossProfit, margin, nil
}

var _ outbound.FinanceRepository = (*FinanceSQLiteRepository)(nil)
