package rest

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Smart AI POS API
// @version 1.0
// @description High-performance POS API with offline SQLite resilience, KDS, and Midtrans payment gateway.
// @host localhost:8080
// @BasePath /api/v1
func (s *Server) registerSwaggerRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
