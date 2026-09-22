package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type PaymentSQLiteRepository struct {
	db *sql.DB
}

func NewPaymentSQLiteRepository(db *sql.DB) *PaymentSQLiteRepository {
	return &PaymentSQLiteRepository{db: db}
}

func (r *PaymentSQLiteRepository) CreatePayment(ctx context.Context, p *domain.OrderPayment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	query := `INSERT INTO order_payments (id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, sync_status, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'PENDING', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID.String(), p.OrderID.String(), p.PaymentMethod, p.Gateway, p.GatewayTransactionID,
		p.Amount, p.Status, p.SnapToken, p.SnapRedirectURL, p.RawResponse,
	)
	if err != nil {
		return fmt.Errorf("failed to create order payment record: %w", err)
	}

	return nil
}

func (r *PaymentSQLiteRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error) {
	query := `SELECT id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at
	          FROM order_payments WHERE order_id = ? LIMIT 1`

	var p domain.OrderPayment
	var idStr, orderIDStr string
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&idStr, &orderIDStr, &p.PaymentMethod, &p.Gateway, &p.GatewayTransactionID,
		&p.Amount, &p.Status, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}
	p.ID, _ = uuid.Parse(idStr)
	p.OrderID, _ = uuid.Parse(orderIDStr)

	return &p, nil
}

func (r *PaymentSQLiteRepository) GetByGatewayTxID(ctx context.Context, txID string) (*domain.OrderPayment, error) {
	query := `SELECT id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at
	          FROM order_payments WHERE gateway_transaction_id = ? OR order_id = ? OR id = ? LIMIT 1`

	var p domain.OrderPayment
	var idStr, orderIDStr string
	err := r.db.QueryRowContext(ctx, query, txID, txID, txID).Scan(
		&idStr, &orderIDStr, &p.PaymentMethod, &p.Gateway, &p.GatewayTransactionID,
		&p.Amount, &p.Status, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get payment by gateway tx ID: %w", err)
	}
	p.ID, _ = uuid.Parse(idStr)
	p.OrderID, _ = uuid.Parse(orderIDStr)

	return &p, nil
}

func (r *PaymentSQLiteRepository) UpdateStatus(ctx context.Context, gatewayTxID string, status domain.PaymentStatus, rawResponse string) error {
	query := `UPDATE order_payments SET status = ?, raw_response = COALESCE(NULLIF(?, ''), raw_response), updated_at = CURRENT_TIMESTAMP WHERE gateway_transaction_id = ? OR order_id = ? OR id = ?`

	_, err := r.db.ExecContext(ctx, query, status, rawResponse, gatewayTxID, gatewayTxID, gatewayTxID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

var _ outbound.PaymentRepository = (*PaymentSQLiteRepository)(nil)
