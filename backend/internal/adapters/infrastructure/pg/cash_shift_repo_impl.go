package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type CashShiftPGRepository struct {
	db *sql.DB
}

func NewCashShiftPGRepository(db *sql.DB) *CashShiftPGRepository {
	return &CashShiftPGRepository{db: db}
}

func (r *CashShiftPGRepository) OpenShift(ctx context.Context, shift *domain.CashShift) error {
	if shift.ID == uuid.Nil {
		shift.ID = uuid.New()
	}

	query := `INSERT INTO cash_shifts (id, user_id, status, opening_cash, notes, opened_at) 
	          VALUES ($1, $2, 'OPEN', $3, $4, CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query, shift.ID, shift.UserID, shift.OpeningCash, shift.Notes)
	if err != nil {
		return fmt.Errorf("failed to open cash shift: %w", err)
	}

	return nil
}

func (r *CashShiftPGRepository) GetCurrentShift(ctx context.Context, userID string) (*domain.CashShift, error) {
	query := `SELECT id, user_id, status, opening_cash, closing_cash, expected_cash, total_cash_sales, total_qris_sales, total_debit_sales, notes, opened_at, closed_at 
	          FROM cash_shifts WHERE user_id = $1 AND status = 'OPEN' LIMIT 1`

	var s domain.CashShift
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.Status, &s.OpeningCash, &s.ClosingCash, &s.ExpectedCash,
		&s.TotalCashSales, &s.TotalQrisSales, &s.TotalDebitSales, &s.Notes, &s.OpenedAt, &s.ClosedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound // or custom err
		}
		return nil, fmt.Errorf("failed to get current shift: %w", err)
	}

	return &s, nil
}

func (r *CashShiftPGRepository) CloseShift(ctx context.Context, shiftID string, closingCash float64, expectedCash float64, totalSales float64, notes string) error {
	parsedID, err := uuid.Parse(shiftID)
	if err != nil {
		return fmt.Errorf("invalid shift ID: %w", err)
	}

	query := `UPDATE cash_shifts SET status = 'CLOSED', closing_cash = $1, expected_cash = $2, total_cash_sales = $3, notes = COALESCE(NULLIF($4, ''), notes), closed_at = CURRENT_TIMESTAMP WHERE id = $5`

	_, err = r.db.ExecContext(ctx, query, closingCash, expectedCash, totalSales, notes, parsedID)
	if err != nil {
		return fmt.Errorf("failed to close shift: %w", err)
	}

	return nil
}

func (r *CashShiftPGRepository) CalculateTotalSales(ctx context.Context, openedAt time.Time) (float64, error) {
	var totalSales float64
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE created_at >= $1 AND status != 'CANCELLED'`
	err := r.db.QueryRowContext(ctx, query, openedAt).Scan(&totalSales)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total sales: %w", err)
	}
	return totalSales, nil
}

var _ outbound.CashShiftRepository = (*CashShiftPGRepository)(nil)
