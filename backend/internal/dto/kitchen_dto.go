package dto

import "github.com/pos-backend/internal/domain"

// UpdateOrderStatusRequest defines request DTO for updating order status in KDS.
type UpdateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" binding:"required,oneof=PENDING PREPARING READY SERVED COMPLETED CANCELLED"`
}
