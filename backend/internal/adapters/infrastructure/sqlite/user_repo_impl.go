package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type UserSQLiteRepository struct {
	db *sql.DB
}

func NewUserSQLiteRepository(db *sql.DB) *UserSQLiteRepository {
	return &UserSQLiteRepository{db: db}
}

func (r *UserSQLiteRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, full_name, created_at, updated_at FROM users WHERE username = ?`
	var u domain.User
	var idStr string
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&idStr,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.FullName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	u.ID, _ = uuid.Parse(idStr)
	return &u, nil
}

func (r *UserSQLiteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, full_name, created_at, updated_at FROM users WHERE id = ?`
	var u domain.User
	var idStr string
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.FullName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	u.ID, _ = uuid.Parse(idStr)
	return &u, nil
}

func (r *UserSQLiteRepository) Create(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	query := `INSERT INTO users (id, username, password_hash, role, full_name, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query, u.ID.String(), u.Username, u.PasswordHash, u.Role, u.FullName)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

var _ outbound.UserRepository = (*UserSQLiteRepository)(nil)
