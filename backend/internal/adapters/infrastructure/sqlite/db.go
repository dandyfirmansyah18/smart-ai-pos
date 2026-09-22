package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create sqlite directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite works best with 1 open connection to avoid database is locked errors

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run sqlite migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		sku TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL CHECK (price >= 0),
		stock_quantity INTEGER NOT NULL CHECK (stock_quantity >= 0),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'CASHIER',
		full_name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		transaction_id TEXT UNIQUE NOT NULL,
		total_amount REAL NOT NULL CHECK (total_amount >= 0),
		status TEXT NOT NULL DEFAULT 'PENDING',
		idempotency_key TEXT UNIQUE,
		sync_status TEXT NOT NULL DEFAULT 'PENDING',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id TEXT PRIMARY KEY,
		order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id TEXT NOT NULL REFERENCES products(id),
		quantity INTEGER NOT NULL CHECK (quantity > 0),
		unit_price REAL NOT NULL CHECK (unit_price >= 0),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS access_menus (
		id TEXT PRIMARY KEY,
		key TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		icon TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS access_menus_per_roles (
		id TEXT PRIMARY KEY,
		role TEXT NOT NULL,
		menu_key TEXT NOT NULL REFERENCES access_menus(key) ON DELETE CASCADE,
		can_access INTEGER NOT NULL DEFAULT 1,
		UNIQUE(role, menu_key)
	);

	CREATE TABLE IF NOT EXISTS cash_shifts (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id),
		status TEXT NOT NULL DEFAULT 'OPEN',
		opening_cash REAL NOT NULL CHECK (opening_cash >= 0),
		closing_cash REAL DEFAULT 0 CHECK (closing_cash >= 0),
		expected_cash REAL DEFAULT 0 CHECK (expected_cash >= 0),
		total_cash_sales REAL DEFAULT 0 CHECK (total_cash_sales >= 0),
		total_qris_sales REAL DEFAULT 0 CHECK (total_qris_sales >= 0),
		total_debit_sales REAL DEFAULT 0 CHECK (total_debit_sales >= 0),
		notes TEXT,
		sync_status TEXT NOT NULL DEFAULT 'PENDING',
		opened_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		closed_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS order_payments (
		id TEXT PRIMARY KEY,
		order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		payment_method TEXT NOT NULL DEFAULT 'CASH',
		gateway TEXT NOT NULL DEFAULT 'MIDTRANS',
		gateway_transaction_id TEXT,
		amount REAL NOT NULL CHECK (amount >= 0),
		status TEXT NOT NULL DEFAULT 'PENDING',
		snap_token TEXT,
		snap_redirect_url TEXT,
		raw_response TEXT,
		sync_status TEXT NOT NULL DEFAULT 'PENDING',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS receipt_audits (
		id TEXT PRIMARY KEY,
		merchant_name TEXT,
		receipt_date TEXT,
		total_amount REAL,
		raw_ocr_json TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(schema)
	return err
}
