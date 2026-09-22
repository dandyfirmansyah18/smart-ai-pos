package rest

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

type Server struct {
	router            *gin.Engine
	cfg               *config.Config
	productRepo       outbound.ProductRepository
	orderRepo         outbound.OrderRepository
	orderUseCase      inbound.OrderUseCase
	hub               *ws.Hub
	receiptUseCase    inbound.ReceiptUseCase
	authUseCase       inbound.AuthUseCase
	cashShiftUseCase  inbound.CashShiftUseCase
	financeUseCase    inbound.FinanceUseCase
	accessMenuUseCase inbound.AccessMenuUseCase
	paymentUseCase    inbound.PaymentUseCase
	syncUseCase       inbound.SyncUseCase
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
	var csUcase inbound.CashShiftUseCase
	var finUcase inbound.FinanceUseCase
	var amUcase inbound.AccessMenuUseCase
	var payUcase inbound.PaymentUseCase
	var syncUcase inbound.SyncUseCase

	for _, dep := range optionalDeps {
		switch d := dep.(type) {
		case inbound.ReceiptUseCase:
			ru = d
		case inbound.AuthUseCase:
			au = d
		case outbound.OrderRepository:
			ordRepo = d
		case inbound.CashShiftUseCase:
			csUcase = d
		case inbound.FinanceUseCase:
			finUcase = d
		case inbound.AccessMenuUseCase:
			amUcase = d
		case inbound.PaymentUseCase:
			payUcase = d
		case inbound.SyncUseCase:
			syncUcase = d
		}
	}

	s := &Server{
		router:            r,
		cfg:               cfg,
		productRepo:       productRepo,
		orderRepo:         ordRepo,
		orderUseCase:      orderUseCase,
		hub:               hub,
		receiptUseCase:    ru,
		authUseCase:       au,
		cashShiftUseCase:  csUcase,
		financeUseCase:    finUcase,
		accessMenuUseCase: amUcase,
		paymentUseCase:    payUcase,
		syncUseCase:       syncUcase,
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

	s.registerSwaggerRoutes()

	api := s.router.Group("/api")
	{
		s.registerProductRoutes(api)
		s.registerOrderRoutes(api)
		s.registerReceiptRoutes(api)
		s.registerKitchenRoutes(api)
		s.registerAccessMenuRoutes(api)
		s.registerCashShiftRoutes(api)
		s.registerFinanceRoutes(api)
		s.registerPaymentRoutes(api)
		s.registerSyncRoutes(api)
		s.registerAuthRoutes(api)
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() error {
	addr := ":" + s.cfg.Port
	return s.router.Run(addr)
}
