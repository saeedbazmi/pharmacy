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
	if cfg.PriceJumpRatio != 0.70 {
		t.Errorf("PriceJumpRatio = %g, want 0.70", cfg.PriceJumpRatio)
	}
	if cfg.OfferStaleAfter != 24*time.Hour {
		t.Errorf("OfferStaleAfter = %v, want 24h", cfg.OfferStaleAfter)
	}
	if cfg.OfferCriticalAfter != 72*time.Hour {
		t.Errorf("OfferCriticalAfter = %v, want 72h", cfg.OfferCriticalAfter)
	}
	if cfg.CrawlTimeout != 15*time.Second {
		t.Errorf("CrawlTimeout = %v, want 15s", cfg.CrawlTimeout)
	}
	if !cfg.OTPPrintCode {
		t.Error("OTPPrintCode should default to true in development")
	}
	if cfg.OTPPepper != "dev-only-otp-pepper" {
		t.Errorf("OTPPepper = %q", cfg.OTPPepper)
	}
	if cfg.WorkerHTTPAddr != ":8081" {
		t.Errorf("WorkerHTTPAddr = %q, want :8081", cfg.WorkerHTTPAddr)
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
	t.Setenv("OTP_PEPPER", "production-otp-pepper-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
	if cfg.OTPPrintCode {
		t.Error("OTPPrintCode must default to false in production")
	}
}

func TestLoadProductionRequiresOTPPepper(t *testing.T) {
	t.Setenv("DATABASE_URL", validDSN)
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTP_PEPPER", "")
	t.Setenv("OTP_PRINT_CODE", "false")

	if _, err := Load(); err == nil {
		t.Fatal("production must require OTP_PEPPER")
	}
}
