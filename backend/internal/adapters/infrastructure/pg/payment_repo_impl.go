package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type PaymentPGRepository struct {
	db *sql.DB
}

func NewPaymentPGRepository(db *sql.DB) *PaymentPGRepository {
	return &PaymentPGRepository{db: db}
}

func (r *PaymentPGRepository) CreatePayment(ctx context.Context, p *domain.OrderPayment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	query := `INSERT INTO order_payments (id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.OrderID, p.PaymentMethod, p.Gateway, p.GatewayTransactionID,
		p.Amount, p.Status, p.SnapToken, p.SnapRedirectURL, p.RawResponse,
	)
	if err != nil {
		return fmt.Errorf("failed to create order payment record: %w", err)
	}

	return nil
}

func (r *PaymentPGRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error) {
	parsedID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	query := `SELECT id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at
	          FROM order_payments WHERE order_id = $1 LIMIT 1`

	var p domain.OrderPayment
	err = r.db.QueryRowContext(ctx, query, parsedID).Scan(
		&p.ID, &p.OrderID, &p.PaymentMethod, &p.Gateway, &p.GatewayTransactionID,
		&p.Amount, &p.Status, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	return &p, nil
}

func (r *PaymentPGRepository) GetByGatewayTxID(ctx context.Context, txID string) (*domain.OrderPayment, error) {
	query := `SELECT id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at
	          FROM order_payments WHERE gateway_transaction_id = $1 OR order_id::text = $1 OR id::text = $1 LIMIT 1`

	var p domain.OrderPayment
	err := r.db.QueryRowContext(ctx, query, txID).Scan(
		&p.ID, &p.OrderID, &p.PaymentMethod, &p.Gateway, &p.GatewayTransactionID,
		&p.Amount, &p.Status, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get payment by gateway tx ID: %w", err)
	}

	return &p, nil
}

func (r *PaymentPGRepository) UpdateStatus(ctx context.Context, gatewayTxID string, status domain.PaymentStatus, rawResponse string) error {
	query := `UPDATE order_payments SET status = $1, raw_response = COALESCE(NULLIF($2, ''), raw_response), updated_at = CURRENT_TIMESTAMP WHERE gateway_transaction_id = $3 OR order_id::text = $3 OR id::text = $3`

	_, err := r.db.ExecContext(ctx, query, status, rawResponse, gatewayTxID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

var _ outbound.PaymentRepository = (*PaymentPGRepository)(nil)
