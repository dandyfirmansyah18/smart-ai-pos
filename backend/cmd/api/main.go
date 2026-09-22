package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/adapters/infrastructure/payment"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/adapters/infrastructure/redis"
	"github.com/pos-backend/internal/adapters/infrastructure/resilient"
	"github.com/pos-backend/internal/adapters/infrastructure/sqlite"
	"github.com/pos-backend/internal/adapters/infrastructure/vision"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/pkg/logger"
)

func main() {
	// 1. Load Viper Configuration & Initialize Logger
	cfg := config.Load()
	logger.Init(cfg.LogLevel)
	logger.Log.Info().Str("port", cfg.Port).Str("env", cfg.Env).Msg("Starting Smart AI POS Engine Server...")

	// 2. Initialize WebSocket Broadcaster Hub & start background event loop
	wsHub := ws.NewHub()
	go wsHub.Run()

	var pgDB *sql.DB
	var sqliteDB *sql.DB
	var err error

	// Always initialize local SQLite database for offline resilience
	sqliteDB, err = sqlite.NewSQLiteDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite database fallback at %s: %v", cfg.SQLitePath, err)
	}
	defer sqliteDB.Close()

	if cfg.DBType != "sqlite" {
		pgDB, err = pg.NewPostgresDB(cfg)
		if err != nil {
			log.Printf("Notice: PostgreSQL initialization error (%v). Auto-failover active.", err)
		} else {
			pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if pingErr := pgDB.PingContext(pingCtx); pingErr != nil {
				log.Printf("Notice: PostgreSQL ping failed (%v). Auto-failover active using SQLite at %s", pingErr, cfg.SQLitePath)
			} else {
				log.Printf("Successfully connected to central PostgreSQL database at %s:%s.", cfg.DBHost, cfg.DBPort)
			}
			cancel()
			defer pgDB.Close()
		}
	} else {
		log.Printf("Running in LOCAL OFFLINE mode using SQLite at %s", cfg.SQLitePath)
	}

	resilientMgr := resilient.NewResilientDBManager(pgDB, sqliteDB, cfg.AutoFailover)

	// 4. Initialize Redis client & Lock Service
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Notice: Redis connection error (%v). Make sure Redis is running on %s.", err, cfg.RedisAddr)
	}
	lockService := redis.NewRedisLockService(redisClient)

	// 5. Initialize Driven Adapters (Resilient Repositories & Clients)
	productRepo := resilient.NewResilientProductRepository(resilientMgr)
	orderRepo := resilient.NewResilientOrderRepository(resilientMgr)
	receiptRepo := resilient.NewResilientReceiptRepository(resilientMgr)
	userRepo := resilient.NewResilientUserRepository(resilientMgr)
	cashShiftRepo := resilient.NewResilientCashShiftRepository(resilientMgr)
	financeRepo := resilient.NewResilientFinanceRepository(resilientMgr)
	accessMenuRepo := resilient.NewResilientAccessMenuRepository(resilientMgr)
	paymentRepo := resilient.NewResilientPaymentRepository(resilientMgr)
	syncRepo := sqlite.NewSyncSQLiteRepository(sqliteDB)

	visionClient := vision.NewVisionClientImpl(cfg)
	midtransClient := payment.NewMidtransClient(cfg)

	// 6. Initialize Driver Use Cases
	orderUseCase := inbound.NewOrderUseCaseImpl(pgDB, productRepo, orderRepo, lockService, wsHub)
	receiptUseCase := inbound.NewReceiptUseCaseImpl(visionClient, receiptRepo)
	authUseCase := inbound.NewAuthUseCaseImpl(userRepo, cfg.JWTSecret)
	cashShiftUseCase := inbound.NewCashShiftUseCaseImpl(cashShiftRepo)
	financeUseCase := inbound.NewFinanceUseCaseImpl(financeRepo)
	accessMenuUseCase := inbound.NewAccessMenuUseCaseImpl(accessMenuRepo)
	paymentUseCase := inbound.NewPaymentUseCaseImpl(paymentRepo, midtransClient, orderRepo, wsHub)
	syncUseCase := inbound.NewSyncUseCaseImpl(syncRepo, cfg)

	// Background Auto-Sync Worker (runs periodically when SQLite fallback is active)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if !resilientMgr.IsPGHealthy(context.Background()) {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := syncUseCase.TriggerSync(ctx, "")
			cancel()
			if err == nil {
				log.Println("Background auto-sync completed successfully.")
			}
		}
	}()

	// 7. Initialize REST & WebSocket HTTP Server
	server := rest.NewServer(cfg, productRepo, orderUseCase, wsHub, receiptUseCase, authUseCase, orderRepo, cashShiftUseCase, financeUseCase, accessMenuUseCase, paymentUseCase, syncUseCase)

	log.Printf("Starting Smart AI POS Engine Server on port :%s (env: %s)...", cfg.Port, cfg.Env)
	log.Printf("Real-time WebSocket endpoint available at ws://localhost:%s/ws", cfg.Port)
	log.Printf("AI Receipt Scanning endpoint available at POST http://localhost:%s/api/receipts/scan", cfg.Port)
	log.Printf("Authentication endpoints available at POST http://localhost:%s/api/auth/login & GET http://localhost:%s/api/auth/me", cfg.Port, cfg.Port)

	if err := server.Run(); err != nil {
		log.Fatalf("Server shutdown with error: %v", err)
	}
}
