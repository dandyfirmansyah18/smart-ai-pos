package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/domain"
)

type dummyAccessMenuUC struct {
	menus    []domain.AccessMenu
	mappings []domain.RoleAccessConfig
	err      error
}

func (d *dummyAccessMenuUC) GetMenus(ctx context.Context) ([]domain.AccessMenu, error) {
	return d.menus, d.err
}

func (d *dummyAccessMenuUC) GetRoleAccess(ctx context.Context) ([]domain.RoleAccessConfig, error) {
	return d.mappings, d.err
}

func (d *dummyAccessMenuUC) UpdateAccess(ctx context.Context, role, menuKey string, canAccess bool) error {
	return d.err
}

func TestAccessMenuHandler_GetAccessMenus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &dummyAccessMenuUC{
		menus: []domain.AccessMenu{{Key: "dashboard", Name: "Dashboard"}},
	}
	h := rest.NewAccessMenuHandler(uc)

	r := gin.New()
	r.GET("/api/access-menus", h.GetAccessMenus)

	req, _ := http.NewRequest(http.MethodGet, "/api/access-menus", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAccessMenuHandler_GetRoleAccessMappings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &dummyAccessMenuUC{
		mappings: []domain.RoleAccessConfig{{Role: "admin", MenuKey: "dashboard", CanAccess: true}},
	}
	h := rest.NewAccessMenuHandler(uc)

	r := gin.New()
	r.GET("/api/access-menus/roles", h.GetRoleAccessMappings)

	req, _ := http.NewRequest(http.MethodGet, "/api/access-menus/roles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
