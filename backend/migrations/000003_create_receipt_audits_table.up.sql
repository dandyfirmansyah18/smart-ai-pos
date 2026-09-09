CREATE TABLE IF NOT EXISTS receipt_audits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_name VARCHAR(255) NOT NULL,
    receipt_date TIMESTAMP WITH TIME ZONE,
    total_amount NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
    raw_ocr_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_receipt_audits_merchant ON receipt_audits(merchant_name);
