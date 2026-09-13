package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type ReceiptUseCase interface {
	ScanReceipt(ctx context.Context, filename string, imageBytes []byte) (*domain.ReceiptAudit, error)
	ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error)
}
