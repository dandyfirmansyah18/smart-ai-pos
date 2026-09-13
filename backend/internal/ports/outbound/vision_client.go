package outbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type VisionClient interface {
	AnalyzeReceipt(ctx context.Context, imageBytes []byte) (*domain.ReceiptOCRResult, error)
}
