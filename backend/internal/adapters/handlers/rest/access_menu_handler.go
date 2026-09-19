package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/ports/inbound"
)

type AccessMenuHandler struct {
	useCase inbound.AccessMenuUseCase
}

func NewAccessMenuHandler(useCase inbound.AccessMenuUseCase) *AccessMenuHandler {
	return &AccessMenuHandler{useCase: useCase}
}

// GetAccessMenus GET /api/access-menus
func (h *AccessMenuHandler) GetAccessMenus(c *gin.Context) {
	menus, err := h.useCase.GetMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch access menus", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, menus)
}

// GetRoleAccessMappings GET /api/access-menus/roles
func (h *AccessMenuHandler) GetRoleAccessMappings(c *gin.Context) {
	mappings, err := h.useCase.GetRoleAccess(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role access mappings", "details": err.Error()})
		return
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

	err := h.useCase.UpdateAccess(c.Request.Context(), req.Role, req.MenuKey, req.CanAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role access", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role access updated successfully"})
}
