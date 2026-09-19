// Command opsuser creates the first data-ops login for local development.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "opsuser failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	username := strings.TrimSpace(os.Getenv("OPS_USERNAME"))
	password := os.Getenv("OPS_PASSWORD")
	role := strings.TrimSpace(os.Getenv("OPS_ROLE"))
	if username == "" {
		username = "ops"
	}
	if password == "" {
		password = "ops-dev-password"
	}
	if role == "" {
		role = ops.RoleDataOps
	}
	hash, err := ops.HashPassword(password)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := postgres.Open(ctx, postgres.Options{
		DatabaseURL:  cfg.DatabaseURL,
		MaxConns:     2,
		QueryTimeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	id, err := ops.NewStore(pool).InsertUser(ctx, username, hash, role)
	if err != nil {
		return fmt.Errorf("insert user %s: %w", username, err)
	}
	fmt.Printf("created ops user %s id=%d role=%s\n", username, id, role)
	return nil
}
