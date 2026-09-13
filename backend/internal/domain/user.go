package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUnauthorizedRole   = errors.New("unauthorized role access")
)

type UserRole string

const (
	RoleAdmin     UserRole = "ADMIN"
	RoleCashier   UserRole = "CASHIER"
	RoleKitchen   UserRole = "KITCHEN"
	RoleWarehouse UserRole = "WAREHOUSE"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
