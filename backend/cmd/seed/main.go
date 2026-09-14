package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/internal/domain"
)

func main() {
	cfg := config.Load()
	db, err := pg.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v. Make sure PostgreSQL is running on %s:%s.", err, cfg.DBHost, cfg.DBPort)
	}

	ctx := context.Background()

	// 1. Ensure seeds_history table exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS seeds_history (
		seed_name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		log.Fatalf("Failed to create seeds_history table: %v", err)
	}

	seedName := "initial_products_catalog_v1"

	// 2. Check if seed has already been applied
	var appliedAt sql.NullTime
	err = db.QueryRowContext(ctx, "SELECT applied_at FROM seeds_history WHERE seed_name = $1", seedName).Scan(&appliedAt)
	if err == nil && appliedAt.Valid {
		log.Printf("Seed '%s' has already been applied on %s. Skipping seeder execution.", seedName, appliedAt.Time.Format("2006-01-02 15:04:05"))
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatalf("Failed checking seeds history: %v", err)
	}

	seedProducts := []domain.Product{
		{
			ID:            uuid.New(),
			SKU:           "SKU-COFFEE-001",
			Name:          "Espresso Double Shot",
			Description:   "Rich blend double espresso shot",
			Price:         3.50,
			StockQuantity: 100,
		},
		{
			ID:            uuid.New(),
			SKU:           "SKU-COFFEE-002",
			Name:          "Iced Oat Latte",
			Description:   "Cold brewed espresso with organic oat milk",
			Price:         5.00,
			StockQuantity: 80,
		},
		{
			ID:            uuid.New(),
			SKU:           "SKU-FOOD-001",
			Name:          "Avocado Toast",
			Description:   "Sourdough toast topped with fresh avocado and seeds",
			Price:         8.50,
			StockQuantity: 50,
		},
		{
			ID:            uuid.New(),
			SKU:           "SKU-FOOD-002",
			Name:          "Croissant Butter",
			Description:   "Freshly baked French butter croissant",
			Price:         4.00,
			StockQuantity: 60,
		},
		{
			ID:            uuid.New(),
			SKU:           "SKU-DRINK-001",
			Name:          "Matcha Green Tea Latte",
			Description:   "Premium Uji matcha with steamed milk",
			Price:         5.50,
			StockQuantity: 75,
		},
	}

	log.Printf("Seeding initial products catalog (%s) into database...", seedName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO products (id, sku, name, description, price, stock_quantity, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (sku) DO UPDATE SET 
	              name = EXCLUDED.name,
	              description = EXCLUDED.description,
	              price = EXCLUDED.price,
	              stock_quantity = EXCLUDED.stock_quantity,
	              updated_at = CURRENT_TIMESTAMP`

	for _, p := range seedProducts {
		_, err := tx.ExecContext(ctx, query, p.ID, p.SKU, p.Name, p.Description, p.Price, p.StockQuantity)
		if err != nil {
			log.Fatalf("Failed seeding product %s (%s): %v", p.Name, p.SKU, err)
		}
		log.Printf("Seeded product: %s | SKU: %s | Price: $%.2f | Stock: %d", p.Name, p.SKU, p.Price, p.StockQuantity)
	}

	// Record seed execution in seeds_history
	_, err = tx.ExecContext(ctx, "INSERT INTO seeds_history (seed_name) VALUES ($1) ON CONFLICT (seed_name) DO NOTHING", seedName)
	if err != nil {
		log.Fatalf("Failed recording seed history: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed committing seed transaction: %v", err)
	}

	fmt.Printf("\nSuccessfully executed and recorded seeder '%s' with %d products!\n", seedName, len(seedProducts))
}
