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

	"github.com/saeedbazmi/pharmacy/backend/internal/module/alerting"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/identity"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/ingest"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/redirect"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/config"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/jobq"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/postgres"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(workerLiveCheck())
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "worker failed to start: %v\n", err)
		os.Exit(1)
	}
}

func workerLiveCheck() int {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8081/healthz")
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

	httpClient := crawlhttp.NewWithTimeout(cfg.CrawlTimeout)
	fetchers := ingest.NewRegistry()
	fetchers.Register("darukade", ingest.NewDarukadeFetcher(httpClient))
	fetchers.Register("rosha", ingest.NewRoshaFetcher(httpClient))

	svc := ingest.NewService(ingest.NewStore(pool, cfg.PriceJumpRatio, log), fetchers, log)
	runner := ingest.NewRunner(pool, svc, log, 2)
	queries := store.New(pool)
	queue := jobq.New(pool)
	runner.Handle(jobq.KindEnsurePartitions, func(ctx context.Context, _ jobq.Job) error {
		return redirect.EnsureMonths(ctx, queries)
	})

	var sms identity.SMSSender
	if cfg.IsProduction() {
		sms = identity.NewUnavailableSender()
	} else {
		sms = identity.NewDevSender(cfg.OTPPrintCode, nil)
	}
	alertSvc := alerting.NewService(alerting.NewStore(pool), sms, cfg.OfferStaleAfter, log)
	runner.Handle(jobq.KindEvaluateAlerts, func(ctx context.Context, _ jobq.Job) error {
		return alertSvc.Evaluate(ctx)
	})

	if _, err := queue.EnqueueEnsurePartitions(ctx); err != nil {
		return fmt.Errorf("schedule partitions: %w", err)
	}
	if _, err := queue.EnqueueEvaluateAlerts(ctx); err != nil {
		return fmt.Errorf("schedule alerts: %w", err)
	}
	if err := runner.ScheduleDue(ctx); err != nil {
		return fmt.Errorf("initial schedule: %w", err)
	}

	healthMux := http.NewServeMux()
	healthMux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	healthMux.Handle("GET /readyz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingCtx, cancel := context.WithTimeout(r.Context(), cfg.DBTimeout)
		defer cancel()
		if err := pool.Ping(pingCtx); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","database":"unreachable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","database":"ok"}`))
	}))
	healthSrv := &http.Server{
		Addr:              cfg.WorkerHTTPAddr,
		Handler:           healthMux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.InfoContext(ctx, "worker.health_started", "addr", cfg.WorkerHTTPAddr)
		if err := healthSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.ErrorContext(ctx, "worker.health_failed", "error", err.Error())
		}
	}()
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = healthSrv.Shutdown(shutCtx)
	}()

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := queue.EnqueueEnsurePartitions(ctx); err != nil {
					log.WarnContext(ctx, "schedule.partitions_failed", "error", err.Error())
				}
				if _, err := queue.EnqueueEvaluateAlerts(ctx); err != nil {
					log.WarnContext(ctx, "schedule.alerts_failed", "error", err.Error())
				}
				if err := runner.ScheduleDue(ctx); err != nil {
					log.WarnContext(ctx, "schedule.failed", "error", err.Error())
				}
			}
		}
	}()

	log.InfoContext(ctx, "worker.started", "env", cfg.Env)
	if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
		return err
	}
	log.Info("worker.stopped")
	return nil
}
