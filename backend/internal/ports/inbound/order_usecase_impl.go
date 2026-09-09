package inbound

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
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

type updatedStockInfo struct {
	sku      string
	newStock int
}

func (s *OrderUseCaseImpl) Checkout(ctx context.Context, req CheckoutRequest) (*domain.Order, error) {
	if req.IdempotencyKey == "" {
		return nil, domain.ErrInvalidOrder
	}
	if len(req.Items) == 0 {
		return nil, domain.ErrInvalidOrder
	}

	// 1. Acquire Redis lock for idempotency key
	lockKey := "lock:checkout:" + req.IdempotencyKey
	acquired, err := s.lockService.AcquireLock(ctx, lockKey, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire checkout lock: %w", err)
	}
	if !acquired {
		return nil, domain.ErrDuplicateIdempotencyKey
	}
	defer s.lockService.ReleaseLock(ctx, lockKey)

	// 2. Start PostgreSQL Transaction (if DB is provided)
	var tx *sql.Tx
	if s.db != nil {
		var txErr error
		tx, txErr = s.db.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", txErr)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	// 3. Check if order with idempotency key already exists in DB
	existingOrder, err := s.orderRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
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
	stockUpdates := make([]updatedStockInfo, 0, len(req.Items))

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, domain.ErrInvalidOrder
		}

		product, err := s.productRepo.GetBySKUWithLock(ctx, tx, item.SKU)
		if err != nil {
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("%w: SKU %s has stock %d, requested %d",
				domain.ErrInsufficientStock, item.SKU, product.StockQuantity, item.Quantity)
		}

		newStock := product.StockQuantity - item.Quantity
		if err := s.productRepo.UpdateStock(ctx, tx, item.SKU, newStock); err != nil {
			return nil, err
		}

		stockUpdates = append(stockUpdates, updatedStockInfo{sku: item.SKU, newStock: newStock})

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

	order := &domain.Order{
		ID:             orderID,
		TransactionID:  "TXN-" + uuid.New().String(),
		TotalAmount:    totalAmount,
		Status:         domain.StatusCompleted,
		IdempotencyKey: req.IdempotencyKey,
		Items:          orderItems,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 6. Save Order & OrderItems
	if err := s.orderRepo.CreateOrder(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 7. Commit Transaction
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		tx = nil
	}

	// 8. Broadcast real-time stock update events to connected WebSocket clients
	if s.broadcaster != nil {
		for _, update := range stockUpdates {
			s.broadcaster.BroadcastStockUpdate(update.sku, update.newStock)
		}
	}

	return order, nil
}

// Compile-time check
var _ OrderUseCase = (*OrderUseCaseImpl)(nil)
