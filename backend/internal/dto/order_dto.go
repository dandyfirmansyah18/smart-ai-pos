package dto

import "time"

// CheckoutItemRequest defines the request DTO for an item in a checkout order.
type CheckoutItemRequest struct {
	SKU      string `json:"sku" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

// CheckoutRequest defines the request DTO for creating an order checkout.
type CheckoutRequest struct {
	IdempotencyKey string                `json:"idempotency_key" binding:"required"`
	Items          []CheckoutItemRequest `json:"items" binding:"required,dive"`
}

// StockUpdateInfo defines DTO for real-time stock updates.
type StockUpdateInfo struct {
	SKU      string `json:"sku"`
	NewStock int    `json:"new_stock"`
}

// OrderItemResponse defines the response DTO for an order item.
type OrderItemResponse struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// OrderResponse defines the response DTO for an order.
type OrderResponse struct {
	ID             string              `json:"id"`
	TransactionID  string              `json:"transaction_id"`
	TotalAmount    float64             `json:"total_amount"`
	Status         string              `json:"status"`
	IdempotencyKey string              `json:"idempotency_key"`
	Items          []OrderItemResponse `json:"items"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}
