package inbound

import (
	"context"

	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type AccessMenuUseCaseImpl struct {
	repo outbound.AccessMenuRepository
}

func NewAccessMenuUseCaseImpl(repo outbound.AccessMenuRepository) *AccessMenuUseCaseImpl {
	return &AccessMenuUseCaseImpl{repo: repo}
}

func (u *AccessMenuUseCaseImpl) GetMenus(ctx context.Context) ([]domain.AccessMenu, error) {
	return u.repo.ListMenus(ctx)
}

func (u *AccessMenuUseCaseImpl) GetRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error) {
	return u.repo.ListRoleAccess(ctx)
}

func (u *AccessMenuUseCaseImpl) UpdateAccess(ctx context.Context, role string, menuKey string, canAccess bool) error {
	return u.repo.UpdateRoleAccess(ctx, role, menuKey, canAccess)
}

var _ AccessMenuUseCase = (*AccessMenuUseCaseImpl)(nil)
