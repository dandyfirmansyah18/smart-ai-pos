package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
)

type AccessMenuHandler struct {
	db *sql.DB
}

func NewAccessMenuHandler(db *sql.DB) *AccessMenuHandler {
	return &AccessMenuHandler{db: db}
}

// GetAccessMenus GET /api/access-menus
func (h *AccessMenuHandler) GetAccessMenus(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(), "SELECT id, key, name, path, icon FROM access_menus ORDER BY name ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch access menus", "details": err.Error()})
		return
	}
	defer rows.Close()

	var menus []domain.AccessMenu
	for rows.Next() {
		var m domain.AccessMenu
		if err := rows.Scan(&m.ID, &m.Key, &m.Name, &m.Path, &m.Icon); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan access menu"})
			return
		}
		menus = append(menus, m)
	}

	c.JSON(http.StatusOK, menus)
}

// GetRoleAccessMappings GET /api/access-menus/roles
func (h *AccessMenuHandler) GetRoleAccessMappings(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(), "SELECT role, menu_key, can_access FROM access_menus_per_roles")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role access mappings", "details": err.Error()})
		return
	}
	defer rows.Close()

	var mappings []domain.RoleAccessConfig
	for rows.Next() {
		var r domain.RoleAccessConfig
		if err := rows.Scan(&r.Role, &r.MenuKey, &r.CanAccess); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan mapping"})
			return
		}
		mappings = append(mappings, r)
	}

	c.JSON(http.StatusOK, mappings)
}

// UpdateRoleAccess PUT /api/access-menus/roles (Admin only)
func (h *AccessMenuHandler) UpdateRoleAccess(c *gin.Context) {
	var req struct {
		Role      string `json:"role" binding:"required"`
		MenuKey   string `json:"menu_key" binding:"required"`
		CanAccess bool   `json:"can_access"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	query := `INSERT INTO access_menus_per_roles (role, menu_key, can_access) 
	          VALUES ($1, $2, $3) 
	          ON CONFLICT (role, menu_key) DO UPDATE SET can_access = EXCLUDED.can_access`

	_, err := h.db.ExecContext(c.Request.Context(), query, req.Role, req.MenuKey, req.CanAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role access", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role access updated successfully"})
}
