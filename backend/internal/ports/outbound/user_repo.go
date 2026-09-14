package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}
