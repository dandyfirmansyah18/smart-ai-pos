package main

import (
	"log"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/adapters/infrastructure/redis"
	"github.com/pos-backend/internal/ports/inbound"
)

func main() {
	// 1. Load Viper Configuration
	cfg := config.Load()

	// 2. Initialize PostgreSQL connection pool
	db, err := pg.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Warning: PostgreSQL ping failed (%v). Ensure PostgreSQL is running on %s:%s.", err, cfg.DBHost, cfg.DBPort)
	}

	// 3. Initialize Redis client & Lock Service
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Notice: Redis connection error (%v). Make sure Redis is running on %s.", err, cfg.RedisAddr)
	}
	lockService := redis.NewRedisLockService(redisClient)

	// 4. Initialize Driven Adapters (Repositories)
	productRepo := pg.NewProductPGRepository(db)
	orderRepo := pg.NewOrderPGRepository(db)

	// 5. Initialize Driver Use Cases
	orderUseCase := inbound.NewOrderUseCaseImpl(db, productRepo, orderRepo, lockService)

	// 6. Initialize REST HTTP Server & Routes
	server := rest.NewServer(cfg, productRepo, orderUseCase)

	log.Printf("Starting Smart AI POS Engine REST Server on port :%s (env: %s)...", cfg.Port, cfg.Env)
	if err := server.Run(); err != nil {
		log.Fatalf("Server shutdown with error: %v", err)
	}
}
