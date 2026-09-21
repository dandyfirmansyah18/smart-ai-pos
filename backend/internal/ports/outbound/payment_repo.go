package outbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *domain.OrderPayment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error)
	GetByGatewayTxID(ctx context.Context, txID string) (*domain.OrderPayment, error)
	UpdateStatus(ctx context.Context, gatewayTxID string, status domain.PaymentStatus, rawResponse string) error
}

type PaymentGatewayClient interface {
	CreateSnapToken(ctx context.Context, orderID string, amount float64, customerName string) (token string, redirectURL string, err error)
}
