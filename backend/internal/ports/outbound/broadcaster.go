package outbound

import "github.com/pos-backend/internal/domain"

type EventBroadcaster interface {
	BroadcastStockUpdate(sku string, newStock int)
	BroadcastOrderStatusUpdate(orderID string, status domain.OrderStatus)
}
