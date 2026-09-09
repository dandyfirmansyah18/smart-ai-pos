package outbound

import (
	"context"
	"database/sql"

	"github.com/pos-backend/internal/domain"
)

type MockProductRepository struct {
	Products map[string]*domain.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		Products: make(map[string]*domain.Product),
	}
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	p, ok := m.Products[sku]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (m *MockProductRepository) GetBySKUWithLock(ctx context.Context, tx *sql.Tx, sku string) (*domain.Product, error) {
	return m.GetBySKU(ctx, sku)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, tx *sql.Tx, sku string, newQty int) error {
	p, ok := m.Products[sku]
	if !ok {
		return domain.ErrProductNotFound
	}
	if newQty < 0 {
		return domain.ErrInsufficientStock
	}
	p.StockQuantity = newQty
	return nil
}

func (m *MockProductRepository) ListAll(ctx context.Context) ([]domain.Product, error) {
	result := make([]domain.Product, 0, len(m.Products))
	for _, p := range m.Products {
		result = append(result, *p)
	}
	return result, nil
}

// Compile-time check to ensure MockProductRepository implements ProductRepository
var _ ProductRepository = (*MockProductRepository)(nil)
