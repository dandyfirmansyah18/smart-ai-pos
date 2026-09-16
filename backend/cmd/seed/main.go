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
	downFlag := flag.Bool("down", false, "Rollback seed data")
	seedsDir := flag.String("dir", "seeds", "Path to seeds directory")
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

	// 1. Ensure seeds_history table exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS seeds_history (
		seed_name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create seeds_history table: %v", err)
	}

	pattern := "*.up.sql"
	if *downFlag {
		pattern = "*.down.sql"
	}

	files, err := filepath.Glob(filepath.Join(*seedsDir, pattern))
	if err != nil {
		log.Fatalf("Failed to search seed files: %v", err)
	}

	if len(files) == 0 {
		log.Printf("No seed files found in %s matching %s", *seedsDir, pattern)
		return
	}

	sort.Strings(files)

	// Fetch applied seeds
	appliedSeeds := make(map[string]bool)
	rows, err := db.Query("SELECT seed_name FROM seeds_history")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err == nil {
				appliedSeeds[s] = true
			}
		}
	}

	if *downFlag {
		// Reverse order for down seeds
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}

		log.Printf("Executing seed rollbacks (DOWN)...")
		for _, file := range files {
			seedName := filepath.Base(file)
			upSeedName := ""
			if len(file) > 9 {
				upSeedName = filepath.Base(file[:len(file)-9] + ".up.sql")
			}
			if !appliedSeeds[upSeedName] && !appliedSeeds[seedName] {
				log.Printf("Seed %s not applied, skipping rollback.", seedName)
				continue
			}

			content, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read seed file %s: %v", file, err)
			}

			log.Printf("Rolling back seed: %s", seedName)
			if _, err := db.Exec(string(content)); err != nil {
				log.Fatalf("Failed rolling back seed %s: %v", file, err)
			}

			// Remove from seeds_history
			_, _ = db.Exec("DELETE FROM seeds_history WHERE seed_name = $1 OR seed_name = $2", seedName, upSeedName)
		}
		fmt.Println("Seed rollbacks executed successfully!")
		return
	}

	log.Printf("Checking seed files (UP)...")
	runCount := 0
	for _, file := range files {
		seedName := filepath.Base(file)
		if appliedSeeds[seedName] {
			log.Printf("Seed already applied, skipping: %s", seedName)
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read seed file %s: %v", file, err)
		}

		log.Printf("Running seed: %s", seedName)
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Failed executing seed %s: %v", file, err)
		}

		// Record applied seed name
		if _, err := db.Exec("INSERT INTO seeds_history (seed_name) VALUES ($1) ON CONFLICT (seed_name) DO NOTHING", seedName); err != nil {
			log.Fatalf("Failed recording seed history %s: %v", seedName, err)
		}
		runCount++
	}

	if runCount == 0 {
		log.Println("All seed files are already up to date. No new seeds to run.")
	} else {
		fmt.Printf("Successfully executed %d new seed files!\n", runCount)
	}
}
