package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type ReceiptPGRepository struct {
	db *sql.DB
}

func NewReceiptPGRepository(db *sql.DB) *ReceiptPGRepository {
	return &ReceiptPGRepository{db: db}
}

func (r *ReceiptPGRepository) SaveAudit(ctx context.Context, audit *domain.ReceiptAudit) error {
	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}

	jsonBytes, err := json.Marshal(audit.RawOCRJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal raw OCR JSON: %w", err)
	}

	query := `INSERT INTO receipt_audits (id, merchant_name, receipt_date, total_amount, raw_ocr_json, created_at) 
	          VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`

	_, err = r.db.ExecContext(ctx, query, audit.ID, audit.MerchantName, audit.ReceiptDate, audit.TotalAmount, jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to insert receipt audit record: %w", err)
	}

	return nil
}

func (r *ReceiptPGRepository) ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error) {
	query := `SELECT id, merchant_name, receipt_date, total_amount, raw_ocr_json, created_at 
	          FROM receipt_audits ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query receipt audits: %w", err)
	}
	defer rows.Close()

	var audits []domain.ReceiptAudit
	for rows.Next() {
		var a domain.ReceiptAudit
		var jsonBytes []byte
		if err := rows.Scan(&a.ID, &a.MerchantName, &a.ReceiptDate, &a.TotalAmount, &jsonBytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning receipt audit row: %w", err)
		}

		if len(jsonBytes) > 0 {
			_ = json.Unmarshal(jsonBytes, &a.RawOCRJSON)
		}

		audits = append(audits, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating receipt audit rows: %w", err)
	}

	return audits, nil
}

// Compile-time check
var _ outbound.ReceiptRepository = (*ReceiptPGRepository)(nil)
