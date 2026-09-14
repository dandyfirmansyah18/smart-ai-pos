package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type UserPGRepository struct {
	db *sql.DB
}

func NewUserPGRepository(db *sql.DB) *UserPGRepository {
	return &UserPGRepository{db: db}
}

func (r *UserPGRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, full_name, created_at, updated_at FROM users WHERE username = $1`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID,
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
	return &u, nil
}

func (r *UserPGRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, full_name, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
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
	return &u, nil
}

func (r *UserPGRepository) Create(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	query := `INSERT INTO users (id, username, password_hash, role, full_name, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.PasswordHash, u.Role, u.FullName)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

var _ outbound.UserRepository = (*UserPGRepository)(nil)
