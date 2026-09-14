package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
)

type AuthUseCase interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
