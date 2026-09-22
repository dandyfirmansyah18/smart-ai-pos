package outbound

import (
	"context"
	"database/sql"

	"github.com/pos-backend/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, tx *sql.Tx, key string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	ListActiveOrders(ctx context.Context) ([]domain.Order, error)
}
