package outbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
)

type AccessMenuRepository interface {
	ListMenus(ctx context.Context) ([]domain.AccessMenu, error)
	ListRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error)
	UpdateRoleAccess(ctx context.Context, role string, menuKey string, canAccess bool) error
}
