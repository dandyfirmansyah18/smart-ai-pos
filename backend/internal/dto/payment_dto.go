package dto

import "github.com/pos-backend/internal/domain"

type CreatePaymentChargeRequest struct {
	OrderID       string               `json:"order_id" binding:"required"`
	Amount        float64              `json:"amount" binding:"required,gt=0"`
	PaymentMethod domain.PaymentMethod `json:"payment_method"`
}

type MidtransWebhookRequest struct {
	OrderID           string `json:"order_id" binding:"required"`
	TransactionStatus string `json:"transaction_status"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key,omitempty"`
}

type PaymentChargeResponse struct {
	ID                   string               `json:"id"`
	OrderID              string               `json:"order_id"`
	PaymentMethod        domain.PaymentMethod `json:"payment_method"`
	Gateway              domain.PaymentGateway `json:"gateway"`
	GatewayTransactionID string               `json:"gateway_transaction_id"`
	Amount               float64              `json:"amount"`
	Status               domain.PaymentStatus `json:"status"`
	SnapToken            string               `json:"snap_token,omitempty"`
	SnapRedirectURL      string               `json:"snap_redirect_url,omitempty"`
}

type MidtransTransactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type MidtransCustomerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type MidtransSnapRequest struct {
	TransactionDetails MidtransTransactionDetails `json:"transaction_details"`
	CustomerDetails    MidtransCustomerDetails    `json:"customer_details,omitempty"`
}

type MidtransSnapResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}
