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

type ProductSQLiteRepository struct {
	db *sql.DB
}

func NewProductSQLiteRepository(db *sql.DB) *ProductSQLiteRepository {
	return &ProductSQLiteRepository{db: db}
}

func (r *ProductSQLiteRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at 
	          FROM products WHERE sku = ?`

	var p domain.Product
	var idStr string
	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&idStr, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	p.ID, _ = uuid.Parse(idStr)

	return &p, nil
}

func (r *ProductSQLiteRepository) GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error) {
	query := `SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at 
	          FROM products WHERE sku = ?`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, sku)
	} else {
		row = r.db.QueryRowContext(ctx, query, sku)
	}

	var p domain.Product
	var idStr string
	err := row.Scan(
		&idStr, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product with lock by SKU: %w", err)
	}
	p.ID, _ = uuid.Parse(idStr)

	return &p, nil
}

func (r *ProductSQLiteRepository) UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error {
	if newQty < 0 {
		return domain.ErrInsufficientStock
	}

	query := `UPDATE products SET stock_quantity = ?, updated_at = CURRENT_TIMESTAMP WHERE sku = ?`

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

func (r *ProductSQLiteRepository) ListAll(ctx context.Context) ([]domain.Product, error) {
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
		var idStr string
		if err := rows.Scan(
			&idStr, &p.SKU, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		p.ID, _ = uuid.Parse(idStr)
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductSQLiteRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	query := `INSERT INTO products (id, sku, name, description, price, stock_quantity, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query, product.ID.String(), product.SKU, product.Name, product.Description, product.Price, product.StockQuantity)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

var _ outbound.ProductRepository = (*ProductSQLiteRepository)(nil)
