package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
)

func (s *Server) registerProductRoutes(api *gin.RouterGroup) {
	productHandler := NewProductHandler(s.productRepo)
	api.GET("/products", productHandler.ListProducts)
	api.GET("/products/:sku", productHandler.GetProductBySKU)

	if s.authUseCase != nil {
		productProtected := api.Group("/products")
		productProtected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
		productProtected.Use(middleware.RequireRole("ADMIN", "WAREHOUSE"))
		{
			productProtected.POST("", productHandler.CreateProduct)
		}
	} else {
		api.POST("/products", productHandler.CreateProduct)
	}
}
