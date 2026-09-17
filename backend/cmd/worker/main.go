// Command worker runs scheduled background work: source synchronisation,
// product matching and price alerts.
//
// M0 only establishes the process, its configuration and its shutdown path. The
// job queue and the ingestion pipeline arrive with M1 (PH1-009).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "worker failed to start: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New("worker", cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Open(ctx, postgres.Options{
		DatabaseURL:  cfg.DatabaseURL,
		MaxConns:     cfg.DBMaxConns,
		QueryTimeout: cfg.DBTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	log.InfoContext(ctx, "worker.started", "env", cfg.Env)

	// Placeholder loop: it proves configuration, database connectivity and
	// shutdown work end to end. PH1-009 replaces it with the real job runner.
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker.stopped")
			return nil
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, cfg.DBTimeout)
			err := pool.Ping(pingCtx)
			cancel()
			if err != nil {
				log.WarnContext(ctx, "worker.database_unreachable", "error", err.Error())
				continue
			}
			log.DebugContext(ctx, "worker.idle")
		}
	}
}
