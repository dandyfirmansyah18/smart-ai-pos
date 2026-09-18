package rest

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest/middleware"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

type Server struct {
	router         *gin.Engine
	cfg            *config.Config
	productRepo    outbound.ProductRepository
	orderRepo      outbound.OrderRepository
	orderUseCase   inbound.OrderUseCase
	hub            *ws.Hub
	receiptUseCase inbound.ReceiptUseCase
	authUseCase    inbound.AuthUseCase
	db             *sql.DB
}

func NewServer(
	cfg *config.Config,
	productRepo outbound.ProductRepository,
	orderUseCase inbound.OrderUseCase,
	hub *ws.Hub,
	optionalDeps ...interface{},
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

	var ru inbound.ReceiptUseCase
	var au inbound.AuthUseCase
	var ordRepo outbound.OrderRepository
	var database *sql.DB

	for _, dep := range optionalDeps {
		switch d := dep.(type) {
		case inbound.ReceiptUseCase:
			ru = d
		case inbound.AuthUseCase:
			au = d
		case outbound.OrderRepository:
			ordRepo = d
		case *sql.DB:
			database = d
		}
	}

	s := &Server{
		router:         r,
		cfg:            cfg,
		productRepo:    productRepo,
		orderRepo:      ordRepo,
		orderUseCase:   orderUseCase,
		hub:            hub,
		receiptUseCase: ru,
		authUseCase:    au,
		db:             database,
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

	if s.hub != nil {
		s.router.GET("/ws", s.hub.ServeWS)
	}

	api := s.router.Group("/api")
	{
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

		orderHandler := NewOrderHandler(s.orderUseCase)
		api.POST("/orders/checkout", orderHandler.Checkout)

		if s.receiptUseCase != nil {
			receiptHandler := NewReceiptHandler(s.receiptUseCase)
			api.POST("/receipts/scan", receiptHandler.ScanReceipt)
			api.GET("/receipts/audits", receiptHandler.ListAudits)
		}

		if s.orderRepo != nil {
			kitchenHandler := NewKitchenHandler(s.orderRepo, s.hub)
			kitchenGroup := api.Group("/kitchen")
			if s.authUseCase != nil {
				kitchenGroup.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
				kitchenGroup.Use(middleware.RequireRole("ADMIN", "KITCHEN", "CASHIER"))
			}
			{
				kitchenGroup.GET("/orders", kitchenHandler.ListActiveOrders)
				kitchenGroup.PATCH("/orders/:id/status", kitchenHandler.UpdateOrderStatus)
			}
		}

		if s.db != nil {
			accessHandler := NewAccessMenuHandler(s.db)
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

		if s.authUseCase != nil {
			authHandler := NewAuthHandler(s.authUseCase)
			api.POST("/auth/login", authHandler.Login)

			authProtected := api.Group("")
			authProtected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
			{
				authProtected.GET("/auth/me", authHandler.Me)
			}
		}
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() error {
	addr := ":" + s.cfg.Port
	return s.router.Run(addr)
}
