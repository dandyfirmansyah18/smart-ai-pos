package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReceiptNotFound = errors.New("receipt audit record not found")
	ErrInvalidReceipt  = errors.New("invalid receipt image or payload")
)

type OCRItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type ReceiptOCRResult struct {
	MerchantName string    `json:"merchant_name"`
	Date         time.Time `json:"date"`
	Items        []OCRItem `json:"items"`
	TotalAmount  float64   `json:"total_amount"`
}

type ReceiptAudit struct {
	ID           uuid.UUID        `json:"id"`
	MerchantName string           `json:"merchant_name"`
	ReceiptDate  time.Time        `json:"receipt_date"`
	TotalAmount  float64          `json:"total_amount"`
	RawOCRJSON   ReceiptOCRResult `json:"raw_ocr_json"`
	CreatedAt    time.Time        `json:"created_at"`
}
