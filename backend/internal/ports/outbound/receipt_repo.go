package outbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type ReceiptRepository interface {
	SaveAudit(ctx context.Context, audit *domain.ReceiptAudit) error
	ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error)
}
