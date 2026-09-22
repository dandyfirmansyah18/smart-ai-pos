package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type OrderSQLiteRepository struct {
	db *sql.DB
}

func NewOrderSQLiteRepository(db *sql.DB) *OrderSQLiteRepository {
	return &OrderSQLiteRepository{db: db}
}

func (r *OrderSQLiteRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	orderQuery := `INSERT INTO orders (id, transaction_id, total_amount, status, idempotency_key, sync_status, created_at, updated_at) 
	               VALUES (?, ?, ?, ?, ?, 'PENDING', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, orderQuery, order.ID.String(), order.TransactionID, order.TotalAmount, order.Status, order.IdempotencyKey)
	} else {
		_, err = r.db.ExecContext(ctx, orderQuery, order.ID.String(), order.TransactionID, order.TotalAmount, order.Status, order.IdempotencyKey)
	}

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at) 
	              VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	for i := range order.Items {
		item := &order.Items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.OrderID = order.ID

		if tx != nil {
			_, err = tx.ExecContext(ctx, itemQuery, item.ID.String(), item.OrderID.String(), item.ProductID.String(), item.Quantity, item.UnitPrice)
		} else {
			_, err = r.db.ExecContext(ctx, itemQuery, item.ID.String(), item.OrderID.String(), item.ProductID.String(), item.Quantity, item.UnitPrice)
		}

		if err != nil {
			return fmt.Errorf("failed to insert order item (Product ID %s): %w", item.ProductID, err)
		}
	}

	return nil
}

func (r *OrderSQLiteRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	orderQuery := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	               FROM orders WHERE id = ?`

	var o domain.Order
	var idStr string
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&idStr, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}
	o.ID, _ = uuid.Parse(idStr)

	items, err := r.getOrderItems(ctx, nil, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderSQLiteRepository) GetByIdempotencyKey(ctx context.Context, tx *sql.Tx, key string) (*domain.Order, error) {
	if key == "" {
		return nil, domain.ErrOrderNotFound
	}

	orderQuery := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	               FROM orders WHERE idempotency_key = ?`

	var o domain.Order
	var idStr string
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, orderQuery, key).Scan(
			&idStr, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
		)
	} else {
		err = r.db.QueryRowContext(ctx, orderQuery, key).Scan(
			&idStr, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
		)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by idempotency key: %w", err)
	}
	o.ID, _ = uuid.Parse(idStr)

	items, err := r.getOrderItems(ctx, tx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderSQLiteRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderSQLiteRepository) ListActiveOrders(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	           FROM orders WHERE status IN ('PENDING', 'PREPARING', 'READY') ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active orders: %w", err)
	}

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		var idStr string
		if err := rows.Scan(&idStr, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		o.ID, _ = uuid.Parse(idStr)
		orders = append(orders, o)
	}
	rows.Close()

	for i := range orders {
		items, err := r.getOrderItems(ctx, nil, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderSQLiteRepository) getOrderItems(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) ([]domain.OrderItem, error) {
	itemsQuery := `SELECT oi.id, oi.order_id, oi.product_id, COALESCE(p.sku, ''), COALESCE(p.name, 'Product'), oi.quantity, oi.unit_price 
	               FROM order_items oi LEFT JOIN products p ON oi.product_id = p.id WHERE oi.order_id = ?`

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, itemsQuery, orderID.String())
	} else {
		rows, err = r.db.QueryContext(ctx, itemsQuery, orderID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var idStr, orderIDStr, prodIDStr string
		if err := rows.Scan(&idStr, &orderIDStr, &prodIDStr, &item.SKU, &item.Name, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("failed to scan order item row: %w", err)
		}
		item.ID, _ = uuid.Parse(idStr)
		item.OrderID, _ = uuid.Parse(orderIDStr)
		item.ProductID, _ = uuid.Parse(prodIDStr)
		items = append(items, item)
	}

	return items, nil
}

var _ outbound.OrderRepository = (*OrderSQLiteRepository)(nil)
