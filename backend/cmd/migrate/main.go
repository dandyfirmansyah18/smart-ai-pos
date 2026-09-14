package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/infrastructure/pg"
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

	// 1. Ensure schema_migrations table exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
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

	// Fetch applied migrations
	appliedVersions := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err == nil {
				appliedVersions[v] = true
			}
		}
	}

	if *downFlag {
		// Reverse order for down migrations
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}

		log.Printf("Executing rollbacks (DOWN)...")
		for _, file := range files {
			version := filepath.Base(file)
			upVersion := ""
			if len(file) > 9 {
				upVersion = filepath.Base(file[:len(file)-9] + ".up.sql")
			}
			if !appliedVersions[upVersion] && !appliedVersions[version] {
				log.Printf("Migration %s not applied, skipping rollback.", version)
				continue
			}

			content, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read migration file %s: %v", file, err)
			}

			log.Printf("Rolling back migration: %s", version)
			if _, err := db.Exec(string(content)); err != nil {
				log.Fatalf("Failed rolling back migration %s: %v", file, err)
			}

			// Remove from schema_migrations
			_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = $1 OR version = $2", version, upVersion)
		}
		fmt.Println("Rollback migrations executed successfully!")
		return
	}

	log.Printf("Checking migrations (UP)...")
	runCount := 0
	for _, file := range files {
		version := filepath.Base(file)
		if appliedVersions[version] {
			log.Printf("Migration already applied, skipping: %s", version)
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		log.Printf("Running migration: %s", version)
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Failed executing migration %s: %v", file, err)
		}

		// Record applied migration version
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", version); err != nil {
			log.Fatalf("Failed recording migration version %s: %v", version, err)
		}
		runCount++
	}

	if runCount == 0 {
		log.Println("Database is already up to date. No new migrations to run.")
	} else {
		fmt.Printf("Successfully executed %d new migrations!\n", runCount)
	}
}
