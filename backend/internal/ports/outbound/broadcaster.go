package outbound

type EventBroadcaster interface {
	BroadcastStockUpdate(sku string, newStock int)
}
