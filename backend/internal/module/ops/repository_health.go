package ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

func (r *repository) ListSourceHealth(ctx context.Context) ([]SourceHealth, error) {
	rows, err := r.q.ListSourceHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("list source health: %w", err)
	}
	out := make([]SourceHealth, 0, len(rows))
	for _, row := range rows {
		item := SourceHealth{
			ID:               row.ID,
			PharmacyID:       row.PharmacyID,
			PharmacyName:     row.PharmacyName,
			PharmacySlug:     row.PharmacySlug,
			Kind:             row.Kind,
			ScheduleInterval: row.ScheduleInterval,
			Enabled:          row.IsEnabled,
			LastRunAt:        row.LastRunAt,
			LastStatus:       deref(row.LastStatus),
			LastError:        deref(row.LastError),
			RejectCount:      row.PriceRejectCount,
			LowTrust:         row.PriceRejectCount >= lowTrustRejects,
		}
		if row.RunID > 0 {
			run := SyncRun{
				ID:         row.RunID,
				Status:     row.RunStatus,
				StartedAt:  row.RunStartedAt,
				FinishedAt: row.RunFinishedAt,
				OkCount:    row.RunOkCount,
				FailCount:  row.RunFailCount,
				Error:      deref(row.RunError),
			}
			item.LatestRun = &run
			if row.RunFinishedAt != nil {
				item.RunDurationMS = row.RunFinishedAt.Sub(row.RunStartedAt).Milliseconds()
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *repository) ListSyncSuccessRates(ctx context.Context, since time.Time, sourceID int64) (map[int64]successRate, error) {
	rows, err := r.q.ListSyncSuccessRates(ctx, store.ListSyncSuccessRatesParams{Since: since, SourceID: sourceID})
	if err != nil {
		return nil, fmt.Errorf("list sync success rates: %w", err)
	}
	out := make(map[int64]successRate, len(rows))
	for _, row := range rows {
		out[row.SourceID] = successRate{Runs: row.Runs, Succeeded: row.Succeeded, Failed: row.Failed}
	}
	return out, nil
}

func (r *repository) ListSourceRuns(ctx context.Context, sourceID int64, limit int32) ([]SyncRun, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	rows, err := r.q.ListSourceSyncRuns(ctx, store.ListSourceSyncRunsParams{SourceID: sourceID, PageLimit: limit})
	if err != nil {
		return nil, fmt.Errorf("list source runs: %w", err)
	}
	out := make([]SyncRun, 0, len(rows))
	for _, row := range rows {
		out = append(out, SyncRun{
			ID: row.ID, Status: row.Status, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
			OkCount: row.OkCount, FailCount: row.FailCount, Error: deref(row.Error),
		})
	}
	return out, nil
}

func (r *repository) ListStaleSources(ctx context.Context, now time.Time) ([]StaleSource, error) {
	rows, err := r.q.ListStaleSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("list stale sources: %w", err)
	}
	out := make([]StaleSource, 0, len(rows))
	for _, row := range rows {
		age := now.Sub(row.OldestSeen).Hours()
		if age < 0 {
			age = 0
		}
		out = append(out, StaleSource{
			ID: row.ID, PharmacyName: row.PharmacyName, PharmacySlug: row.PharmacySlug,
			OldestSeen: row.OldestSeen, OfferCount: row.OfferCount, AgeHours: age,
		})
	}
	return out, nil
}

func (r *repository) ListSuspicious(ctx context.Context, limit, offset int32) ([]SuspiciousOffer, int64, error) {
	total, err := r.q.CountSuspiciousOffers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count suspicious offers: %w", err)
	}
	rows, err := r.q.ListSuspiciousOffers(ctx, store.ListSuspiciousOffersParams{PageLimit: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("list suspicious offers: %w", err)
	}
	out := make([]SuspiciousOffer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSuspiciousRow(row))
	}
	return out, total, nil
}

func (r *repository) GetSuspicious(ctx context.Context, id int64) (SuspiciousOffer, error) {
	row, err := r.q.GetSuspiciousOffer(ctx, id)
	if err != nil {
		return SuspiciousOffer{}, mapNotFound(err)
	}
	proposed := int64(0)
	if row.ProposedPriceRial != nil {
		proposed = *row.ProposedPriceRial
	}
	if row.Status != "suspicious" || row.ProposedPriceRial == nil {
		return SuspiciousOffer{}, ErrNotSuspicious
	}
	return SuspiciousOffer{
		Status: row.Status,
		ID:     row.ID, ProductID: row.ProductID, PharmacyID: row.PharmacyID,
		SourceID: row.SourceID, CurrentPriceRial: row.PriceRial, ProposedPriceRial: proposed,
		ProposedAt: row.ProposedAt, InStock: row.InStock,
	}, nil
}

func (r *repository) ApproveSuspicious(ctx context.Context, id int64) (SuspiciousOffer, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return SuspiciousOffer{}, fmt.Errorf("begin approve price: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	current, err := q.GetSuspiciousOffer(ctx, id)
	if err != nil {
		return SuspiciousOffer{}, mapNotFound(err)
	}
	if current.Status != "suspicious" || current.ProposedPriceRial == nil {
		return SuspiciousOffer{}, ErrNotSuspicious
	}

	row, err := q.ApproveSuspiciousOffer(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SuspiciousOffer{}, ErrNotSuspicious
		}
		return SuspiciousOffer{}, fmt.Errorf("approve suspicious offer %d: %w", id, err)
	}
	if err := q.InsertPriceHistory(ctx, store.InsertPriceHistoryParams{
		OfferID: row.ID, ProductID: row.ProductID, PharmacyID: row.PharmacyID,
		PriceRial: row.PriceRial, InStock: row.InStock,
	}); err != nil {
		return SuspiciousOffer{}, fmt.Errorf("insert approved price history: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return SuspiciousOffer{}, err
	}
	return SuspiciousOffer{
		ID: row.ID, ProductID: row.ProductID, PharmacyID: row.PharmacyID,
		CurrentPriceRial: row.PriceRial, ProposedPriceRial: row.PriceRial, InStock: row.InStock,
	}, nil
}

func (r *repository) RejectSuspicious(ctx context.Context, id, sourceID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reject price: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	if err := q.RejectSuspiciousOffer(ctx, id); err != nil {
		return fmt.Errorf("reject suspicious offer %d: %w", id, err)
	}
	if sourceID > 0 {
		if err := q.IncrementSourcePriceRejects(ctx, sourceID); err != nil {
			return fmt.Errorf("increment source rejects: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *repository) ClickReport(ctx context.Context, since, until time.Time) (ClickReport, error) {
	total, err := r.q.CountClicksInRange(ctx, store.CountClicksInRangeParams{RangeStart: since, RangeEnd: until})
	if err != nil {
		return ClickReport{}, fmt.Errorf("count clicks: %w", err)
	}
	phRows, err := r.q.ListClicksByPharmacy(ctx, store.ListClicksByPharmacyParams{RangeStart: since, RangeEnd: until})
	if err != nil {
		return ClickReport{}, fmt.Errorf("clicks by pharmacy: %w", err)
	}
	prRows, err := r.q.ListTopClickedProducts(ctx, store.ListTopClickedProductsParams{RangeStart: since, RangeEnd: until})
	if err != nil {
		return ClickReport{}, fmt.Errorf("top clicked products: %w", err)
	}
	pharmacies := make([]ClickPharmacy, 0, len(phRows))
	for _, row := range phRows {
		pharmacies = append(pharmacies, ClickPharmacy{ID: row.ID, Slug: row.Slug, Name: row.Name, Clicks: row.Clicks})
	}
	products := make([]ClickProduct, 0, len(prRows))
	for _, row := range prRows {
		products = append(products, ClickProduct{ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, Clicks: row.Clicks})
	}
	out := ClickReport{Since: since, Until: until, Total: total, Pharmacies: pharmacies, TopProducts: products}
	if len(pharmacies) > 0 {
		top := pharmacies[0]
		out.TopPharmacy = &top
	}
	return out, nil
}

func mapSuspiciousRow(row store.ListSuspiciousOffersRow) SuspiciousOffer {
	proposed := int64(0)
	if row.ProposedPriceRial != nil {
		proposed = *row.ProposedPriceRial
	}
	return SuspiciousOffer{
		ID: row.ID, ProductID: row.ProductID, ProductSlug: row.ProductSlug, ProductName: row.ProductName,
		PharmacyID: row.PharmacyID, PharmacyName: row.PharmacyName, PharmacySlug: row.PharmacySlug,
		SourceID: row.SourceID, SourceRejectCount: row.SourceRejectCount,
		CurrentPriceRial: row.PriceRial, ProposedPriceRial: proposed, ProposedAt: row.ProposedAt,
		ProductURL: row.ProductUrl, LastSeenAt: row.LastSeenAt, InStock: row.InStock,
	}
}
