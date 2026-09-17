// Command migrate applies the embedded SQL migrations in filename order.
// Migrations are forward only: an applied file is never edited, a fix always
// arrives as a new file.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/saeedbazmi/pharmacy/backend/db"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New("migrate", cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := postgres.Open(ctx, postgres.Options{
		DatabaseURL:  cfg.DatabaseURL,
		MaxConns:     2,
		QueryTimeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, conn.Conn())
	if err != nil {
		return err
	}

	files, err := migrationFiles()
	if err != nil {
		return err
	}

	pending := 0
	for _, name := range files {
		if applied[name] {
			continue
		}
		pending++
		started := time.Now()
		if err := applyMigration(ctx, conn.Conn(), name); err != nil {
			log.ErrorContext(ctx, "migration.failed", "version", name, "error", err.Error())
			return err
		}
		log.InfoContext(ctx, "migration.applied",
			"version", name, "duration_ms", time.Since(started).Milliseconds())
	}

	log.InfoContext(ctx, "migration.completed",
		"applied", pending, "total", len(files), "already_applied", len(applied))
	return nil
}

func appliedVersions(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func migrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(db.Migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return nil, errors.New("no migrations embedded in the binary")
	}
	sort.Strings(names)
	return names, nil
}

// applyMigration runs one file and records it in the same transaction, so a
// failure leaves neither a half-applied schema nor a false version record.
func applyMigration(ctx context.Context, conn *pgx.Conn, name string) error {
	body, err := db.Migrations.ReadFile("migrations/" + name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}
