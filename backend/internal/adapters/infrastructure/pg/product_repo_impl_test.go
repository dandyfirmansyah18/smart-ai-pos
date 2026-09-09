package pg_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/domain"
)

func TestProductPGRepository_GetBySKU(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewProductPGRepository(db)
	ctx := context.Background()

	prodID := uuid.New()
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "sku", "name", "description", "price", "stock_quantity", "created_at", "updated_at"}).
		AddRow(prodID, "SKU-TEST", "Product Test", "Desc", 99.50, 20, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at FROM products WHERE sku = $1")).
		WithArgs("SKU-TEST").
		WillReturnRows(rows)

	prod, err := repo.GetBySKU(ctx, "SKU-TEST")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if prod.SKU != "SKU-TEST" || prod.StockQuantity != 20 {
		t.Errorf("unexpected product data: %+v", prod)
	}
}

func TestProductPGRepository_GetBySKUWithLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewProductPGRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	prodID := uuid.New()
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "sku", "name", "description", "price", "stock_quantity", "created_at", "updated_at"}).
		AddRow(prodID, "SKU-LOCK", "Locked Product", "Desc", 150.00, 10, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at FROM products WHERE sku = $1 FOR UPDATE")).
		WithArgs("SKU-LOCK").
		WillReturnRows(rows)

	prod, err := repo.GetBySKUWithLock(ctx, tx, "SKU-LOCK")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if prod.SKU != "SKU-LOCK" {
		t.Errorf("expected SKU-LOCK, got %s", prod.SKU)
	}
}

func TestProductPGRepository_UpdateStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewProductPGRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE products SET stock_quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE sku = $2")).
		WithArgs(5, "SKU-LOCK").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateStock(ctx, tx, "SKU-LOCK", 5)
	if err != nil {
		t.Fatalf("expected no error updating stock, got %v", err)
	}
}

func TestProductPGRepository_UpdateStock_InsufficientStock(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewProductPGRepository(db)
	ctx := context.Background()

	err = repo.UpdateStock(ctx, nil, "SKU-LOCK", -1)
	if err != domain.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}
