package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type ProductPGRepository struct {
	db *sql.DB
}

func NewProductPGRepository(db *sql.DB) *ProductPGRepository {
	return &ProductPGRepository{db: db}
}

func (r *ProductPGRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at 
	          FROM products WHERE sku = $1`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return &p, nil
}

func (r *ProductPGRepository) GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error) {
	query := `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at 
	          FROM products WHERE sku = $1 FOR UPDATE`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, sku)
	} else {
		row = r.db.QueryRowContext(ctx, query, sku)
	}

	var p domain.Product
	err := row.Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product with lock by SKU: %w", err)
	}

	return &p, nil
}

func (r *ProductPGRepository) UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error {
	if newQty < 0 {
		return domain.ErrInsufficientStock
	}

	query := `UPDATE products SET stock_quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE sku = $2`

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.ExecContext(ctx, query, newQty, sku)
	} else {
		res, err = r.db.ExecContext(ctx, query, newQty, sku)
	}

	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductPGRepository) ListAll(ctx context.Context) ([]domain.Product, error) {
	query := `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at 
	          FROM products ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}

// Compile-time check
var _ outbound.ProductRepository = (*ProductPGRepository)(nil)
