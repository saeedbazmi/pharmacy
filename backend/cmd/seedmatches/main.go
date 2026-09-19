// Command seedmatches queues pending match_candidates on existing source items
// so the ops panel has real review work.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seedmatches failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := postgres.Open(ctx, postgres.Options{
		DatabaseURL:  cfg.DatabaseURL,
		MaxConns:     2,
		QueryTimeout: 20 * time.Second,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	tag, err := pool.Exec(ctx, `
INSERT INTO match_candidates (source_item_id, suggested_product_id, score, status, reason)
SELECT si.id, si.product_id, 0.9100, 'pending', 'ops-panel-review'
FROM source_items si
WHERE si.product_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM match_candidates mc
      WHERE mc.source_item_id = si.id AND mc.status = 'pending'
  )
ORDER BY si.id
LIMIT 50
`)
	if err != nil {
		return err
	}
	fmt.Printf("queued %d pending match candidates\n", tag.RowsAffected())
	return nil
}
