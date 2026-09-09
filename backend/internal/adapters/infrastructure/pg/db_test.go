package pg_test

import (
	"os"
	"testing"

	"github.com/pos-backend/internal/adapters/infrastructure/pg"
	"github.com/pos-backend/pkg/config"
)

func TestConfigAndDBSetup(t *testing.T) {
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg := config.Load()
	expectedDSN := "postgres://testuser:secret@127.0.0.1:5432/testdb?sslmode=disable"
	if cfg.DSN() != expectedDSN {
		t.Errorf("expected DSN %s, got %s", expectedDSN, cfg.DSN())
	}

	db, err := pg.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("expected no error setting up DB pool, got %v", err)
	}
	defer db.Close()
}
