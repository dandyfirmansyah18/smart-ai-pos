package inbound

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/outbound"
	"github.com/pos-backend/pkg/utils"
)

type AuthUseCaseImpl struct {
	userRepo  outbound.UserRepository
	jwtSecret string
}

func NewAuthUseCaseImpl(userRepo outbound.UserRepository, jwtSecret string) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthUseCaseImpl) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := utils.GenerateJWTToken(user.ID, user.Username, string(user.Role), s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	userResp := dto.UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Role:      user.Role,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return &dto.LoginResponse{
		Token: token,
		User:  userResp,
	}, nil
}

func (s *AuthUseCaseImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

var _ AuthUseCase = (*AuthUseCaseImpl)(nil)
