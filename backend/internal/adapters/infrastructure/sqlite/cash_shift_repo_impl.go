package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type CashShiftSQLiteRepository struct {
	db *sql.DB
}

func NewCashShiftSQLiteRepository(db *sql.DB) *CashShiftSQLiteRepository {
	return &CashShiftSQLiteRepository{db: db}
}

func (r *CashShiftSQLiteRepository) OpenShift(ctx context.Context, shift *domain.CashShift) error {
	if shift.ID == uuid.Nil {
		shift.ID = uuid.New()
	}

	query := `INSERT INTO cash_shifts (id, user_id, status, opening_cash, notes, sync_status, opened_at) 
	          VALUES (?, ?, 'OPEN', ?, ?, 'PENDING', CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query, shift.ID.String(), shift.UserID.String(), shift.OpeningCash, shift.Notes)
	if err != nil {
		return fmt.Errorf("failed to open cash shift: %w", err)
	}

	return nil
}

func (r *CashShiftSQLiteRepository) GetCurrentShift(ctx context.Context, userID string) (*domain.CashShift, error) {
	query := `SELECT id, user_id, status, opening_cash, closing_cash, expected_cash, total_cash_sales, total_qris_sales, total_debit_sales, notes, opened_at, closed_at 
	          FROM cash_shifts WHERE user_id = ? AND status = 'OPEN' LIMIT 1`

	var s domain.CashShift
	var idStr, userIdStr string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&idStr, &userIdStr, &s.Status, &s.OpeningCash, &s.ClosingCash, &s.ExpectedCash,
		&s.TotalCashSales, &s.TotalQrisSales, &s.TotalDebitSales, &s.Notes, &s.OpenedAt, &s.ClosedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get current shift: %w", err)
	}
	s.ID, _ = uuid.Parse(idStr)
	s.UserID, _ = uuid.Parse(userIdStr)

	return &s, nil
}

func (r *CashShiftSQLiteRepository) CloseShift(ctx context.Context, shiftID string, closingCash float64, expectedCash float64, totalSales float64, notes string) error {
	query := `UPDATE cash_shifts SET status = 'CLOSED', closing_cash = ?, expected_cash = ?, total_cash_sales = ?, notes = COALESCE(NULLIF(?, ''), notes), closed_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, closingCash, expectedCash, totalSales, notes, shiftID)
	if err != nil {
		return fmt.Errorf("failed to close shift: %w", err)
	}

	return nil
}

func (r *CashShiftSQLiteRepository) CalculateTotalSales(ctx context.Context, openedAt time.Time) (float64, error) {
	var totalSales float64
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE created_at >= ? AND status != 'CANCELLED'`
	err := r.db.QueryRowContext(ctx, query, openedAt).Scan(&totalSales)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total sales: %w", err)
	}
	return totalSales, nil
}

var _ outbound.CashShiftRepository = (*CashShiftSQLiteRepository)(nil)
