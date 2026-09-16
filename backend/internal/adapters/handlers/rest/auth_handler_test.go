package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
	"github.com/pos-backend/pkg/utils"
)

func TestAuthHandler_LoginAndMe(t *testing.T) {
	mockUserRepo := outbound.NewMockUserRepository()
	hashedPwd, _ := utils.HashPassword("secret123")
	userID := uuid.New()
	testUser := &domain.User{
		ID:           userID,
		Username:     "cashier1",
		PasswordHash: hashedPwd,
		Role:         domain.RoleCashier,
		FullName:     "John Cashier",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_ = mockUserRepo.Create(context.Background(), testUser)

	cfg := &config.Config{Port: "8080", Env: "test", JWTSecret: "test-secret"}
	authUseCase := inbound.NewAuthUseCaseImpl(mockUserRepo, cfg.JWTSecret)
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()
	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)

	server := rest.NewServer(cfg, mockProductRepo, orderUseCase, nil, authUseCase)

	// 1. Test Login Success
	loginReq := dto.LoginRequest{
		Username: "cashier1",
		Password: "secret123",
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d, body: %s", w.Code, w.Body.String())
	}

	var loginResp dto.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Errorf("expected JWT token in login response")
	}
	if loginResp.User.Username != "cashier1" {
		t.Errorf("expected username cashier1, got %s", loginResp.User.Username)
	}

	// 2. Test Login Failure (Invalid Password)
	badLoginReq := dto.LoginRequest{
		Username: "cashier1",
		Password: "wrongpassword",
	}
	badBody, _ := json.Marshal(badLoginReq)
	reqBad, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(badBody))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()

	server.Router().ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for bad password, got %d", wBad.Code)
	}

	// 3. Test GET /api/auth/me with Valid Token
	reqMe, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResp.Token)
	wMe := httptest.NewRecorder()

	server.Router().ServeHTTP(wMe, reqMe)
	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/auth/me, got %d, body: %s", wMe.Code, wMe.Body.String())
	}

	var meResp dto.UserResponse
	if err := json.Unmarshal(wMe.Body.Bytes(), &meResp); err != nil {
		t.Fatalf("failed to unmarshal me response: %v", err)
	}
	if meResp.Username != "cashier1" {
		t.Errorf("expected username cashier1 in /me, got %s", meResp.Username)
	}

	// 4. Test GET /api/auth/me without Token (Unauthorized)
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	wNoAuth := httptest.NewRecorder()
	server.Router().ServeHTTP(wNoAuth, reqNoAuth)
	if wNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got %d", wNoAuth.Code)
	}
}
