package outbound

type EventBroadcaster interface {
	BroadcastStockUpdate(sku string, newStock int)
	BroadcastOrderStatusUpdate(orderID string, status string)
}
