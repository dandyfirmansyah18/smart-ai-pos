package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type OrderPGRepository struct {
	db *sql.DB
}

func NewOrderPGRepository(db *sql.DB) *OrderPGRepository {
	return &OrderPGRepository{db: db}
}

func (r *OrderPGRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	orderQuery := `INSERT INTO orders (id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at) 
	               VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, orderQuery, order.ID, order.TransactionID, order.TotalAmount, order.Status, order.IdempotencyKey)
	} else {
		_, err = r.db.ExecContext(ctx, orderQuery, order.ID, order.TransactionID, order.TotalAmount, order.Status, order.IdempotencyKey)
	}

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at) 
	              VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`

	for i := range order.Items {
		item := &order.Items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.OrderID = order.ID

		if tx != nil {
			_, err = tx.ExecContext(ctx, itemQuery, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
		} else {
			_, err = r.db.ExecContext(ctx, itemQuery, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
		}

		if err != nil {
			return fmt.Errorf("failed to insert order item (Product ID %s): %w", item.ProductID, err)
		}
	}

	return nil
}

func (r *OrderPGRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID format: %w", err)
	}

	orderQuery := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	               FROM orders WHERE id = $1`

	var o domain.Order
	err = r.db.QueryRowContext(ctx, orderQuery, parsedID).Scan(
		&o.ID, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderPGRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	if key == "" {
		return nil, domain.ErrOrderNotFound
	}

	orderQuery := `SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at 
	               FROM orders WHERE idempotency_key = $1`

	var o domain.Order
	err := r.db.QueryRowContext(ctx, orderQuery, key).Scan(
		&o.ID, &o.TransactionID, &o.TotalAmount, &o.Status, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by idempotency key: %w", err)
	}

	items, err := r.getOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderPGRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid order ID format: %w", err)
	}

	query := `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, status, parsedID)
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

func (r *OrderPGRepository) getOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	itemsQuery := `SELECT id, order_id, product_id, quantity, unit_price 
	               FROM order_items WHERE order_id = $1`

	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("failed to scan order item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order item rows: %w", err)
	}

	return items, nil
}

// Compile-time check
var _ outbound.OrderRepository = (*OrderPGRepository)(nil)
