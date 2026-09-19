package ingest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
)

// Store is the persistence the sync service needs. Declared here so tests can
// substitute a fake without importing pgx.
type Store interface {
	Source(ctx context.Context, id int64) (Source, error)
	DueSources(ctx context.Context) ([]Source, error)
	StartRun(ctx context.Context, sourceID int64) (int64, error)
	FinishRun(ctx context.Context, id int64, status string, ok, fail int, errMsg string) error
	MarkSource(ctx context.Context, id int64, status, errMsg string) error
	PersistItem(ctx context.Context, src Source, item RawItem) (persistResult, error)
}

// Service runs one source sync: fetch → store raw → match → upsert offers.
type Service struct {
	repo     Store
	fetchers *Registry
	log      *slog.Logger
}

func NewService(store Store, fetchers *Registry, log *slog.Logger) *Service {
	return &Service{
		repo:     store,
		fetchers: fetchers,
		log:      log,
	}
}

// DueSources returns enabled sources whose schedule is due.
func (s *Service) DueSources(ctx context.Context) ([]Source, error) {
	return s.repo.DueSources(ctx)
}

// SyncSource fetches one source and writes catalogue rows. It is idempotent.
func (s *Service) SyncSource(ctx context.Context, sourceID int64) error {
	src, err := s.repo.Source(ctx, sourceID)
	if err != nil {
		return err
	}
	name, err := fetcherName(src.Config)
	if err != nil {
		return err
	}
	fetcher, ok := s.fetchers.Get(name)
	if !ok {
		return fmt.Errorf("no fetcher registered for %q", name)
	}

	s.log.InfoContext(ctx, "sync.started", "source_id", src.ID, "fetcher", name)

	runID, err := s.repo.StartRun(ctx, src.ID)
	if err != nil {
		return err
	}

	items, fetchErr := fetcher.Fetch(ctx, src)
	if fetchErr != nil {
		_ = s.repo.FinishRun(ctx, runID, "failed", 0, 0, fetchErr.Error())
		_ = s.repo.MarkSource(ctx, src.ID, "failed", fetchErr.Error())
		if crawlhttp.IsTransient(fetchErr) {
			s.log.WarnContext(ctx, "crawl.failed", "source_id", src.ID, "error", fetchErr.Error(), "class", "transient")
		} else {
			s.log.ErrorContext(ctx, "crawl.failed", "source_id", src.ID, "error", fetchErr.Error(), "class", "structural")
		}
		return fetchErr
	}

	var okCount, failCount, created, linked, queued, priceChanges int
	for _, item := range items {
		res, err := s.repo.PersistItem(ctx, src, item)
		if err != nil {
			failCount++
			s.log.WarnContext(ctx, "sync.item_failed",
				"source_id", src.ID, "external_id", item.ExternalID, "error", err.Error())
			continue
		}
		okCount++
		if res.created {
			created++
		}
		if res.linked {
			linked++
		}
		if res.queued {
			queued++
		}
		if res.priceChanged {
			priceChanges++
		}
	}

	status := "succeeded"
	errMsg := ""
	if failCount > 0 && okCount == 0 {
		status = "failed"
		errMsg = "every item failed"
	}
	if err := s.repo.FinishRun(ctx, runID, status, okCount, failCount, errMsg); err != nil {
		return err
	}
	if err := s.repo.MarkSource(ctx, src.ID, status, errMsg); err != nil {
		return err
	}

	s.log.InfoContext(ctx, "match.linked",
		"source_id", src.ID, "created", created, "linked", linked)
	s.log.InfoContext(ctx, "match.queued",
		"source_id", src.ID, "count", queued)
	s.log.InfoContext(ctx, "sync.completed",
		"source_id", src.ID,
		"ok", okCount,
		"failed", failCount,
		"created", created,
		"linked", linked,
		"queued", queued)
	s.log.InfoContext(ctx, "offer.updated",
		"source_id", src.ID, "count", priceChanges)

	if status == "failed" {
		return fmt.Errorf("sync source %d: %s", src.ID, errMsg)
	}
	return nil
}
