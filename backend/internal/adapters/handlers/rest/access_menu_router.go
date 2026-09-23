package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
)

func (s *Server) registerAccessMenuRoutes(api *gin.RouterGroup) {
	if s.accessMenuUseCase == nil {
		return
	}
	accessHandler := NewAccessMenuHandler(s.accessMenuUseCase)
	api.GET("/access-menus", accessHandler.GetAccessMenus)
	api.GET("/access-menus/roles", accessHandler.GetRoleAccessMappings)

	roleProtected := api.Group("/access-menus")
	if s.authUseCase != nil {
		roleProtected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
		roleProtected.Use(middleware.RequireRole("ADMIN"))
	}
	{
		roleProtected.PUT("/roles", accessHandler.UpdateRoleAccess)
	}
}
