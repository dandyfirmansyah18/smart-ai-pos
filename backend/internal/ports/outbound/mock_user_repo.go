package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
)

type MockUserRepository struct {
	Users map[string]*domain.User
	IDs   map[uuid.UUID]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*domain.User),
		IDs:   make(map[uuid.UUID]*domain.User),
	}
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, ok := m.Users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.IDs[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserRepository) Create(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	m.Users[u.Username] = u
	m.IDs[u.ID] = u
	return nil
}

var _ UserRepository = (*MockUserRepository)(nil)
