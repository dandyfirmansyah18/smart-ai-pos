package outbound

import (
	"context"
	"database/sql"

	"github.com/pos-backend/internal/domain"
)

type MockOrderRepository struct {
	Orders map[string]*domain.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		Orders: make(map[string]*domain.Order),
	}
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	if order.IdempotencyKey != "" {
		for _, o := range m.Orders {
			if o.IdempotencyKey == order.IdempotencyKey {
				return domain.ErrDuplicateIdempotencyKey
			}
		}
	}
	m.Orders[order.ID.String()] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	o, ok := m.Orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return o, nil
}

func (m *MockOrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	for _, o := range m.Orders {
		if o.IdempotencyKey == key {
			return o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	o, ok := m.Orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	o.Status = status
	return nil
}

// Compile-time check to ensure MockOrderRepository implements OrderRepository
var _ OrderRepository = (*MockOrderRepository)(nil)
