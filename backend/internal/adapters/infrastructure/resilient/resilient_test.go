package resilient_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/infrastructure/resilient"
	"github.com/pos-backend/internal/adapters/infrastructure/sqlite"
	"github.com/pos-backend/internal/domain"
)

func setupTestSQLite(t *testing.T) *sql.DB {
	db, err := sqlite.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory sqlite db: %v", err)
	}
	return db
}

func TestResilientDBManager_PGFailoverToSQLite(t *testing.T) {
	sqliteDB := setupTestSQLite(t)
	defer sqliteDB.Close()

	// Simulate closed/unreachable Postgres DB connection
	closedPGDB, err := sql.Open("pgx", "postgres://invalid:invalid@127.0.0.1:9999/invalid?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatalf("failed to open dummy pg db: %v", err)
	}
	defer closedPGDB.Close()

	mgr := resilient.NewResilientDBManager(closedPGDB, sqliteDB, true)
	ctx := context.Background()

	if mgr.IsPGHealthy(ctx) {
		t.Errorf("expected PG health check to return false for unreachable DB host")
	}

	// BeginTx should fallback to SQLite without error
	tx, isSQLite, err := mgr.BeginTx(ctx)
	if err != nil {
		t.Fatalf("expected BeginTx to succeed via SQLite fallback, got error: %v", err)
	}
	if !isSQLite {
		t.Errorf("expected isSQLite to be true for fallback transaction")
	}
	_ = tx.Rollback()
}

func TestResilientOrderRepository_ListAndStatusUpdateOffline(t *testing.T) {
	sqliteDB := setupTestSQLite(t)
	defer sqliteDB.Close()

	mgr := resilient.NewResilientDBManager(nil, sqliteDB, true)
	orderRepo := resilient.NewResilientOrderRepository(mgr)
	ctx := context.Background()

	// 1. Create order in SQLite fallback
	orderID := uuid.New()
	order := &domain.Order{
		ID:             orderID,
		TransactionID:  "TXN-TEST-100",
		TotalAmount:    25.50,
		Status:         domain.OrderStatusPending,
		IdempotencyKey: "IDEM-RESILIENT-1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := orderRepo.CreateOrder(ctx, nil, order)
	if err != nil {
		t.Fatalf("expected CreateOrder to succeed in SQLite fallback, got %v", err)
	}

	// 2. Kitchen / Active orders list must return the order
	activeOrders, err := orderRepo.ListActiveOrders(ctx)
	if err != nil {
		t.Fatalf("expected ListActiveOrders to succeed in SQLite fallback, got %v", err)
	}
	if len(activeOrders) != 1 {
		t.Fatalf("expected 1 active order, got %d", len(activeOrders))
	}
	if activeOrders[0].ID != orderID {
		t.Errorf("expected order ID %s, got %s", orderID, activeOrders[0].ID)
	}

	// 3. Update order status to PREPARING
	err = orderRepo.UpdateStatus(ctx, orderID.String(), domain.OrderStatusPreparing)
	if err != nil {
		t.Fatalf("expected UpdateStatus to succeed in SQLite fallback, got %v", err)
	}

	// 4. Verify updated status
	fetchedOrder, err := orderRepo.GetByID(ctx, orderID.String())
	if err != nil {
		t.Fatalf("expected GetByID to succeed, got %v", err)
	}
	if fetchedOrder.Status != domain.OrderStatusPreparing {
		t.Errorf("expected status PREPARING, got %s", fetchedOrder.Status)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
