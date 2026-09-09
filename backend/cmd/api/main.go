package main

import (
	"log"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/adapters/handlers/ws"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/adapters/infrastructure/redis"
	"github.com/pos-backend/internal/ports/inbound"
)

func main() {
	// 1. Load Viper Configuration
	cfg := config.Load()

	// 2. Initialize WebSocket Broadcaster Hub & start background event loop
	wsHub := ws.NewHub()
	go wsHub.Run()

	// 3. Initialize PostgreSQL connection pool
	db, err := pg.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Warning: PostgreSQL ping failed (%v). Ensure PostgreSQL is running on %s:%s.", err, cfg.DBHost, cfg.DBPort)
	}

	// 4. Initialize Redis client & Lock Service
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Notice: Redis connection error (%v). Make sure Redis is running on %s.", err, cfg.RedisAddr)
	}
	lockService := redis.NewRedisLockService(redisClient)

	// 5. Initialize Driven Adapters (Repositories)
	productRepo := pg.NewProductPGRepository(db)
	orderRepo := pg.NewOrderPGRepository(db)

	// 6. Initialize Driver Use Cases with WebSocket Broadcaster
	orderUseCase := inbound.NewOrderUseCaseImpl(db, productRepo, orderRepo, lockService, wsHub)

	// 7. Initialize REST & WebSocket HTTP Server
	server := rest.NewServer(cfg, productRepo, orderUseCase, wsHub)

	log.Printf("Starting Smart AI POS Engine Server on port :%s (env: %s)...", cfg.Port, cfg.Env)
	log.Printf("Real-time WebSocket endpoint available at ws://localhost:%s/ws", cfg.Port)

	if err := server.Run(); err != nil {
		log.Fatalf("Server shutdown with error: %v", err)
	}
}
