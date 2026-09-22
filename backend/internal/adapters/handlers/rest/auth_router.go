package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
)

func (s *Server) registerAuthRoutes(api *gin.RouterGroup) {
	if s.authUseCase == nil {
		return
	}
	authHandler := NewAuthHandler(s.authUseCase)
	api.POST("/auth/login", authHandler.Login)

	authProtected := api.Group("")
	authProtected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
	{
		authProtected.GET("/auth/me", authHandler.Me)
	}
}
