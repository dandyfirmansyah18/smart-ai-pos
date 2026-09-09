package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/config"
)

func main() {
	downFlag := flag.Bool("down", false, "Rollback migrations")
	migrationsDir := flag.String("dir", "migrations", "Path to migrations directory")
	flag.Parse()

	cfg := config.Load()
	db, err := pg.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %v. Make sure PostgreSQL is running on %s:%s.", err, cfg.DBHost, cfg.DBPort)
	}

	pattern := "*.up.sql"
	if *downFlag {
		pattern = "*.down.sql"
	}

	files, err := filepath.Glob(filepath.Join(*migrationsDir, pattern))
	if err != nil {
		log.Fatalf("Failed to search migration files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No migration files found in %s matching %s", *migrationsDir, pattern)
		return
	}

	sort.Strings(files)
	if *downFlag {
		// Reverse order for down migrations
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	log.Printf("Executing migrations (direction: %s, count: %d)...", map[bool]string{false: "UP", true: "DOWN"}[*downFlag], len(files))

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		log.Printf("Running migration: %s", filepath.Base(file))
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Failed executing migration %s: %v", file, err)
		}
	}

	fmt.Println("Migrations executed successfully!")
}
