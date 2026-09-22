package resilient

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/adapters/infrastructure/sqlite"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

// Helper to check if an error is connection/network related
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "bad connection") ||
		strings.Contains(s, "dial tcp") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "closed") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "no connection") ||
		strings.Contains(s, "server closed") ||
		strings.Contains(s, "eof") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "unreachable") ||
		strings.Contains(s, "refused") ||
		strings.Contains(s, "timeout") ||
		strings.Contains(s, "cannot connect") ||
		strings.Contains(s, "database is closed") ||
		strings.Contains(s, "driver: bad connection")
}

// ResilientDBManager manages dual PostgreSQL and SQLite database connections with dynamic failover.
type ResilientDBManager struct {
	pgDB         *sql.DB
	sqliteDB     *sql.DB
	autoFailover bool

	mu              sync.RWMutex
	isPGHealthy     bool
	lastCheckTime   time.Time
	checkTTL        time.Duration
}

func NewResilientDBManager(pgDB *sql.DB, sqliteDB *sql.DB, autoFailover bool) *ResilientDBManager {
	mgr := &ResilientDBManager{
		pgDB:         pgDB,
		sqliteDB:     sqliteDB,
		autoFailover: autoFailover,
		checkTTL:     5 * time.Second,
	}
	// Initial health check
	mgr.IsPGHealthy(context.Background())
	return mgr
}

func (m *ResilientDBManager) IsPGHealthy(ctx context.Context) bool {
	if m == nil || m.pgDB == nil {
		return false
	}

	m.mu.RLock()
	if time.Since(m.lastCheckTime) < m.checkTTL {
		healthy := m.isPGHealthy
		m.mu.RUnlock()
		return healthy
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Re-check after acquiring write lock
	if time.Since(m.lastCheckTime) < m.checkTTL {
		return m.isPGHealthy
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := m.pgDB.PingContext(pingCtx)
	m.lastCheckTime = time.Now()
	if err != nil || IsConnectionError(err) {
		if m.isPGHealthy {
			log.Printf("[ResilientDBManager] Warning: PostgreSQL health check failed (%v). Failing over to SQLite.", err)
		}
		m.isPGHealthy = false
	} else {
		if !m.isPGHealthy && m.lastCheckTime.After(time.Time{}) {
			log.Printf("[ResilientDBManager] Info: PostgreSQL connection recovered.")
		}
		m.isPGHealthy = true
	}
	return m.isPGHealthy
}

func (m *ResilientDBManager) BeginTx(ctx context.Context) (*sql.Tx, bool, error) {
	if m.IsPGHealthy(ctx) {
		tx, err := m.pgDB.BeginTx(ctx, nil)
		if err == nil {
			return tx, false, nil
		}
		if IsConnectionError(err) {
			log.Printf("[ResilientDBManager] PostgreSQL BeginTx failed with connection error (%v). Falling back to SQLite.", err)
			m.mu.Lock()
			m.isPGHealthy = false
			m.mu.Unlock()
		} else {
			return nil, false, err
		}
	}

	if m.sqliteDB != nil {
		tx, err := m.sqliteDB.BeginTx(ctx, nil)
		if err != nil {
			return nil, true, err
		}
		return tx, true, nil
	}

	return nil, false, sql.ErrTxDone
}

func (m *ResilientDBManager) GetPGDB() *sql.DB {
	return m.pgDB
}

func (m *ResilientDBManager) GetSQLiteDB() *sql.DB {
	return m.sqliteDB
}

// ResilientOrderRepository
type ResilientOrderRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.OrderPGRepository
	sqliteRepo *sqlite.OrderSQLiteRepository
}

func NewResilientOrderRepository(mgr *ResilientDBManager) *ResilientOrderRepository {
	var pgRepo *pg.OrderPGRepository
	var sqliteRepo *sqlite.OrderSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewOrderPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewOrderSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientOrderRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientOrderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	if tx != nil {
		// Try using the provided transaction
		if r.pgRepo != nil {
			err := r.pgRepo.CreateOrder(ctx, tx, order)
			if err == nil {
				return nil
			}
			if !IsConnectionError(err) {
				return err
			}
		}
		if r.sqliteRepo != nil {
			return r.sqliteRepo.CreateOrder(ctx, tx, order)
		}
	}

	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.CreateOrder(ctx, nil, order)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CreateOrder(ctx, nil, order)
	}
	return sql.ErrConnDone
}

func (r *ResilientOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		o, err := r.pgRepo.GetByID(ctx, id)
		if err == nil {
			return o, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetByID(ctx, id)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientOrderRepository) GetByIdempotencyKey(ctx context.Context, tx *sql.Tx, key string) (*domain.Order, error) {
	if tx != nil {
		if r.pgRepo != nil {
			o, err := r.pgRepo.GetByIdempotencyKey(ctx, tx, key)
			if err == nil {
				return o, nil
			}
			if !IsConnectionError(err) {
				return nil, err
			}
		}
		if r.sqliteRepo != nil {
			return r.sqliteRepo.GetByIdempotencyKey(ctx, tx, key)
		}
	}

	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		o, err := r.pgRepo.GetByIdempotencyKey(ctx, nil, key)
		if err == nil {
			return o, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetByIdempotencyKey(ctx, nil, key)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.UpdateStatus(ctx, id, status)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			// If not found on PG, check SQLite as well
			if r.sqliteRepo != nil {
				if sqliteErr := r.sqliteRepo.UpdateStatus(ctx, id, status); sqliteErr == nil {
					return nil
				}
			}
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.UpdateStatus(ctx, id, status)
	}
	return sql.ErrConnDone
}

func (r *ResilientOrderRepository) ListActiveOrders(ctx context.Context) ([]domain.Order, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		orders, err := r.pgRepo.ListActiveOrders(ctx)
		if err == nil {
			// Merge active orders from SQLite if any exist
			if r.sqliteRepo != nil {
				sqliteOrders, _ := r.sqliteRepo.ListActiveOrders(ctx)
				if len(sqliteOrders) > 0 {
					seen := make(map[string]bool)
					for _, o := range orders {
						seen[o.ID.String()] = true
					}
					for _, sqo := range sqliteOrders {
						if !seen[sqo.ID.String()] {
							orders = append(orders, sqo)
						}
					}
				}
			}
			return orders, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListActiveOrders(ctx)
	}
	return nil, sql.ErrConnDone
}

var _ outbound.OrderRepository = (*ResilientOrderRepository)(nil)

// ResilientProductRepository
type ResilientProductRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.ProductPGRepository
	sqliteRepo *sqlite.ProductSQLiteRepository
}

func NewResilientProductRepository(mgr *ResilientDBManager) *ResilientProductRepository {
	var pgRepo *pg.ProductPGRepository
	var sqliteRepo *sqlite.ProductSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewProductPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewProductSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientProductRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		p, err := r.pgRepo.GetBySKU(ctx, sku)
		if err == nil {
			return p, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetBySKU(ctx, sku)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientProductRepository) GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error) {
	if tx != nil {
		if r.pgRepo != nil {
			p, err := r.pgRepo.GetBySKUWithLock(ctx, tx, sku)
			if err == nil {
				return p, nil
			}
			if !IsConnectionError(err) {
				return nil, err
			}
		}
		if r.sqliteRepo != nil {
			return r.sqliteRepo.GetBySKUWithLock(ctx, tx, sku)
		}
	}

	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		p, err := r.pgRepo.GetBySKUWithLock(ctx, nil, sku)
		if err == nil {
			return p, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetBySKUWithLock(ctx, nil, sku)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientProductRepository) UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error {
	if tx != nil {
		if r.pgRepo != nil {
			err := r.pgRepo.UpdateStock(ctx, tx, sku, newQty)
			if err == nil {
				return nil
			}
			if !IsConnectionError(err) {
				return err
			}
		}
		if r.sqliteRepo != nil {
			return r.sqliteRepo.UpdateStock(ctx, tx, sku, newQty)
		}
	}

	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.UpdateStock(ctx, nil, sku, newQty)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.UpdateStock(ctx, nil, sku, newQty)
	}
	return sql.ErrConnDone
}

func (r *ResilientProductRepository) ListAll(ctx context.Context) ([]domain.Product, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		products, err := r.pgRepo.ListAll(ctx)
		if err == nil {
			return products, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListAll(ctx)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientProductRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.CreateProduct(ctx, product)
		if err == nil {
			if r.sqliteRepo != nil {
				_ = r.sqliteRepo.CreateProduct(ctx, product)
			}
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CreateProduct(ctx, product)
	}
	return sql.ErrConnDone
}

var _ outbound.ProductRepository = (*ResilientProductRepository)(nil)

// ResilientCashShiftRepository
type ResilientCashShiftRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.CashShiftPGRepository
	sqliteRepo *sqlite.CashShiftSQLiteRepository
}

func NewResilientCashShiftRepository(mgr *ResilientDBManager) *ResilientCashShiftRepository {
	var pgRepo *pg.CashShiftPGRepository
	var sqliteRepo *sqlite.CashShiftSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewCashShiftPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewCashShiftSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientCashShiftRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientCashShiftRepository) OpenShift(ctx context.Context, shift *domain.CashShift) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.OpenShift(ctx, shift)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.OpenShift(ctx, shift)
	}
	return sql.ErrConnDone
}

func (r *ResilientCashShiftRepository) GetCurrentShift(ctx context.Context, userID string) (*domain.CashShift, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		shift, err := r.pgRepo.GetCurrentShift(ctx, userID)
		if err == nil {
			return shift, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetCurrentShift(ctx, userID)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientCashShiftRepository) CloseShift(ctx context.Context, shiftID string, closingCash float64, expectedCash float64, totalSales float64, notes string) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.CloseShift(ctx, shiftID, closingCash, expectedCash, totalSales, notes)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			if r.sqliteRepo != nil {
				if sqliteErr := r.sqliteRepo.CloseShift(ctx, shiftID, closingCash, expectedCash, totalSales, notes); sqliteErr == nil {
					return nil
				}
			}
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CloseShift(ctx, shiftID, closingCash, expectedCash, totalSales, notes)
	}
	return sql.ErrConnDone
}

func (r *ResilientCashShiftRepository) CalculateTotalSales(ctx context.Context, openedAt time.Time) (float64, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		sales, err := r.pgRepo.CalculateTotalSales(ctx, openedAt)
		if err == nil {
			return sales, nil
		}
		if !IsConnectionError(err) {
			return 0, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CalculateTotalSales(ctx, openedAt)
	}
	return 0, sql.ErrConnDone
}

var _ outbound.CashShiftRepository = (*ResilientCashShiftRepository)(nil)

// ResilientPaymentRepository
type ResilientPaymentRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.PaymentPGRepository
	sqliteRepo *sqlite.PaymentSQLiteRepository
}

func NewResilientPaymentRepository(mgr *ResilientDBManager) *ResilientPaymentRepository {
	var pgRepo *pg.PaymentPGRepository
	var sqliteRepo *sqlite.PaymentSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewPaymentPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewPaymentSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientPaymentRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientPaymentRepository) CreatePayment(ctx context.Context, payment *domain.OrderPayment) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.CreatePayment(ctx, payment)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CreatePayment(ctx, payment)
	}
	return sql.ErrConnDone
}

func (r *ResilientPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.OrderPayment, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		p, err := r.pgRepo.GetByOrderID(ctx, orderID)
		if err == nil {
			return p, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetByOrderID(ctx, orderID)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientPaymentRepository) GetByGatewayTxID(ctx context.Context, txID string) (*domain.OrderPayment, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		p, err := r.pgRepo.GetByGatewayTxID(ctx, txID)
		if err == nil {
			return p, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.GetByGatewayTxID(ctx, txID)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientPaymentRepository) UpdateStatus(ctx context.Context, gatewayTxID string, status domain.PaymentStatus, rawResponse string) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.UpdateStatus(ctx, gatewayTxID, status, rawResponse)
		if err == nil {
			return nil
		}
		if !IsConnectionError(err) {
			if r.sqliteRepo != nil {
				if sqliteErr := r.sqliteRepo.UpdateStatus(ctx, gatewayTxID, status, rawResponse); sqliteErr == nil {
					return nil
				}
			}
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.UpdateStatus(ctx, gatewayTxID, status, rawResponse)
	}
	return sql.ErrConnDone
}

var _ outbound.PaymentRepository = (*ResilientPaymentRepository)(nil)

// ResilientFinanceRepository
type ResilientFinanceRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.FinancePGRepository
	sqliteRepo *sqlite.FinanceSQLiteRepository
}

func NewResilientFinanceRepository(mgr *ResilientDBManager) *ResilientFinanceRepository {
	var pgRepo *pg.FinancePGRepository
	var sqliteRepo *sqlite.FinanceSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewFinancePGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewFinanceSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientFinanceRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientFinanceRepository) ListOrderHistory(ctx context.Context) ([]domain.Order, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		orders, err := r.pgRepo.ListOrderHistory(ctx)
		if err == nil {
			return orders, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListOrderHistory(ctx)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientFinanceRepository) CalculateProfitLoss(ctx context.Context) (totalRevenue, totalCOGS, grossProfit, margin float64, err error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		rev, cogs, profit, m, err := r.pgRepo.CalculateProfitLoss(ctx)
		if err == nil {
			return rev, cogs, profit, m, nil
		}
		if !IsConnectionError(err) {
			return 0, 0, 0, 0, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.CalculateProfitLoss(ctx)
	}
	return 0, 0, 0, 0, sql.ErrConnDone
}

var _ outbound.FinanceRepository = (*ResilientFinanceRepository)(nil)

// ResilientUserRepository
type ResilientUserRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.UserPGRepository
	sqliteRepo *sqlite.UserSQLiteRepository
}

func NewResilientUserRepository(mgr *ResilientDBManager) *ResilientUserRepository {
	var pgRepo *pg.UserPGRepository
	var sqliteRepo *sqlite.UserSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewUserPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewUserSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientUserRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		u, err := r.pgRepo.FindByUsername(ctx, username)
		if err == nil {
			return u, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.FindByUsername(ctx, username)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		u, err := r.pgRepo.FindByID(ctx, id)
		if err == nil {
			return u, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.FindByID(ctx, id)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientUserRepository) Create(ctx context.Context, user *domain.User) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.Create(ctx, user)
		if err == nil {
			if r.sqliteRepo != nil {
				_ = r.sqliteRepo.Create(ctx, user)
			}
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.Create(ctx, user)
	}
	return sql.ErrConnDone
}

var _ outbound.UserRepository = (*ResilientUserRepository)(nil)

// ResilientAccessMenuRepository
type ResilientAccessMenuRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.AccessMenuPGRepository
	sqliteRepo *sqlite.AccessMenuSQLiteRepository
}

func NewResilientAccessMenuRepository(mgr *ResilientDBManager) *ResilientAccessMenuRepository {
	var pgRepo *pg.AccessMenuPGRepository
	var sqliteRepo *sqlite.AccessMenuSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewAccessMenuPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewAccessMenuSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientAccessMenuRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientAccessMenuRepository) ListMenus(ctx context.Context) ([]domain.AccessMenu, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		menus, err := r.pgRepo.ListMenus(ctx)
		if err == nil {
			return menus, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListMenus(ctx)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientAccessMenuRepository) ListRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		configs, err := r.pgRepo.ListRoleAccess(ctx)
		if err == nil {
			return configs, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListRoleAccess(ctx)
	}
	return nil, sql.ErrConnDone
}

func (r *ResilientAccessMenuRepository) UpdateRoleAccess(ctx context.Context, role string, menuKey string, canAccess bool) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.UpdateRoleAccess(ctx, role, menuKey, canAccess)
		if err == nil {
			if r.sqliteRepo != nil {
				_ = r.sqliteRepo.UpdateRoleAccess(ctx, role, menuKey, canAccess)
			}
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.UpdateRoleAccess(ctx, role, menuKey, canAccess)
	}
	return sql.ErrConnDone
}

var _ outbound.AccessMenuRepository = (*ResilientAccessMenuRepository)(nil)

// ResilientReceiptRepository
type ResilientReceiptRepository struct {
	mgr        *ResilientDBManager
	pgRepo     *pg.ReceiptPGRepository
	sqliteRepo *sqlite.ReceiptSQLiteRepository
}

func NewResilientReceiptRepository(mgr *ResilientDBManager) *ResilientReceiptRepository {
	var pgRepo *pg.ReceiptPGRepository
	var sqliteRepo *sqlite.ReceiptSQLiteRepository
	if mgr.GetPGDB() != nil {
		pgRepo = pg.NewReceiptPGRepository(mgr.GetPGDB())
	}
	if mgr.GetSQLiteDB() != nil {
		sqliteRepo = sqlite.NewReceiptSQLiteRepository(mgr.GetSQLiteDB())
	}
	return &ResilientReceiptRepository{
		mgr:        mgr,
		pgRepo:     pgRepo,
		sqliteRepo: sqliteRepo,
	}
}

func (r *ResilientReceiptRepository) SaveAudit(ctx context.Context, audit *domain.ReceiptAudit) error {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		err := r.pgRepo.SaveAudit(ctx, audit)
		if err == nil {
			if r.sqliteRepo != nil {
				_ = r.sqliteRepo.SaveAudit(ctx, audit)
			}
			return nil
		}
		if !IsConnectionError(err) {
			return err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.SaveAudit(ctx, audit)
	}
	return sql.ErrConnDone
}

func (r *ResilientReceiptRepository) ListAudits(ctx context.Context) ([]domain.ReceiptAudit, error) {
	if r.mgr.IsPGHealthy(ctx) && r.pgRepo != nil {
		audits, err := r.pgRepo.ListAudits(ctx)
		if err == nil {
			return audits, nil
		}
		if !IsConnectionError(err) {
			return nil, err
		}
	}
	if r.sqliteRepo != nil {
		return r.sqliteRepo.ListAudits(ctx)
	}
	return nil, sql.ErrConnDone
}

var _ outbound.ReceiptRepository = (*ResilientReceiptRepository)(nil)
