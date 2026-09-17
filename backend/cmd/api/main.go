// Command api serves the public API, the data ops panel API and the admin API.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apphttp "github.com/saeedbazmi/pharmacy/backend/internal/http"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api failed to start: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New("api", cfg.LogLevel)

	// Signals cancel this context, which then drains the server.
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

	catalogService := catalog.NewService(catalog.NewRepository(pool), log)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: apphttp.NewRouter(apphttp.Deps{
			Log:            log,
			DB:             pool,
			Catalog:        catalogService,
			RequestTimeout: cfg.RequestTimeout,
			DBTimeout:      cfg.DBTimeout,
			MaxBodyBytes:   cfg.MaxBodyBytes,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 5*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "api.started", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Info("api.shutdown_requested")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("api.stopped")
	return nil
}
