package dto

import (
	"time"

	"github.com/pos-backend/internal/domain"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        string          `json:"id"`
	Username  string          `json:"username"`
	Role      domain.UserRole `json:"role"`
	FullName  string          `json:"full_name"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
