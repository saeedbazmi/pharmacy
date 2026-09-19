package ops

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ListHealth(ctx context.Context, window time.Duration) ([]SourceHealth, error) {
	if window <= 0 {
		window = defaultHealthWindow
	}
	items, err := s.store.ListSourceHealth(ctx)
	if err != nil {
		return nil, err
	}
	rates, err := s.store.ListSyncSuccessRates(ctx, s.now().Add(-window), 0)
	if err != nil {
		return nil, err
	}
	now := s.now()
	for i := range items {
		if rate, ok := rates[items[i].ID]; ok {
			items[i].Runs = rate.Runs
			items[i].Succeeded = rate.Succeeded
			if rate.Runs > 0 {
				items[i].SuccessRate = float64(rate.Succeeded) / float64(rate.Runs)
			}
		}
		items[i].Overdue = sourceOverdue(items[i].LastRunAt, items[i].ScheduleInterval, items[i].Enabled, now)
	}
	return items, nil
}

func (s *Service) SourceRuns(ctx context.Context, id int64) ([]SyncRun, error) {
	if _, err := s.store.GetSource(ctx, id); err != nil {
		return nil, err
	}
	return s.store.ListSourceRuns(ctx, id, 20)
}

func (s *Service) ListStaleSources(ctx context.Context) ([]StaleSource, error) {
	return s.store.ListStaleSources(ctx, s.now())
}

func (s *Service) ListSuspicious(ctx context.Context, page, size int) (Page[SuspiciousOffer], error) {
	page, size = clampPage(page, size)
	items, total, err := s.store.ListSuspicious(ctx, int32(size), int32((page-1)*size))
	if err != nil {
		return Page[SuspiciousOffer]{}, err
	}
	return Page[SuspiciousOffer]{Items: items, Total: total, Page: page, Size: size}, nil
}

func (s *Service) ApprovePrice(ctx context.Context, actor Actor, id int64) error {
	before, err := s.store.GetSuspicious(ctx, id)
	if err != nil {
		return err
	}
	after, err := s.store.ApproveSuspicious(ctx, id)
	if err != nil {
		return err
	}
	s.log.InfoContext(ctx, "price.approved",
		"actor_id", actor.ID, "offer_id", id,
		"old_price_rial", before.CurrentPriceRial, "new_price_rial", after.CurrentPriceRial)
	s.audit(ctx, actor, "offer", fmt.Sprint(id), "approve_price", before, after)
	return nil
}

func (s *Service) RejectPrice(ctx context.Context, actor Actor, id int64) error {
	before, err := s.store.GetSuspicious(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.RejectSuspicious(ctx, id, before.SourceID); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "price.rejected",
		"actor_id", actor.ID, "offer_id", id, "source_id", before.SourceID)
	s.audit(ctx, actor, "offer", fmt.Sprint(id), "reject_price", before, map[string]int64{
		"kept_price_rial": before.CurrentPriceRial,
		"source_id":       before.SourceID,
	})
	return nil
}

func (s *Service) ClickReport(ctx context.Context, since, until time.Time) (ClickReport, error) {
	since, until, err := normalizeRange(since, until, s.now(), defaultClickWindow)
	if err != nil {
		return ClickReport{}, err
	}
	return s.store.ClickReport(ctx, since, until)
}

func sourceOverdue(lastRun *time.Time, schedule string, enabled bool, now time.Time) bool {
	if !enabled {
		return false
	}
	gap := parseInterval(schedule)
	if lastRun == nil {
		return true
	}
	return lastRun.Add(gap).Before(now)
}

func parseInterval(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Hour
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	parts := strings.Fields(raw)
	if len(parts) == 2 {
		n, err := strconv.Atoi(parts[0])
		if err == nil {
			switch {
			case strings.HasPrefix(parts[1], "day"):
				return time.Duration(n) * 24 * time.Hour
			case strings.HasPrefix(parts[1], "hour"):
				return time.Duration(n) * time.Hour
			case strings.HasPrefix(parts[1], "min"):
				return time.Duration(n) * time.Minute
			}
		}
	}
	var h, m, sec int
	if _, err := fmt.Sscanf(raw, "%d:%d:%d", &h, &m, &sec); err == nil {
		return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	}
	return time.Hour
}

func normalizeRange(since, until, now time.Time, fallback time.Duration) (time.Time, time.Time, error) {
	if until.IsZero() {
		until = now
	}
	if since.IsZero() {
		since = until.Add(-fallback)
	}
	if !since.Before(until) {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return since.UTC(), until.UTC(), nil
}
