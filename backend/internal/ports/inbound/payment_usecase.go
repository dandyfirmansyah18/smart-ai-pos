package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type PaymentUseCase interface {
	CreatePayment(ctx context.Context, orderID string, amount float64, paymentMethod domain.PaymentMethod) (*domain.OrderPayment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error)
	HandleWebhook(ctx context.Context, gatewayTxID string, transactionStatus string, rawPayload string) error
}
