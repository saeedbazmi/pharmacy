// Package config loads and validates process configuration from the
// environment. Missing or malformed values fail at startup with a clear
// message rather than surfacing later as a confusing runtime error.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
)

// Config holds every setting the api and worker processes need.
type Config struct {
	Env                string
	LogLevel           slog.Level
	HTTPAddr           string
	DatabaseURL        string
	DBMaxConns         int32
	DBTimeout          time.Duration
	RequestTimeout     time.Duration
	ShutdownTimeout    time.Duration
	MaxBodyBytes       int64
	PriceJumpRatio     float64
	OfferStaleAfter    time.Duration
	OfferCriticalAfter time.Duration
	CrawlTimeout       time.Duration
	OTPPepper          string
	OTPPrintCode       bool
	WorkerHTTPAddr     string
}

// IsProduction reports whether the process runs in the production environment.
func (c Config) IsProduction() bool { return c.Env == "production" }

// Load reads configuration from the environment and validates it.
func Load() (Config, error) {
	var problems []string
	fail := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	cfg := Config{
		Env:            optionalString("APP_ENV", "development"),
		HTTPAddr:       optionalString("HTTP_ADDR", ":8080"),
		WorkerHTTPAddr: optionalString("WORKER_HTTP_ADDR", ":8081"),
	}

	switch cfg.Env {
	case "development", "production":
	default:
		fail("APP_ENV must be development or production, got %q", cfg.Env)
	}

	level, err := logger.ParseLevel(optionalString("LOG_LEVEL", "info"))
	if err != nil {
		fail("LOG_LEVEL: %v", err)
	}
	cfg.LogLevel = level

	cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if cfg.DatabaseURL == "" {
		fail("DATABASE_URL is required")
	}

	maxConns, err := optionalInt("DB_MAX_CONNS", 10)
	if err != nil {
		fail("%v", err)
	} else if maxConns < 1 {
		fail("DB_MAX_CONNS must be at least 1, got %d", maxConns)
	}
	cfg.DBMaxConns = int32(maxConns)

	if cfg.DBTimeout, err = optionalDuration("DB_TIMEOUT", 3*time.Second); err != nil {
		fail("%v", err)
	}
	if cfg.RequestTimeout, err = optionalDuration("HTTP_REQUEST_TIMEOUT", 15*time.Second); err != nil {
		fail("%v", err)
	}
	if cfg.ShutdownTimeout, err = optionalDuration("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		fail("%v", err)
	}

	maxBody, err := optionalInt("HTTP_MAX_BODY_BYTES", 1<<20)
	if err != nil {
		fail("%v", err)
	} else if maxBody < 1024 {
		fail("HTTP_MAX_BODY_BYTES must be at least 1024, got %d", maxBody)
	}
	cfg.MaxBodyBytes = int64(maxBody)

	ratio, err := optionalFloat("PRICE_JUMP_RATIO", 0.70)
	if err != nil {
		fail("%v", err)
	} else if ratio <= 0 || ratio > 5 {
		fail("PRICE_JUMP_RATIO must be between 0 exclusive and 5, got %g", ratio)
	}
	cfg.PriceJumpRatio = ratio

	if cfg.OfferStaleAfter, err = optionalDuration("OFFER_STALE_AFTER", 24*time.Hour); err != nil {
		fail("%v", err)
	}
	if cfg.OfferCriticalAfter, err = optionalDuration("OFFER_CRITICAL_AFTER", 72*time.Hour); err != nil {
		fail("%v", err)
	}
	if cfg.OfferCriticalAfter < cfg.OfferStaleAfter {
		fail("OFFER_CRITICAL_AFTER must be greater than or equal to OFFER_STALE_AFTER")
	}
	if cfg.CrawlTimeout, err = optionalDuration("CRAWL_TIMEOUT", 15*time.Second); err != nil {
		fail("%v", err)
	}

	printDefault := cfg.Env != "production"
	if cfg.OTPPrintCode, err = optionalBool("OTP_PRINT_CODE", printDefault); err != nil {
		fail("%v", err)
	}
	if cfg.IsProduction() && cfg.OTPPrintCode {
		fail("OTP_PRINT_CODE must be false in production")
	}
	pepperDefault := ""
	if !cfg.IsProduction() {
		pepperDefault = "dev-only-otp-pepper"
	}
	cfg.OTPPepper = optionalString("OTP_PEPPER", pepperDefault)
	if cfg.IsProduction() && len(cfg.OTPPepper) < 16 {
		fail("OTP_PEPPER is required in production and must be at least 16 characters")
	}

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

func optionalString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func optionalInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q", key, raw)
	}
	return v, nil
}

func optionalFloat(key string, fallback float64) (float64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number, got %q", key, raw)
	}
	return v, nil
}

func optionalDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 3s or 500ms, got %q", key, raw)
	}
	if v <= 0 {
		return 0, errors.New(key + " must be greater than zero")
	}
	return v, nil
}

func optionalBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false, got %q", key, raw)
	}
	return v, nil
}
