package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
	PaymentStatusExpired PaymentStatus = "EXPIRED"
)

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodMidtrans PaymentMethod = "MIDTRANS"
	PaymentMethodQris     PaymentMethod = "QRIS"
	PaymentMethodDebit    PaymentMethod = "DEBIT"
)

type PaymentGateway string

const (
	GatewayMidtrans PaymentGateway = "MIDTRANS"
	GatewayCash     PaymentGateway = "CASH"
)

type OrderPayment struct {
	ID                   uuid.UUID      `json:"id"`
	OrderID              uuid.UUID      `json:"order_id"`
	PaymentMethod        PaymentMethod  `json:"payment_method"`
	Gateway              PaymentGateway `json:"gateway"`
	GatewayTransactionID string         `json:"gateway_transaction_id"`
	Amount               float64        `json:"amount"`
	Status               PaymentStatus  `json:"status"`
	SnapToken            string         `json:"snap_token,omitempty"`
	SnapRedirectURL      string         `json:"snap_redirect_url,omitempty"`
	RawResponse          string         `json:"raw_response,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}
