package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type AccessMenuSQLiteRepository struct {
	db *sql.DB
}

func NewAccessMenuSQLiteRepository(db *sql.DB) *AccessMenuSQLiteRepository {
	return &AccessMenuSQLiteRepository{db: db}
}

func (r *AccessMenuSQLiteRepository) ListMenus(ctx context.Context) ([]domain.AccessMenu, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, key, name, path, icon FROM access_menus ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch access menus: %w", err)
	}
	defer rows.Close()

	var menus []domain.AccessMenu
	for rows.Next() {
		var m domain.AccessMenu
		if err := rows.Scan(&m.ID, &m.Key, &m.Name, &m.Path, &m.Icon); err != nil {
			return nil, fmt.Errorf("failed to scan access menu: %w", err)
		}
		menus = append(menus, m)
	}

	if menus == nil {
		menus = []domain.AccessMenu{}
	}

	return menus, nil
}

func (r *AccessMenuSQLiteRepository) ListRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT role, menu_key, can_access FROM access_menus_per_roles")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role access mappings: %w", err)
	}
	defer rows.Close()

	var mappings []domain.RoleAccessConfig
	for rows.Next() {
		var item domain.RoleAccessConfig
		if err := rows.Scan(&item.Role, &item.MenuKey, &item.CanAccess); err != nil {
			return nil, fmt.Errorf("failed to scan mapping: %w", err)
		}
		mappings = append(mappings, item)
	}

	if mappings == nil {
		mappings = []domain.RoleAccessConfig{}
	}

	return mappings, nil
}

func (r *AccessMenuSQLiteRepository) UpdateRoleAccess(ctx context.Context, role string, menuKey string, canAccess bool) error {
	query := `INSERT INTO access_menus_per_roles (id, role, menu_key, can_access) 
	          VALUES (lower(hex(randomblob(16))), ?, ?, ?) 
	          ON CONFLICT(role, menu_key) DO UPDATE SET can_access = excluded.can_access`

	_, err := r.db.ExecContext(ctx, query, role, menuKey, canAccess)
	if err != nil {
		return fmt.Errorf("failed to update role access: %w", err)
	}

	return nil
}

var _ outbound.AccessMenuRepository = (*AccessMenuSQLiteRepository)(nil)
