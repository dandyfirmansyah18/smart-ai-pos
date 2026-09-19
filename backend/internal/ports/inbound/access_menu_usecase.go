package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type AccessMenuUseCase interface {
	GetMenus(ctx context.Context) ([]domain.AccessMenu, error)
	GetRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error)
	UpdateAccess(ctx context.Context, role string, menuKey string, canAccess bool) error
}
