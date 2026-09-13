package dto

import "time"

// ReceiptOCRItemDTO defines item DTO extracted from receipt OCR.
type ReceiptOCRItemDTO struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
}

// ReceiptAuditResponse defines response DTO for scanned receipt audit logs.
type ReceiptAuditResponse struct {
	ID           string              `json:"id"`
	FileName     string              `json:"file_name"`
	MerchantName string              `json:"merchant_name"`
	Date         time.Time           `json:"date"`
	TotalAmount  float64             `json:"total_amount"`
	RawResult    string              `json:"raw_result"`
	Items        []ReceiptOCRItemDTO `json:"items,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}
