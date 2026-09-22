package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type ReceiptSQLiteRepository struct {
	db *sql.DB
}

func NewReceiptSQLiteRepository(db *sql.DB) *ReceiptSQLiteRepository {
	return &ReceiptSQLiteRepository{db: db}
}

func (r *ReceiptSQLiteRepository) SaveAudit(ctx context.Context, audit *domain.ReceiptAudit) error {
	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}

	jsonBytes, err := json.Marshal(audit.RawOCRJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal raw OCR JSON: %w", err)
	}

	query := `INSERT INTO receipt_audits (id, merchant_name, receipt_date, total_amount, raw_ocr_json, created_at) 
	          VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err = r.db.ExecContext(ctx, query, audit.ID.String(), audit.MerchantName, audit.ReceiptDate, audit.TotalAmount, jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to insert receipt audit record: %w", err)
	}

	return nil
}

func (r *ReceiptSQLiteRepository) ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error) {
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
		var idStr string
		var jsonBytes []byte
		if err := rows.Scan(&idStr, &a.MerchantName, &a.ReceiptDate, &a.TotalAmount, &jsonBytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning receipt audit row: %w", err)
		}
		a.ID, _ = uuid.Parse(idStr)

		if len(jsonBytes) > 0 {
			_ = json.Unmarshal(jsonBytes, &a.RawOCRJSON)
		}

		audits = append(audits, a)
	}

	return audits, nil
}

var _ outbound.ReceiptRepository = (*ReceiptSQLiteRepository)(nil)
