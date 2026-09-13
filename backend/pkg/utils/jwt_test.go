package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/pkg/utils"
)

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	secret := "super-secret-jwt-key-32-chars-minimum!"
	userID := uuid.New()
	username := "john_cashier"
	role := "CASHIER"

	// 1. Test Valid Token Generation & Validation
	tokenStr, err := utils.GenerateJWTToken(userID, username, role, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	claims, err := utils.ValidateJWTToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("expected token validation to succeed, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("expected Username %s, got %s", username, claims.Username)
	}
	if claims.Role != role {
		t.Errorf("expected Role %s, got %s", role, claims.Role)
	}

	// 2. Test Invalid Secret Key
	_, err = utils.ValidateJWTToken(tokenStr, "wrong-secret-key!")
	if err == nil {
		t.Errorf("expected validation failure with wrong secret key")
	}
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// 3. Test Expired Token
	expiredTokenStr, err := utils.GenerateJWTToken(userID, username, role, secret, -10*time.Second)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = utils.ValidateJWTToken(expiredTokenStr, secret)
	if err == nil {
		t.Errorf("expected validation failure for expired token")
	}
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}
