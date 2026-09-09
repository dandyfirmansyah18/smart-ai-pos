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

func TestOrderPGRepository_CreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewOrderPGRepository(db)
	ctx := context.Background()

	orderID := uuid.New()
	prodID := uuid.New()
	order := &domain.Order{
		ID:             orderID,
		TransactionID:  "TXN-8888",
		TotalAmount:    100.00,
		Status:         domain.StatusPending,
		IdempotencyKey: "IDEM-KEY-8888",
		Items: []domain.OrderItem{
			{
				ID:        uuid.New(),
				OrderID:   orderID,
				ProductID: prodID,
				Quantity:  2,
				UnitPrice: 50.00,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO orders (id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at)")).
		WithArgs(order.ID, "TXN-8888", 100.00, domain.StatusPending, "IDEM-KEY-8888").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at)")).
		WithArgs(sqlmock.AnyArg(), order.ID, prodID, 2, 50.00).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateOrder(ctx, tx, order)
	if err != nil {
		t.Fatalf("expected no error creating order, got %v", err)
	}
}

func TestOrderPGRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewOrderPGRepository(db)
	ctx := context.Background()

	orderID := uuid.New()
	itemID := uuid.New()
	prodID := uuid.New()
	now := time.Now()

	orderRows := sqlmock.NewRows([]string{"id", "transaction_id", "total_amount", "status", "idempotency_key", "created_at", "updated_at"}).
		AddRow(orderID, "TXN-1234", 150.00, domain.StatusCompleted, "IDEM-1234", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at FROM orders WHERE id = $1")).
		WithArgs(orderID).
		WillReturnRows(orderRows)

	itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "unit_price"}).
		AddRow(itemID, orderID, prodID, 3, 50.00)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, order_id, product_id, quantity, unit_price FROM order_items WHERE order_id = $1")).
		WithArgs(orderID).
		WillReturnRows(itemRows)

	foundOrder, err := repo.GetByID(ctx, orderID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundOrder.TransactionID != "TXN-1234" {
		t.Errorf("expected TXN-1234, got %s", foundOrder.TransactionID)
	}
	if len(foundOrder.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(foundOrder.Items))
	}
}

func TestOrderPGRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := pg.NewOrderPGRepository(db)
	ctx := context.Background()

	orderID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2")).
		WithArgs(domain.StatusCompleted, orderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateStatus(ctx, orderID.String(), domain.StatusCompleted)
	if err != nil {
		t.Fatalf("expected no error updating status, got %v", err)
	}
}
