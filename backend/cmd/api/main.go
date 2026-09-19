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
	"github.com/saeedbazmi/pharmacy/backend/internal/module/admin"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/alerting"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/identity"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/ingest"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/redirect"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/search"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/cache"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/jobq"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(liveCheck())
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api failed to start: %v\n", err)
		os.Exit(1)
	}
}

// liveCheck is used by the container healthcheck. It only asks whether this
// process is serving; readiness of dependencies is /readyz.
func liveCheck() int {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8080/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
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

	catalogService := catalog.NewService(catalog.NewRepositoryWithFreshness(pool, cfg.OfferCriticalAfter), log)
	catalogService.SetFreshness(cfg.OfferStaleAfter, cfg.OfferCriticalAfter)
	clickRepo := redirect.NewRepository(pool)
	clickRecorder := redirect.NewRecorder(clickRepo, log)
	defer clickRecorder.Stop()
	redirectHandler := redirect.NewHandler(redirect.NewService(clickRepo, clickRecorder, log))
	searchRecorder := search.NewRecorder(search.NewLogStore(pool), log)
	defer searchRecorder.Stop()
	searchService := search.NewService(
		search.NewPostgresWithFreshness(pool, cfg.OfferCriticalAfter),
		searchRecorder,
		cache.NewTTL[search.Result](30*time.Second, 512),
		log,
	)
	searchHandler := search.NewHandler(searchService, apphttp.ErrorWriter(log))

	httpClient := crawlhttp.NewWithTimeout(cfg.CrawlTimeout)
	fetchers := ingest.NewRegistry()
	fetchers.Register("darukade", ingest.NewDarukadeFetcher(httpClient))
	fetchers.Register("rosha", ingest.NewRoshaFetcher(httpClient))

	var sms identity.SMSSender
	if cfg.IsProduction() {
		sms = identity.NewUnavailableSender()
	} else {
		sms = identity.NewDevSender(cfg.OTPPrintCode, nil)
	}
	identityHandler := identity.NewHandler(
		identity.NewService(identity.NewStore(pool), sms, cfg.OTPPepper, log),
		apphttp.ErrorWriter(log),
		cfg.IsProduction(),
	)
	alertingHandler := alerting.NewHandler(
		alerting.NewService(alerting.NewStore(pool), sms, cfg.OfferStaleAfter, log),
		apphttp.ErrorWriter(log),
	)
	opsService := ops.NewService(ops.NewStore(pool), fetchers, jobq.New(pool), log)
	opsHandler := ops.NewHandler(opsService, apphttp.ErrorWriter(log), cfg.IsProduction())
	adminHandler := admin.NewHandler(
		admin.NewService(admin.NewStore(pool), log),
		opsService,
		apphttp.ErrorWriter(log),
	)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: apphttp.NewRouter(apphttp.Deps{
			Log:            log,
			DB:             pool,
			Catalog:        catalogService,
			Search:         searchHandler,
			Redirect:       redirectHandler,
			Ops:            opsHandler,
			Identity:       identityHandler,
			Alerting:       alertingHandler,
			Admin:          adminHandler,
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
