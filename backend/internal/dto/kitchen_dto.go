package dto

// UpdateOrderStatusRequest defines request DTO for updating order status in KDS.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=PENDING PREPARING READY SERVED COMPLETED CANCELLED"`
}
