package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

type repository struct {
	pool      *pgxpool.Pool
	q         *store.Queries
	jumpRatio float64
	log       *slog.Logger
}

func newRepository(pool *pgxpool.Pool, jumpRatio float64, log *slog.Logger) *repository {
	if jumpRatio <= 0 {
		jumpRatio = 0.70
	}
	if log == nil {
		log = slog.Default()
	}
	return &repository{pool: pool, q: store.New(pool), jumpRatio: jumpRatio, log: log}
}

// NewStore is the production Store backed by Postgres.
func NewStore(pool *pgxpool.Pool, jumpRatio float64, log *slog.Logger) Store {
	return newRepository(pool, jumpRatio, log)
}

func (r *repository) Source(ctx context.Context, id int64) (Source, error) {
	row, err := r.q.GetDataSource(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Source{}, fmt.Errorf("data source %d: not found", id)
		}
		return Source{}, fmt.Errorf("get data source %d: %w", id, err)
	}
	return Source{
		ID:         row.ID,
		PharmacyID: row.PharmacyID,
		Kind:       row.Kind,
		Config:     row.Config,
	}, nil
}

func (r *repository) DueSources(ctx context.Context) ([]Source, error) {
	rows, err := r.q.ListDueSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("list due sources: %w", err)
	}
	out := make([]Source, 0, len(rows))
	for _, row := range rows {
		out = append(out, Source{
			ID:         row.ID,
			PharmacyID: row.PharmacyID,
			Kind:       row.Kind,
			Config:     row.Config,
		})
	}
	return out, nil
}

func (r *repository) StartRun(ctx context.Context, sourceID int64) (int64, error) {
	id, err := r.q.InsertSyncRun(ctx, sourceID)
	if err != nil {
		return 0, fmt.Errorf("insert sync run: %w", err)
	}
	return id, nil
}

func (r *repository) FinishRun(ctx context.Context, id int64, status string, ok, fail int, errMsg string) error {
	return r.q.FinishSyncRun(ctx, store.FinishSyncRunParams{
		Status:    status,
		OkCount:   int32(ok),
		FailCount: int32(fail),
		Error:     strPtr(errMsg),
		ID:        id,
	})
}

func (r *repository) MarkSource(ctx context.Context, id int64, status, errMsg string) error {
	return r.q.MarkSourceRun(ctx, store.MarkSourceRunParams{
		LastStatus: strPtr(status),
		LastError:  strPtr(errMsg),
		ID:         id,
	})
}

type persistResult struct {
	created      bool
	linked       bool
	queued       bool
	priceChanged bool
}

func (r *repository) PersistItem(ctx context.Context, src Source, item RawItem) (persistResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return persistResult{}, fmt.Errorf("begin item tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	raw := item.Raw
	if len(raw) == 0 {
		raw, _ = json.Marshal(item)
	}

	srcRow, err := q.UpsertSourceItem(ctx, store.UpsertSourceItemParams{
		SourceID:   src.ID,
		ExternalID: item.ExternalID,
		ProductID:  nil,
		Raw:        raw,
	})
	if err != nil {
		return persistResult{}, fmt.Errorf("upsert source item %s: %w", item.ExternalID, err)
	}

	var byGTIN *ProductRef
	if item.GTIN != "" {
		row, err := q.GetProductByGTIN(ctx, strPtr(item.GTIN))
		if err == nil {
			byGTIN = &ProductRef{ID: row.ID, Slug: row.Slug}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return persistResult{}, fmt.Errorf("lookup gtin: %w", err)
		}
	}
	var byIRC *ProductRef
	if item.IRC != "" {
		row, err := q.GetProductByIRC(ctx, strPtr(item.IRC))
		if err == nil {
			byIRC = &ProductRef{ID: row.ID, Slug: row.Slug}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return persistResult{}, fmt.Errorf("lookup irc: %w", err)
		}
	}
	var byName *ProductRef
	if name := textfa.Normalize(item.NameFa); name != "" {
		row, err := q.GetProductByNormalizedName(ctx, name)
		if err == nil {
			byName = &ProductRef{ID: row.ID, Slug: row.Slug}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return persistResult{}, fmt.Errorf("lookup name %q: %w", name, err)
		}
	}

	dec := DecideMatch(item, byGTIN, byIRC, byName)
	var result persistResult
	var productID int64

	switch dec.Action {
	case ActionLink:
		productID = dec.ProductID
		result.linked = true
		if err := q.LinkSourceItemProduct(ctx, store.LinkSourceItemProductParams{
			ID:        srcRow.ID,
			ProductID: &productID,
		}); err != nil {
			return persistResult{}, fmt.Errorf("link source item: %w", err)
		}
		if err := applySourceFields(ctx, q, productID, item); err != nil {
			return persistResult{}, err
		}
	case ActionCreate:
		id, err := r.createProduct(ctx, q, item)
		if err != nil {
			return persistResult{}, err
		}
		productID = id
		result.created = true
		if err := q.LinkSourceItemProduct(ctx, store.LinkSourceItemProductParams{
			ID:        srcRow.ID,
			ProductID: &productID,
		}); err != nil {
			return persistResult{}, fmt.Errorf("link new product: %w", err)
		}
	case ActionQueue:
		if err := q.InsertMatchCandidate(ctx, store.InsertMatchCandidateParams{
			SourceItemID:       srcRow.ID,
			SuggestedProductID: nil,
			Score:              pgtype.Numeric{},
			Reason:             dec.Reason,
		}); err != nil {
			return persistResult{}, fmt.Errorf("queue match candidate: %w", err)
		}
		result.queued = true
		if err := tx.Commit(ctx); err != nil {
			return persistResult{}, err
		}
		return result, nil
	default:
		return persistResult{}, fmt.Errorf("unknown match action %q", dec.Action)
	}

	changed, err := r.upsertOffer(ctx, q, src.PharmacyID, productID, srcRow.ID, item)
	if err != nil {
		return persistResult{}, err
	}
	result.priceChanged = changed

	if err := tx.Commit(ctx); err != nil {
		return persistResult{}, fmt.Errorf("commit item %s: %w", item.ExternalID, err)
	}
	return result, nil
}

func (r *repository) createProduct(ctx context.Context, q *store.Queries, item RawItem) (int64, error) {
	var brandID *int64
	if item.BrandName != "" {
		id, err := ensureBrand(ctx, q, item.BrandName)
		if err != nil {
			return 0, err
		}
		brandID = &id
	}
	base := productSlug(item)
	slug := uniqueSlug(base, func(s string) bool {
		ok, err := q.SlugExists(ctx, s)
		if err != nil {
			return true
		}
		return ok
	})
	id, err := q.InsertProduct(ctx, store.InsertProductParams{
		Slug:           slug,
		NameFa:         item.NameFa,
		NameEn:         strPtr(item.NameEn),
		ImageUrl:       strPtr(item.ImageURL),
		BrandID:        brandID,
		Gtin:           strPtr(item.GTIN),
		Irc:            strPtr(item.IRC),
		NameNormalized: textfa.Normalize(item.NameFa),
		SearchDocument: textfa.Document(item.NameFa, item.NameEn, item.BrandName),
	})
	if err != nil {
		return 0, fmt.Errorf("insert product %s: %w", item.ExternalID, err)
	}
	if err := q.SetProductSourceSnapshot(ctx, store.SetProductSourceSnapshotParams{
		ID:             id,
		SourceSnapshot: sourceSnapshot(item),
	}); err != nil {
		return 0, fmt.Errorf("set source snapshot: %w", err)
	}
	return id, nil
}

func applySourceFields(ctx context.Context, q *store.Queries, productID int64, item RawItem) error {
	return q.UpdateProductFromSource(ctx, store.UpdateProductFromSourceParams{
		NameFa:         item.NameFa,
		NameEn:         strPtr(item.NameEn),
		ImageUrl:       strPtr(item.ImageURL),
		SourceSnapshot: sourceSnapshot(item),
		NameNormalized: textfa.Normalize(item.NameFa),
		SearchDocument: textfa.Document(item.NameFa, item.NameEn, item.BrandName),
		ID:             productID,
	})
}

func sourceSnapshot(item RawItem) []byte {
	raw, err := json.Marshal(map[string]string{
		"name_fa":    item.NameFa,
		"name_en":    item.NameEn,
		"image_url":  item.ImageURL,
		"brand_name": item.BrandName,
	})
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func ensureBrand(ctx context.Context, q *store.Queries, name string) (int64, error) {
	slug := brandSlug(name)
	row, err := q.GetBrandBySlug(ctx, slug)
	if err == nil {
		return row.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("get brand %s: %w", slug, err)
	}
	id, err := q.InsertBrand(ctx, store.InsertBrandParams{
		Slug:   slug,
		NameFa: name,
		NameEn: strPtr(name),
	})
	if err != nil {
		return 0, fmt.Errorf("insert brand %s: %w", slug, err)
	}
	return id, nil
}

func (r *repository) upsertOffer(ctx context.Context, q *store.Queries, pharmacyID, productID, sourceItemID int64, item RawItem) (priceChanged bool, err error) {
	old, err := q.GetOfferByProductPharmacy(ctx, store.GetOfferByProductPharmacyParams{
		ProductID:  productID,
		PharmacyID: pharmacyID,
	})
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		return false, fmt.Errorf("get offer: %w", err)
	}

	newPrice := item.PriceRial()
	if !missing && IsSuspiciousJump(old.PriceRial, newPrice, r.jumpRatio) {
		if err := q.MarkOfferSuspicious(ctx, store.MarkOfferSuspiciousParams{
			ID:                old.ID,
			ProposedPriceRial: &newPrice,
			InStock:           item.InStock,
			ProductUrl:        item.ProductURL,
			SourceItemID:      &sourceItemID,
		}); err != nil {
			return false, fmt.Errorf("mark offer suspicious: %w", err)
		}
		r.log.WarnContext(ctx, "price.suspicious",
			"offer_id", old.ID,
			"product_id", productID,
			"pharmacy_id", pharmacyID,
			"old_price_rial", old.PriceRial,
			"new_price_rial", newPrice,
		)
		return false, nil
	}

	offer, err := q.UpsertOffer(ctx, store.UpsertOfferParams{
		ProductID:    productID,
		PharmacyID:   pharmacyID,
		SourceItemID: &sourceItemID,
		PriceRial:    newPrice,
		InStock:      item.InStock,
		ProductUrl:   item.ProductURL,
	})
	if err != nil {
		return false, fmt.Errorf("upsert offer: %w", err)
	}

	changed := missing || old.PriceRial != newPrice
	if changed {
		if err := q.InsertPriceHistory(ctx, store.InsertPriceHistoryParams{
			OfferID:    offer.ID,
			ProductID:  productID,
			PharmacyID: pharmacyID,
			PriceRial:  newPrice,
			InStock:    item.InStock,
		}); err != nil {
			return false, fmt.Errorf("insert price history: %w", err)
		}
	}
	return changed, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
