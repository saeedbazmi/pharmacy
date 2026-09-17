// Package postgres opens and verifies the connection pool used by every
// service. Timeouts come from configuration; nothing here reaches the network
// without one.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Options configures the pool.
type Options struct {
	DatabaseURL  string
	MaxConns     int32
	QueryTimeout time.Duration
}

// Open parses the DSN, opens a pool and verifies connectivity before returning.
func Open(ctx context.Context, opts Options) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(opts.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	cfg.MaxConns = opts.MaxConns
	cfg.MinConns = 0
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 15 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	cfg.ConnConfig.ConnectTimeout = opts.QueryTimeout

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, opts.QueryTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
