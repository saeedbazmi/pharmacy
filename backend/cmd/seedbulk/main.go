package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

const targetBulk = 30000

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed-bulk failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := postgres.Open(ctx, postgres.Options{
		DatabaseURL:  cfg.DatabaseURL,
		MaxConns:     4,
		QueryTimeout: 2 * time.Minute,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	var existing int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE slug LIKE 'bulk-%'`).Scan(&existing); err != nil {
		return err
	}
	need := targetBulk - existing
	if need <= 0 {
		fmt.Printf("bulk seed already present (%d rows)\n", existing)
		return nil
	}
	fmt.Printf("inserting %d bulk products...\n", need)
	return insertBulk(ctx, pool, existing, need)
}

func insertBulk(ctx context.Context, pool *pgxpool.Pool, start, n int) error {
	const batch = 1000
	cols := []string{"slug", "name_fa", "name_en", "generic_name", "status", "name_normalized", "search_document"}
	for offset := 0; offset < n; offset += batch {
		end := offset + batch
		if end > n {
			end = n
		}
		rows := make([][]any, 0, end-offset)
		for i := start + offset; i < start+end; i++ {
			nameFa := fmt.Sprintf("کرم دور چشم نمونه %d", i)
			nameEn := fmt.Sprintf("Eye cream sample %d", i)
			slug := fmt.Sprintf("bulk-eye-cream-%d", i)
			rows = append(rows, []any{
				slug,
				nameFa,
				nameEn,
				"eye cream",
				"published",
				textfa.Normalize(nameFa),
				textfa.Document(nameFa, nameEn, "eye cream"),
			})
		}
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"products"}, cols, pgx.CopyFromRows(rows)); err != nil {
			return fmt.Errorf("copy batch %d: %w", start+offset, err)
		}
	}
	fmt.Printf("inserted %d bulk products\n", n)
	return nil
}
