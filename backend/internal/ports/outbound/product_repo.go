package outbound

import (
	"context"
	"database/sql"

	"github.com/pos-backend/internal/domain"
)

type ProductRepository interface {
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error)
	UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error
	ListAll(ctx context.Context) ([]domain.Product, error)
}
