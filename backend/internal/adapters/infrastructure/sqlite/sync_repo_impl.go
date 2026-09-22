package sqlite

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pos-backend/internal/ports/outbound"
)

type SyncSQLiteRepository struct {
	localDB *sql.DB
}

func NewSyncSQLiteRepository(localDB *sql.DB) *SyncSQLiteRepository {
	return &SyncSQLiteRepository{localDB: localDB}
}

func (r *SyncSQLiteRepository) GetPendingCounts(ctx context.Context) (*outbound.SyncCounts, error) {
	if r.localDB == nil {
		return &outbound.SyncCounts{}, nil
	}

	var pendingOrders, pendingShifts, pendingPayments int

	_ = r.localDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE sync_status = 'PENDING'").Scan(&pendingOrders)
	_ = r.localDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM cash_shifts WHERE sync_status = 'PENDING'").Scan(&pendingShifts)
	_ = r.localDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM order_payments WHERE sync_status = 'PENDING'").Scan(&pendingPayments)

	return &outbound.SyncCounts{
		PendingOrders:   pendingOrders,
		PendingShifts:   pendingShifts,
		PendingPayments: pendingPayments,
	}, nil
}

func (r *SyncSQLiteRepository) CheckPostgresHealth(ctx context.Context, dsn string) bool {
	if dsn == "" {
		return false
	}
	pgDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return false
	}
	defer pgDB.Close()

	return pgDB.PingContext(ctx) == nil
}

func (r *SyncSQLiteRepository) SyncMasterDataFromPostgres(ctx context.Context, dsn string) error {
	if r.localDB == nil {
		return nil
	}

	pgDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer pgDB.Close()

	// 1. Pull products
	rows, err := pgDB.QueryContext(ctx, "SELECT id, sku, name, description, price, stock_quantity, created_at, updated_at FROM products")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, sku, name, description string
			var price float64
			var stock int
			var createdAt, updatedAt time.Time
			if err := rows.Scan(&id, &sku, &name, &description, &price, &stock, &createdAt, &updatedAt); err == nil {
				_, _ = r.localDB.ExecContext(ctx, `
					INSERT INTO products (id, sku, name, description, price, stock_quantity, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
					ON CONFLICT(sku) DO UPDATE SET price = excluded.price, stock_quantity = excluded.stock_quantity, updated_at = excluded.updated_at
				`, id, sku, name, description, price, stock, createdAt, updatedAt)
			}
		}
	}

	// 2. Pull users
	userRows, err := pgDB.QueryContext(ctx, "SELECT id, username, password_hash, role, full_name, created_at, updated_at FROM users")
	if err == nil {
		defer userRows.Close()
		for userRows.Next() {
			var id, username, passwordHash, role, fullName string
			var createdAt, updatedAt time.Time
			if err := userRows.Scan(&id, &username, &passwordHash, &role, &fullName, &createdAt, &updatedAt); err == nil {
				_, _ = r.localDB.ExecContext(ctx, `
					INSERT INTO users (id, username, password_hash, role, full_name, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?)
					ON CONFLICT(username) DO UPDATE SET role = excluded.role, password_hash = excluded.password_hash, updated_at = excluded.updated_at
				`, id, username, passwordHash, role, fullName, createdAt, updatedAt)
			}
		}
	}

	// 3. Pull access menus & role configs
	menuRows, err := pgDB.QueryContext(ctx, "SELECT id, key, name, path, icon, created_at FROM access_menus")
	if err == nil {
		defer menuRows.Close()
		for menuRows.Next() {
			var id, key, name, path, icon string
			var createdAt time.Time
			if err := menuRows.Scan(&id, &key, &name, &path, &icon, &createdAt); err == nil {
				_, _ = r.localDB.ExecContext(ctx, `
					INSERT INTO access_menus (id, key, name, path, icon, created_at)
					VALUES (?, ?, ?, ?, ?, ?)
					ON CONFLICT(key) DO NOTHING
				`, id, key, name, path, icon, createdAt)
			}
		}
	}

	roleRows, err := pgDB.QueryContext(ctx, "SELECT id, role, menu_key, can_access FROM access_menus_per_roles")
	if err == nil {
		defer roleRows.Close()
		for roleRows.Next() {
			var id, role, menuKey string
			var canAccess bool
			if err := roleRows.Scan(&id, &role, &menuKey, &canAccess); err == nil {
				_, _ = r.localDB.ExecContext(ctx, `
					INSERT INTO access_menus_per_roles (id, role, menu_key, can_access)
					VALUES (?, ?, ?, ?)
					ON CONFLICT(role, menu_key) DO UPDATE SET can_access = excluded.can_access
				`, id, role, menuKey, canAccess)
			}
		}
	}

	return nil
}

func (r *SyncSQLiteRepository) PushPendingDataToPostgres(ctx context.Context, dsn string) error {
	if r.localDB == nil {
		return nil
	}

	pgDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer pgDB.Close()

	// 1. Push PENDING ORDERS
	orderRows, err := r.localDB.QueryContext(ctx, "SELECT id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at FROM orders WHERE sync_status = 'PENDING'")
	if err == nil {
		defer orderRows.Close()
		for orderRows.Next() {
			var id, txID, status string
			var idempotencyKey sql.NullString
			var totalAmount float64
			var createdAt, updatedAt time.Time
			if err := orderRows.Scan(&id, &txID, &totalAmount, &status, &idempotencyKey, &createdAt, &updatedAt); err == nil {
				_, pgErr := pgDB.ExecContext(ctx, `
					INSERT INTO orders (id, transaction_id, total_amount, status, idempotency_key, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					ON CONFLICT (transaction_id) DO NOTHING
				`, id, txID, totalAmount, status, idempotencyKey, createdAt, updatedAt)

				if pgErr == nil {
					itemRows, itemErr := r.localDB.QueryContext(ctx, "SELECT id, order_id, product_id, quantity, unit_price, created_at FROM order_items WHERE order_id = ?", id)
					if itemErr == nil {
						for itemRows.Next() {
							var itemId, orderId, prodId string
							var qty int
							var unitPrice float64
							var itemCreated time.Time
							if scanErr := itemRows.Scan(&itemId, &orderId, &prodId, &qty, &unitPrice, &itemCreated); scanErr == nil {
								_, _ = pgDB.ExecContext(ctx, `
									INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at)
									VALUES ($1, $2, $3, $4, $5, $6)
									ON CONFLICT (id) DO NOTHING
								`, itemId, orderId, prodId, qty, unitPrice, itemCreated)
							}
						}
						itemRows.Close()
					}

					_, _ = r.localDB.ExecContext(ctx, "UPDATE orders SET sync_status = 'SYNCED' WHERE id = ?", id)
				}
			}
		}
	}

	// 2. Push PENDING CASH SHIFTS
	shiftRows, err := r.localDB.QueryContext(ctx, "SELECT id, user_id, status, opening_cash, closing_cash, expected_cash, total_cash_sales, total_qris_sales, total_debit_sales, notes, opened_at, closed_at FROM cash_shifts WHERE sync_status = 'PENDING'")
	if err == nil {
		defer shiftRows.Close()
		for shiftRows.Next() {
			var id, userID, status string
			var openingCash, closingCash, expectedCash, totalCashSales, totalQrisSales, totalDebitSales float64
			var notes sql.NullString
			var openedAt time.Time
			var closedAt sql.NullTime
			if err := shiftRows.Scan(&id, &userID, &status, &openingCash, &closingCash, &expectedCash, &totalCashSales, &totalQrisSales, &totalDebitSales, &notes, &openedAt, &closedAt); err == nil {
				_, pgErr := pgDB.ExecContext(ctx, `
					INSERT INTO cash_shifts (id, user_id, status, opening_cash, closing_cash, expected_cash, total_cash_sales, total_qris_sales, total_debit_sales, notes, opened_at, closed_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
					ON CONFLICT (id) DO NOTHING
				`, id, userID, status, openingCash, closingCash, expectedCash, totalCashSales, totalQrisSales, totalDebitSales, notes, openedAt, closedAt)

				if pgErr == nil {
					_, _ = r.localDB.ExecContext(ctx, "UPDATE cash_shifts SET sync_status = 'SYNCED' WHERE id = ?", id)
				}
			}
		}
	}

	// 3. Push PENDING PAYMENTS
	payRows, err := r.localDB.QueryContext(ctx, "SELECT id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at FROM order_payments WHERE sync_status = 'PENDING'")
	if err == nil {
		defer payRows.Close()
		for payRows.Next() {
			var id, orderID, payMethod, gateway, status string
			var gwTxID, snapToken, snapRedirect, rawResp sql.NullString
			var amount float64
			var createdAt, updatedAt time.Time
			if err := payRows.Scan(&id, &orderID, &payMethod, &gateway, &gwTxID, &amount, &status, &snapToken, &snapRedirect, &rawResp, &createdAt, &updatedAt); err == nil {
				_, pgErr := pgDB.ExecContext(ctx, `
					INSERT INTO order_payments (id, order_id, payment_method, gateway, gateway_transaction_id, amount, status, snap_token, snap_redirect_url, raw_response, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
					ON CONFLICT (id) DO NOTHING
				`, id, orderID, payMethod, gateway, gwTxID, amount, status, snapToken, snapRedirect, rawResp, createdAt, updatedAt)

				if pgErr == nil {
					_, _ = r.localDB.ExecContext(ctx, "UPDATE order_payments SET sync_status = 'SYNCED' WHERE id = ?", id)
				}
			}
		}
	}

	return nil
}

var _ outbound.SyncRepository = (*SyncSQLiteRepository)(nil)
