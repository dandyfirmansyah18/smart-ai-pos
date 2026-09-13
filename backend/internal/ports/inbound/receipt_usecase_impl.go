package inbound

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type ReceiptUseCaseImpl struct {
	visionClient outbound.VisionClient
	receiptRepo  outbound.ReceiptRepository
}

func NewReceiptUseCaseImpl(
	visionClient outbound.VisionClient,
	receiptRepo outbound.ReceiptRepository,
) *ReceiptUseCaseImpl {
	return &ReceiptUseCaseImpl{
		visionClient: visionClient,
		receiptRepo:  receiptRepo,
	}
}

func (s *ReceiptUseCaseImpl) ScanReceipt(ctx context.Context, filename string, imageBytes []byte) (*domain.ReceiptAudit, error) {
	if len(imageBytes) == 0 {
		return nil, domain.ErrInvalidReceipt
	}

	ocrResult, err := s.visionClient.AnalyzeReceipt(ctx, imageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed AI receipt OCR analysis: %w", err)
	}

	audit := &domain.ReceiptAudit{
		ID:           uuid.New(),
		MerchantName: ocrResult.MerchantName,
		ReceiptDate:  ocrResult.Date,
		TotalAmount:  ocrResult.TotalAmount,
		RawOCRJSON:   *ocrResult,
		CreatedAt:    time.Now(),
	}

	if s.receiptRepo != nil {
		if err := s.receiptRepo.SaveAudit(ctx, audit); err != nil {
			return nil, fmt.Errorf("failed to save receipt audit: %w", err)
		}
	}

	return audit, nil
}

func (s *ReceiptUseCaseImpl) ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error) {
	if s.receiptRepo == nil {
		return []domain.ReceiptAudit{}, nil
	}
	return s.receiptRepo.ListAudits(ctx)
}

// Compile-time check
var _ ReceiptUseCase = (*ReceiptUseCaseImpl)(nil)
