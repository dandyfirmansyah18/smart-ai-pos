package inbound

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type PaymentUseCaseImpl struct {
	paymentRepo outbound.PaymentRepository
	gateway     outbound.PaymentGatewayClient
	orderRepo   outbound.OrderRepository
	broadcaster outbound.EventBroadcaster
}

func NewPaymentUseCaseImpl(
	paymentRepo outbound.PaymentRepository,
	gateway outbound.PaymentGatewayClient,
	orderRepo outbound.OrderRepository,
	broadcaster ...outbound.EventBroadcaster,
) *PaymentUseCaseImpl {
	var b outbound.EventBroadcaster
	if len(broadcaster) > 0 {
		b = broadcaster[0]
	}
	return &PaymentUseCaseImpl{
		paymentRepo: paymentRepo,
		gateway:     gateway,
		orderRepo:   orderRepo,
		broadcaster: b,
	}
}

func (u *PaymentUseCaseImpl) CreatePayment(ctx context.Context, orderID string, amount float64, paymentMethod domain.PaymentMethod) (*domain.OrderPayment, error) {
	parsedOrderID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	// Check if a payment record already exists for this orderID
	existing, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err == nil && existing != nil && existing.SnapToken != "" && existing.Status == domain.PaymentStatusPending {
		return existing, nil
	}

	token, redirectURL, err := u.gateway.CreateSnapToken(ctx, orderID, amount, "Valued Merchant Customer")
	if err != nil {
		return nil, fmt.Errorf("failed to generate snap payment token: %w", err)
	}

	if paymentMethod == "" {
		paymentMethod = domain.PaymentMethodMidtrans
	}

	payment := &domain.OrderPayment{
		ID:                   uuid.New(),
		OrderID:              parsedOrderID,
		PaymentMethod:        paymentMethod,
		Gateway:              domain.GatewayMidtrans,
		GatewayTransactionID: token,
		Amount:               amount,
		Status:               domain.PaymentStatusPending,
		SnapToken:            token,
		SnapRedirectURL:      redirectURL,
	}

	if err := u.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *PaymentUseCaseImpl) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error) {
	return u.paymentRepo.GetByOrderID(ctx, orderID)
}

func (u *PaymentUseCaseImpl) HandleWebhook(ctx context.Context, gatewayTxID string, transactionStatus string, rawPayload string) error {
	var dbStatus domain.PaymentStatus
	switch transactionStatus {
	case "capture", "settlement", "success":
		dbStatus = domain.PaymentStatusSuccess
	case "pending":
		dbStatus = domain.PaymentStatusPending
	case "deny", "cancel", "expire", "failure":
		dbStatus = domain.PaymentStatusFailed
	default:
		dbStatus = domain.PaymentStatusPending
	}

	err := u.paymentRepo.UpdateStatus(ctx, gatewayTxID, dbStatus, rawPayload)
	if err != nil {
		return err
	}

	payment, err := u.paymentRepo.GetByGatewayTxID(ctx, gatewayTxID)
	if err == nil && payment != nil {
		if dbStatus == domain.PaymentStatusSuccess {
			// When paid, transition order status to PENDING (paid & queued for kitchen)
			_ = u.orderRepo.UpdateStatus(ctx, payment.OrderID.String(), domain.OrderStatusPending)
			if u.broadcaster != nil {
				u.broadcaster.BroadcastOrderStatusUpdate(payment.OrderID.String(), domain.OrderStatusPending)
			}
		} else if dbStatus == domain.PaymentStatusFailed || dbStatus == domain.PaymentStatusExpired {
			// When payment fails/expires, transition order status to EXPIRED
			_ = u.orderRepo.UpdateStatus(ctx, payment.OrderID.String(), domain.OrderStatusExpired)
			if u.broadcaster != nil {
				u.broadcaster.BroadcastOrderStatusUpdate(payment.OrderID.String(), domain.OrderStatusExpired)
			}
		}
	}

	return nil
}

var _ PaymentUseCase = (*PaymentUseCaseImpl)(nil)
