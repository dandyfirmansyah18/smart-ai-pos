-- Create cash_shifts (Buka-Tutup Kasir) table (UP)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'shift_status') THEN
        CREATE TYPE shift_status AS ENUM ('OPEN', 'CLOSED');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS cash_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    status shift_status NOT NULL DEFAULT 'OPEN',
    opening_cash NUMERIC(12, 2) NOT NULL CHECK (opening_cash >= 0),
    closing_cash NUMERIC(12, 2) DEFAULT 0 CHECK (closing_cash >= 0),
    expected_cash NUMERIC(12, 2) DEFAULT 0 CHECK (expected_cash >= 0),
    total_cash_sales NUMERIC(12, 2) DEFAULT 0 CHECK (total_cash_sales >= 0),
    total_qris_sales NUMERIC(12, 2) DEFAULT 0 CHECK (total_qris_sales >= 0),
    total_debit_sales NUMERIC(12, 2) DEFAULT 0 CHECK (total_debit_sales >= 0),
    notes TEXT,
    opened_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_cash_shifts_user ON cash_shifts(user_id);
CREATE INDEX IF NOT EXISTS idx_cash_shifts_status ON cash_shifts(status);
