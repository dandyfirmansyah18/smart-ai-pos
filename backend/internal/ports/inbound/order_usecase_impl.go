package inbound

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/adapters/infrastructure/sqlite"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/outbound"
)

type OrderUseCaseImpl struct {
	db          *sql.DB
	productRepo outbound.ProductRepository
	orderRepo   outbound.OrderRepository
	lockService outbound.LockService
	broadcaster outbound.EventBroadcaster
}

func NewOrderUseCaseImpl(
	db *sql.DB,
	productRepo outbound.ProductRepository,
	orderRepo outbound.OrderRepository,
	lockService outbound.LockService,
	broadcaster ...outbound.EventBroadcaster,
) *OrderUseCaseImpl {
	var b outbound.EventBroadcaster
	if len(broadcaster) > 0 {
		b = broadcaster[0]
	}
	return &OrderUseCaseImpl{
		db:          db,
		productRepo: productRepo,
		orderRepo:   orderRepo,
		lockService: lockService,
		broadcaster: b,
	}
}

func isConnectionError(err error) bool {
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

func (s *OrderUseCaseImpl) Checkout(ctx context.Context, req dto.CheckoutRequest) (*domain.Order, error) {
	if req.IdempotencyKey == "" {
		return nil, domain.ErrInvalidOrder
	}
	if len(req.Items) == 0 {
		return nil, domain.ErrInvalidOrder
	}

	// 1. Acquire Redis lock for idempotency key (if Redis is available)
	lockKey := "lock:checkout:" + req.IdempotencyKey
	if s.lockService != nil {
		acquired, err := s.lockService.AcquireLock(ctx, lockKey, 15*time.Second)
		if err != nil {
			// If Redis is unreachable in offline mode, continue without Redis lock
			if !isConnectionError(err) {
				return nil, fmt.Errorf("failed to acquire checkout lock: %w", err)
			}
		} else if !acquired {
			return nil, domain.ErrDuplicateIdempotencyKey
		} else {
			defer s.lockService.ReleaseLock(ctx, lockKey)
		}
	}

	// 2. Database Connection Check & Transaction Initialization
	var tx *sql.Tx
	activeDB := s.db
	activeProductRepo := s.productRepo
	activeOrderRepo := s.orderRepo

	useSQLite := false
	if activeDB != nil {
		pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
		pingErr := activeDB.PingContext(pingCtx)
		pingCancel()

		if pingErr != nil || isConnectionError(pingErr) {
			useSQLite = true
		} else {
			var txErr error
			tx, txErr = activeDB.BeginTx(ctx, nil)
			if txErr != nil || isConnectionError(txErr) {
				useSQLite = true
			}
		}
	}

	var localDB *sql.DB
	if useSQLite {
		var sqliteErr error
		localDB, sqliteErr = sqlite.NewSQLiteDB("./pos_local.db")
		if sqliteErr != nil || localDB == nil {
			return nil, fmt.Errorf("failed to open local sqlite database fallback: %w", sqliteErr)
		}
		defer localDB.Close()
		activeDB = localDB
		activeProductRepo = sqlite.NewProductSQLiteRepository(localDB)
		activeOrderRepo = sqlite.NewOrderSQLiteRepository(localDB)

		var txErr error
		tx, txErr = activeDB.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, fmt.Errorf("failed to begin sqlite transaction: %w", txErr)
		}
	}

	if tx != nil {
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	// 3. Check if order with idempotency key already exists in DB
	existingOrder, err := activeOrderRepo.GetByIdempotencyKey(ctx, tx, req.IdempotencyKey)
	if err == nil && existingOrder != nil {
		if tx != nil {
			_ = tx.Commit()
			tx = nil
		}
		return existingOrder, nil
	}

	// 4. Process each item: row-level pessimistic locking & stock decrement
	var totalAmount float64
	orderItems := make([]domain.OrderItem, 0, len(req.Items))
	stockUpdates := make([]dto.StockUpdateInfo, 0, len(req.Items))

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, domain.ErrInvalidOrder
		}

		product, err := activeProductRepo.GetBySKUWithLock(ctx, tx, item.SKU)
		if err != nil {
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("%w: SKU %s has stock %d, requested %d",
				domain.ErrInsufficientStock, item.SKU, product.StockQuantity, item.Quantity)
		}

		newStock := product.StockQuantity - item.Quantity
		if err := activeProductRepo.UpdateStock(ctx, tx, item.SKU, newStock); err != nil {
			return nil, err
		}

		stockUpdates = append(stockUpdates, dto.StockUpdateInfo{SKU: item.SKU, NewStock: newStock})

		itemTotal := float64(item.Quantity) * product.Price
		totalAmount += itemTotal

		orderItems = append(orderItems, domain.OrderItem{
			ID:        uuid.New(),
			ProductID: product.ID,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		})
	}

	// 5. Construct Order entity
	now := time.Now()
	orderID := uuid.New()
	for i := range orderItems {
		orderItems[i].OrderID = orderID
	}

	initialStatus := domain.OrderStatusUnpaid
	if req.PaymentMethod == domain.PaymentMethodCash {
		initialStatus = domain.OrderStatusPending
	}

	order := &domain.Order{
		ID:             orderID,
		TransactionID:  "TXN-" + uuid.New().String(),
		TotalAmount:    totalAmount,
		Status:         initialStatus,
		IdempotencyKey: req.IdempotencyKey,
		Items:          orderItems,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 6. Save Order & OrderItems
	if err := activeOrderRepo.CreateOrder(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 7. Commit Transaction
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		tx = nil
	}

	// 8. Broadcast real-time stock update & order status events to connected WebSocket clients
	if s.broadcaster != nil {
		for _, update := range stockUpdates {
			s.broadcaster.BroadcastStockUpdate(update.SKU, update.NewStock)
		}
		if initialStatus != domain.OrderStatusUnpaid {
			s.broadcaster.BroadcastOrderStatusUpdate(orderID.String(), initialStatus)
		}
	}

	return order, nil
}

// Compile-time check
var _ OrderUseCase = (*OrderUseCaseImpl)(nil)
