package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

const validDSN = "postgres://user:pass@localhost:5432/pharmacy?sslmode=disable"

func TestLoadAppliesDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", validDSN)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.DBTimeout != 3*time.Second {
		t.Errorf("DBTimeout = %v, want 3s", cfg.DBTimeout)
	}
	if cfg.DBMaxConns != 10 {
		t.Errorf("DBMaxConns = %d, want 10", cfg.DBMaxConns)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail without DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("error should name the missing variable, got: %v", err)
	}
}

func TestLoadReportsAllProblemsAtOnce(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("DB_TIMEOUT", "soon")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail")
	}
	for _, want := range []string{"DATABASE_URL", "APP_ENV", "LOG_LEVEL", "DB_TIMEOUT"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %s, got: %v", want, err)
		}
	}
}

func TestLoadRejectsNonPositiveDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", validDSN)
	t.Setenv("DB_TIMEOUT", "0s")

	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a zero timeout")
	}
}

func TestIsProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", validDSN)
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
}
