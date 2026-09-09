package outbound

import (
	"context"
	"database/sql"

	"github.com/pos-backend/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
