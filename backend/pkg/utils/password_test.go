package utils_test

import (
	"testing"

	"github.com/pos-backend/pkg/utils"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassword123!"

	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == password {
		t.Errorf("expected hash to be different from original password")
	}

	if !utils.CheckPasswordHash(password, hash) {
		t.Errorf("expected valid password verification to return true")
	}

	if utils.CheckPasswordHash("WrongPassword", hash) {
		t.Errorf("expected invalid password verification to return false")
	}
}
