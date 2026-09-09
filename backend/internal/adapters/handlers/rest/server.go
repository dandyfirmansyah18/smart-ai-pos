package rest

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

type Server struct {
	router       *gin.Engine
	cfg          *config.Config
	productRepo  outbound.ProductRepository
	orderUseCase inbound.OrderUseCase
}

func NewServer(
	cfg *config.Config,
	productRepo outbound.ProductRepository,
	orderUseCase inbound.OrderUseCase,
) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{
		router:       r,
		cfg:          cfg,
		productRepo:  productRepo,
		orderUseCase: orderUseCase,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	api := s.router.Group("/api")
	{
		productHandler := NewProductHandler(s.productRepo)
		api.GET("/products", productHandler.ListProducts)
		api.GET("/products/:sku", productHandler.GetProductBySKU)

		orderHandler := NewOrderHandler(s.orderUseCase)
		api.POST("/orders/checkout", orderHandler.Checkout)
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() error {
	addr := ":" + s.cfg.Port
	return s.router.Run(addr)
}
