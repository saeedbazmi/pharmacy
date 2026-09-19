package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

// Store is the persistence the ops service needs.
type Store interface {
	GetUserByUsername(ctx context.Context, username string) (userRecord, error)
	GetUserByID(ctx context.Context, id int64) (Actor, error)
	InsertUser(ctx context.Context, username, passwordHash, role string) (int64, error)
	InsertSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error
	SessionUser(ctx context.Context, tokenHash string) (Actor, time.Time, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	InsertLoginAttempt(ctx context.Context, username, ip string) error
	CountRecentLogins(ctx context.Context, username, ip string, since time.Time) (int64, error)
	WriteAudit(ctx context.Context, entry AuditEntry) error
	ListAudit(ctx context.Context, actorID int64, since, until time.Time, limit, offset int32) ([]AuditEntry, int64, error)

	ListPharmacies(ctx context.Context) ([]Pharmacy, error)
	GetPharmacy(ctx context.Context, id int64) (Pharmacy, error)
	InsertPharmacy(ctx context.Context, in PharmacyInput) (int64, error)
	UpdatePharmacy(ctx context.Context, id int64, in PharmacyInput) error
	DisablePharmacy(ctx context.Context, id int64) error

	ListSources(ctx context.Context) ([]Source, error)
	GetSource(ctx context.Context, id int64) (Source, error)
	InsertSource(ctx context.Context, in SourceInput) (int64, error)
	UpdateSource(ctx context.Context, id int64, in SourceInput) error
	DisableSource(ctx context.Context, id int64) error
	ListSourceJobs(ctx context.Context, sourceID int64) ([]JobStatus, error)
	LatestSyncRun(ctx context.Context, sourceID int64) (*SyncRun, error)

	ListMatches(ctx context.Context, sourceID int64, minScore float64, limit, offset int32) ([]Match, int64, error)
	GetMatch(ctx context.Context, id int64) (matchRecord, error)
	ApplyMatch(ctx context.Context, in applyMatchInput) error
	ReopenMatch(ctx context.Context, id int64, previousProductID int64) error

	GetProduct(ctx context.Context, id int64) (Product, error)
	ListProducts(ctx context.Context, status, q string, limit, offset int32) ([]Product, int64, error)
	UpdateProduct(ctx context.Context, p Product) error
	CreateProductFromItem(ctx context.Context, item matchRecord, actorID int64) (int64, error)

	ListCategories(ctx context.Context) ([]Category, error)
	GetCategory(ctx context.Context, id int64) (Category, error)
	InsertCategory(ctx context.Context, c Category) (int64, error)
	UpdateCategory(ctx context.Context, c Category) error
	DeleteCategory(ctx context.Context, id, reassignTo int64) error
	CategoryRedirect(ctx context.Context, oldSlug string) (int64, error)
	CountCategoryProducts(ctx context.Context, id int64) (int64, error)
	CountCategoryChildren(ctx context.Context, id int64) (int64, error)

	ListBrands(ctx context.Context) ([]Brand, error)
	GetBrand(ctx context.Context, id int64) (Brand, error)
	InsertBrand(ctx context.Context, b Brand) (int64, error)
	UpdateBrand(ctx context.Context, b Brand) error
	MergeBrands(ctx context.Context, fromID, intoID int64) error
	CountBrandProducts(ctx context.Context, id int64) (int64, error)

	PendingMatchCount(ctx context.Context) (int64, error)

	ListSourceHealth(ctx context.Context) ([]SourceHealth, error)
	ListSyncSuccessRates(ctx context.Context, since time.Time, sourceID int64) (map[int64]successRate, error)
	ListSourceRuns(ctx context.Context, sourceID int64, limit int32) ([]SyncRun, error)
	ListStaleSources(ctx context.Context, now time.Time) ([]StaleSource, error)
	ListSuspicious(ctx context.Context, limit, offset int32) ([]SuspiciousOffer, int64, error)
	GetSuspicious(ctx context.Context, id int64) (SuspiciousOffer, error)
	ApproveSuspicious(ctx context.Context, id int64) (SuspiciousOffer, error)
	RejectSuspicious(ctx context.Context, id, sourceID int64) error
	ClickReport(ctx context.Context, since, until time.Time) (ClickReport, error)
}

type successRate struct {
	Runs      int32
	Succeeded int32
	Failed    int32
}

type userRecord struct {
	Actor
	PasswordHash string
	Active       bool
}

type matchRecord struct {
	Match
	PharmacyID        int64
	PreviousProductID int64
	LinkedProductID   int64
}

type applyMatchInput struct {
	ID                int64
	SourceItemID      int64
	PharmacyID        int64
	ProductID         int64
	PreviousProductID int64
	ActorID           int64
	Status            string
}

type repository struct {
	pool *pgxpool.Pool
	q    *store.Queries
}

// NewStore is the production Store backed by Postgres.
func NewStore(pool *pgxpool.Pool) Store {
	return &repository{pool: pool, q: store.New(pool)}
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (userRecord, error) {
	row, err := r.q.GetInternalUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userRecord{}, ErrInvalidLogin
		}
		return userRecord{}, fmt.Errorf("get user %s: %w", username, err)
	}
	return userRecord{
		Actor:        Actor{ID: row.ID, Username: row.Username, Role: row.Role},
		PasswordHash: row.PasswordHash,
		Active:       row.IsActive,
	}, nil
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (Actor, error) {
	row, err := r.q.GetInternalUserByID(ctx, id)
	if err != nil {
		return Actor{}, fmt.Errorf("get user %d: %w", id, err)
	}
	return Actor{ID: row.ID, Username: row.Username, Role: row.Role}, nil
}

func (r *repository) InsertUser(ctx context.Context, username, passwordHash, role string) (int64, error) {
	return r.q.InsertInternalUser(ctx, store.InsertInternalUserParams{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
	})
}

func (r *repository) InsertSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error {
	return r.q.InsertSession(ctx, store.InsertSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expires,
	})
}

func (r *repository) SessionUser(ctx context.Context, tokenHash string) (Actor, time.Time, error) {
	row, err := r.q.GetSessionUser(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Actor{}, time.Time{}, ErrUnauthorized
		}
		return Actor{}, time.Time{}, fmt.Errorf("get session: %w", err)
	}
	if !row.IsActive {
		return Actor{}, time.Time{}, ErrInactiveUser
	}
	return Actor{ID: row.ID, Username: row.Username, Role: row.Role}, row.ExpiresAt, nil
}

func (r *repository) DeleteSession(ctx context.Context, tokenHash string) error {
	return r.q.DeleteSession(ctx, tokenHash)
}

func (r *repository) InsertLoginAttempt(ctx context.Context, username, ip string) error {
	return r.q.InsertLoginAttempt(ctx, store.InsertLoginAttemptParams{Username: username, Ip: ip})
}

func (r *repository) CountRecentLogins(ctx context.Context, username, ip string, since time.Time) (int64, error) {
	return r.q.CountRecentLoginAttempts(ctx, store.CountRecentLoginAttemptsParams{
		Since:    since,
		Username: username,
		Ip:       ip,
	})
}

func (r *repository) WriteAudit(ctx context.Context, entry AuditEntry) error {
	var actorID *int64
	if entry.ActorID > 0 {
		actorID = &entry.ActorID
	}
	return r.q.InsertAuditLog(ctx, store.InsertAuditLogParams{
		ActorID:    actorID,
		ActorName:  entry.ActorName,
		Entity:     entry.Entity,
		EntityID:   entry.EntityID,
		Action:     entry.Action,
		BeforeData: nonzeroJSON(entry.Before),
		AfterData:  nonzeroJSON(entry.After),
	})
}

func (r *repository) ListAudit(ctx context.Context, actorID int64, since, until time.Time, limit, offset int32) ([]AuditEntry, int64, error) {
	total, err := r.q.CountAuditLogs(ctx, store.CountAuditLogsParams{ActorID: actorID, Since: since, Until: until})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListAuditLogs(ctx, store.ListAuditLogsParams{
		ActorID: actorID, Since: since, Until: until, PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]AuditEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, AuditEntry{
			ID:        row.ID,
			ActorID:   derefInt(row.ActorID),
			ActorName: row.ActorName,
			Entity:    row.Entity,
			EntityID:  row.EntityID,
			Action:    row.Action,
			Before:    row.BeforeData,
			After:     row.AfterData,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, total, nil
}

func (r *repository) ListPharmacies(ctx context.Context) ([]Pharmacy, error) {
	rows, err := r.q.ListOpsPharmacies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Pharmacy, 0, len(rows))
	for _, row := range rows {
		out = append(out, Pharmacy{
			ID: row.ID, Slug: row.Slug, Name: row.Name, SiteDomain: row.SiteDomain,
			LogoURL: deref(row.LogoUrl), Status: row.Status,
		})
	}
	return out, nil
}

func (r *repository) GetPharmacy(ctx context.Context, id int64) (Pharmacy, error) {
	row, err := r.q.GetOpsPharmacy(ctx, id)
	if err != nil {
		return Pharmacy{}, mapNotFound(err)
	}
	return Pharmacy{
		ID: row.ID, Slug: row.Slug, Name: row.Name, SiteDomain: row.SiteDomain,
		LogoURL: deref(row.LogoUrl), Status: row.Status,
	}, nil
}

func (r *repository) InsertPharmacy(ctx context.Context, in PharmacyInput) (int64, error) {
	return r.q.InsertPharmacy(ctx, store.InsertPharmacyParams{
		Slug: in.Slug, Name: in.Name, SiteDomain: in.SiteDomain, LogoUrl: strPtr(in.LogoURL),
	})
}

func (r *repository) UpdatePharmacy(ctx context.Context, id int64, in PharmacyInput) error {
	return r.q.UpdatePharmacy(ctx, store.UpdatePharmacyParams{
		Name: in.Name, SiteDomain: in.SiteDomain, LogoUrl: strPtr(in.LogoURL), Status: in.Status, ID: id,
	})
}

func (r *repository) DisablePharmacy(ctx context.Context, id int64) error {
	return r.q.DisablePharmacy(ctx, id)
}

func (r *repository) ListSources(ctx context.Context) ([]Source, error) {
	rows, err := r.q.ListOpsSources(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Source, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSource(row.ID, row.PharmacyID, row.PharmacyName, row.PharmacySlug, row.Kind, row.Config, row.ScheduleInterval, row.IsEnabled, row.LastRunAt, row.LastStatus, row.LastError))
	}
	return out, nil
}

func (r *repository) GetSource(ctx context.Context, id int64) (Source, error) {
	row, err := r.q.GetOpsSource(ctx, id)
	if err != nil {
		return Source{}, mapNotFound(err)
	}
	return mapSource(row.ID, row.PharmacyID, row.PharmacyName, row.PharmacySlug, row.Kind, row.Config, row.ScheduleInterval, row.IsEnabled, row.LastRunAt, row.LastStatus, row.LastError), nil
}

func (r *repository) InsertSource(ctx context.Context, in SourceInput) (int64, error) {
	return r.q.InsertDataSource(ctx, store.InsertDataSourceParams{
		PharmacyID: in.PharmacyID, Kind: in.Kind, Config: in.Config, ScheduleInterval: in.ScheduleInterval,
	})
}

func (r *repository) UpdateSource(ctx context.Context, id int64, in SourceInput) error {
	return r.q.UpdateDataSource(ctx, store.UpdateDataSourceParams{
		Kind: in.Kind, Config: in.Config, ScheduleInterval: in.ScheduleInterval, IsEnabled: in.Enabled, ID: id,
	})
}

func (r *repository) DisableSource(ctx context.Context, id int64) error {
	return r.q.DisableDataSource(ctx, id)
}

func (r *repository) ListSourceJobs(ctx context.Context, sourceID int64) ([]JobStatus, error) {
	rows, err := r.q.ListSourceJobs(ctx, strconv.FormatInt(sourceID, 10))
	if err != nil {
		return nil, err
	}
	out := make([]JobStatus, 0, len(rows))
	for _, row := range rows {
		out = append(out, JobStatus{
			JobID: row.ID, Kind: row.Kind, Status: row.Status, Attempts: row.Attempts,
			RunAt: row.RunAt, LastError: deref(row.LastError), CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (r *repository) LatestSyncRun(ctx context.Context, sourceID int64) (*SyncRun, error) {
	row, err := r.q.GetLatestSyncRun(ctx, sourceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &SyncRun{
		ID: row.ID, Status: row.Status, StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		OkCount: row.OkCount, FailCount: row.FailCount, Error: deref(row.Error),
	}, nil
}

func (r *repository) ListMatches(ctx context.Context, sourceID int64, minScore float64, limit, offset int32) ([]Match, int64, error) {
	total, err := r.q.CountPendingMatches(ctx, store.CountPendingMatchesParams{SourceID: sourceID, MinScore: minScore})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListPendingMatches(ctx, store.ListPendingMatchesParams{
		SourceID: sourceID, MinScore: minScore, PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]Match, 0, len(rows))
	for _, row := range rows {
		out = append(out, Match{
			ID: row.ID, SourceItemID: row.SourceItemID, SourceID: row.SourceID, ExternalID: row.ExternalID,
			Score: row.Score, Status: row.Status, Reason: row.Reason, Raw: row.Raw,
			CurrentProductID: derefInt(row.CurrentProductID), SuggestedProductID: derefInt(row.SuggestedProductID),
			SuggestedSlug: deref(row.SuggestedSlug), SuggestedName: deref(row.SuggestedName),
			SuggestedImage: deref(row.SuggestedImage), CreatedAt: row.CreatedAt,
		})
	}
	return out, total, nil
}

func (r *repository) GetMatch(ctx context.Context, id int64) (matchRecord, error) {
	row, err := r.q.GetMatchCandidate(ctx, id)
	if err != nil {
		return matchRecord{}, mapNotFound(err)
	}
	return matchRecord{
		Match: Match{
			ID: row.ID, SourceItemID: row.SourceItemID, SourceID: row.SourceID, ExternalID: row.ExternalID,
			Score: row.Score, Status: row.Status, Reason: row.Reason, Raw: row.Raw,
			CurrentProductID: derefInt(row.CurrentProductID), SuggestedProductID: derefInt(row.SuggestedProductID),
		},
		PharmacyID:        row.PharmacyID,
		PreviousProductID: derefInt(row.PreviousProductID),
		LinkedProductID:   derefInt(row.LinkedProductID),
	}, nil
}

func (r *repository) ApplyMatch(ctx context.Context, in applyMatchInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	if in.Status == "approved" {
		if err := q.LinkSourceItemProduct(ctx, store.LinkSourceItemProductParams{
			ID: in.SourceItemID, ProductID: &in.ProductID,
		}); err != nil {
			return fmt.Errorf("link source item: %w", err)
		}
		if err := relinkOffer(ctx, q, in.SourceItemID, in.PharmacyID, in.ProductID); err != nil {
			return err
		}
	}

	if err := q.DecideMatchCandidate(ctx, store.DecideMatchCandidateParams{
		Status:             in.Status,
		DecidedBy:          &in.ActorID,
		PreviousProductID:  intPtr(in.PreviousProductID),
		LinkedProductID:    intPtr(in.ProductID),
		SuggestedProductID: intPtr(in.ProductID),
		ID:                 in.ID,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *repository) ReopenMatch(ctx context.Context, id int64, previousProductID int64) error {
	row, err := r.q.GetMatchCandidate(ctx, id)
	if err != nil {
		return mapNotFound(err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	var prev *int64
	if previousProductID > 0 {
		prev = &previousProductID
	}
	if err := q.LinkSourceItemProduct(ctx, store.LinkSourceItemProductParams{
		ID: row.SourceItemID, ProductID: prev,
	}); err != nil {
		return err
	}
	if previousProductID > 0 {
		if err := relinkOffer(ctx, q, row.SourceItemID, row.PharmacyID, previousProductID); err != nil {
			return err
		}
	}
	if err := q.ReopenMatchCandidate(ctx, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func relinkOffer(ctx context.Context, q *store.Queries, sourceItemID, pharmacyID, productID int64) error {
	offer, err := q.GetOfferBySourceItem(ctx, &sourceItemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	existing, err := q.GetOfferByProductPharmacy(ctx, store.GetOfferByProductPharmacyParams{
		ProductID: productID, PharmacyID: pharmacyID,
	})
	if err == nil && existing.ID != offer.ID {
		if err := q.DeleteOffer(ctx, offer.ID); err != nil {
			return err
		}
		return nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return q.RelinkOfferProduct(ctx, store.RelinkOfferProductParams{ProductID: productID, ID: offer.ID})
}

func (r *repository) GetProduct(ctx context.Context, id int64) (Product, error) {
	row, err := r.q.GetOpsProduct(ctx, id)
	if err != nil {
		return Product{}, mapNotFound(err)
	}
	return Product{
		ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, NameEn: deref(row.NameEn),
		GenericName: deref(row.GenericName), BrandID: derefInt(row.BrandID), BrandName: deref(row.BrandName),
		CategoryID: derefInt(row.CategoryID), CategoryName: deref(row.CategoryName),
		ImageURL: deref(row.ImageUrl), Status: row.Status, LockedFields: row.LockedFields,
		SourceSnapshot: row.SourceSnapshot,
	}, nil
}

func (r *repository) ListProducts(ctx context.Context, status, q string, limit, offset int32) ([]Product, int64, error) {
	var statusArg, qArg *string
	if status != "" {
		statusArg = &status
	}
	if q != "" {
		qArg = &q
	}
	total, err := r.q.CountOpsProducts(ctx, store.CountOpsProductsParams{Status: statusArg, Q: qArg})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListOpsProducts(ctx, store.ListOpsProductsParams{
		Status: statusArg, Q: qArg, PageLimit: limit, PageOffset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]Product, 0, len(rows))
	for _, row := range rows {
		out = append(out, Product{
			ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, Status: row.Status,
			ImageURL: deref(row.ImageUrl), LockedFields: row.LockedFields,
			BrandName: deref(row.BrandName), CategoryName: deref(row.CategoryName),
		})
	}
	return out, total, nil
}

func (r *repository) UpdateProduct(ctx context.Context, p Product) error {
	return r.q.UpdateOpsProduct(ctx, store.UpdateOpsProductParams{
		NameFa: p.NameFa, NameEn: strPtr(p.NameEn), GenericName: strPtr(p.GenericName),
		BrandID: intPtr(p.BrandID), CategoryID: intPtr(p.CategoryID), ImageUrl: strPtr(p.ImageURL),
		Status: p.Status, LockedFields: p.LockedFields, SourceSnapshot: nonzeroJSON(p.SourceSnapshot),
		NameNormalized: textfa.Normalize(p.NameFa), SearchDocument: textfa.Document(p.NameFa, p.NameEn, p.BrandName),
		ID: p.ID,
	})
}

func (r *repository) CreateProductFromItem(ctx context.Context, item matchRecord, actorID int64) (int64, error) {
	parsed := parseRawItem(item.Raw, item.ExternalID)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	slug := uniqueOpsSlug(parsed.SlugHint, parsed.ExternalID, func(s string) bool {
		ok, err := q.SlugExists(ctx, s)
		return err != nil || ok
	})
	id, err := q.InsertProduct(ctx, store.InsertProductParams{
		Slug: slug, NameFa: parsed.NameFa, NameEn: strPtr(parsed.NameEn), ImageUrl: strPtr(parsed.ImageURL),
		NameNormalized: textfa.Normalize(parsed.NameFa), SearchDocument: textfa.Document(parsed.NameFa, parsed.NameEn, parsed.BrandName),
	})
	if err != nil {
		return 0, err
	}
	snap, _ := json.Marshal(map[string]string{
		"name_fa": parsed.NameFa, "name_en": parsed.NameEn, "image_url": parsed.ImageURL, "brand_name": parsed.BrandName,
	})
	if err := q.SetProductSourceSnapshot(ctx, store.SetProductSourceSnapshotParams{ID: id, SourceSnapshot: snap}); err != nil {
		return 0, err
	}
	if err := q.LinkSourceItemProduct(ctx, store.LinkSourceItemProductParams{ID: item.SourceItemID, ProductID: &id}); err != nil {
		return 0, err
	}
	if parsed.PriceToman > 0 || parsed.ProductURL != "" {
		if _, err := q.UpsertOffer(ctx, store.UpsertOfferParams{
			ProductID: id, PharmacyID: item.PharmacyID, SourceItemID: &item.SourceItemID,
			PriceRial: parsed.PriceToman * 10, InStock: parsed.InStock, ProductUrl: parsed.ProductURL,
		}); err != nil {
			return 0, err
		}
	}
	if err := q.DecideMatchCandidate(ctx, store.DecideMatchCandidateParams{
		Status: "approved", DecidedBy: &actorID, PreviousProductID: intPtr(item.CurrentProductID),
		LinkedProductID: &id, SuggestedProductID: &id, ID: item.ID,
	}); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.q.ListOpsCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, Category{ID: row.ID, ParentID: derefInt(row.ParentID), Slug: row.Slug, NameFa: row.NameFa, Position: row.Position})
	}
	return out, nil
}

func (r *repository) GetCategory(ctx context.Context, id int64) (Category, error) {
	row, err := r.q.GetOpsCategory(ctx, id)
	if err != nil {
		return Category{}, mapNotFound(err)
	}
	return Category{ID: row.ID, ParentID: derefInt(row.ParentID), Slug: row.Slug, NameFa: row.NameFa, Position: row.Position}, nil
}

func (r *repository) InsertCategory(ctx context.Context, c Category) (int64, error) {
	return r.q.InsertOpsCategory(ctx, store.InsertOpsCategoryParams{
		ParentID: intPtr(c.ParentID), Slug: c.Slug, NameFa: c.NameFa, Position: c.Position,
	})
}

func (r *repository) UpdateCategory(ctx context.Context, c Category) error {
	old, err := r.GetCategory(ctx, c.ID)
	if err != nil {
		return err
	}
	if old.Slug != c.Slug {
		if err := r.q.InsertCategorySlugRedirect(ctx, store.InsertCategorySlugRedirectParams{
			OldSlug: old.Slug, CategoryID: c.ID,
		}); err != nil {
			return err
		}
	}
	return r.q.UpdateOpsCategory(ctx, store.UpdateOpsCategoryParams{
		ParentID: intPtr(c.ParentID), Slug: c.Slug, NameFa: c.NameFa, Position: c.Position, ID: c.ID,
	})
}

func (r *repository) DeleteCategory(ctx context.Context, id, reassignTo int64) error {
	if reassignTo > 0 {
		if err := r.q.ReassignCategoryProducts(ctx, store.ReassignCategoryProductsParams{NewID: &reassignTo, OldID: &id}); err != nil {
			return err
		}
	}
	return r.q.SoftDeleteCategory(ctx, id)
}

func (r *repository) CategoryRedirect(ctx context.Context, oldSlug string) (int64, error) {
	row, err := r.q.GetCategoryRedirect(ctx, oldSlug)
	if err != nil {
		return 0, mapNotFound(err)
	}
	return row.CategoryID, nil
}

func (r *repository) CountCategoryProducts(ctx context.Context, id int64) (int64, error) {
	return r.q.CountCategoryProducts(ctx, &id)
}

func (r *repository) CountCategoryChildren(ctx context.Context, id int64) (int64, error) {
	return r.q.CountCategoryChildren(ctx, &id)
}

func (r *repository) ListBrands(ctx context.Context) ([]Brand, error) {
	rows, err := r.q.ListOpsBrands(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Brand, 0, len(rows))
	for _, row := range rows {
		out = append(out, Brand{ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, NameEn: deref(row.NameEn)})
	}
	return out, nil
}

func (r *repository) GetBrand(ctx context.Context, id int64) (Brand, error) {
	row, err := r.q.GetOpsBrand(ctx, id)
	if err != nil {
		return Brand{}, mapNotFound(err)
	}
	return Brand{ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, NameEn: deref(row.NameEn)}, nil
}

func (r *repository) InsertBrand(ctx context.Context, b Brand) (int64, error) {
	return r.q.InsertOpsBrand(ctx, store.InsertOpsBrandParams{Slug: b.Slug, NameFa: b.NameFa, NameEn: strPtr(b.NameEn)})
}

func (r *repository) UpdateBrand(ctx context.Context, b Brand) error {
	return r.q.UpdateOpsBrand(ctx, store.UpdateOpsBrandParams{Slug: b.Slug, NameFa: b.NameFa, NameEn: strPtr(b.NameEn), ID: b.ID})
}

func (r *repository) MergeBrands(ctx context.Context, fromID, intoID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	if err := q.ReassignBrandProducts(ctx, store.ReassignBrandProductsParams{NewID: &intoID, OldID: &fromID}); err != nil {
		return err
	}
	if err := q.SoftDeleteBrand(ctx, fromID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *repository) CountBrandProducts(ctx context.Context, id int64) (int64, error) {
	return r.q.CountBrandProducts(ctx, &id)
}

func (r *repository) PendingMatchCount(ctx context.Context) (int64, error) {
	return r.q.CountMatchPendingKPI(ctx)
}

func mapSource(id, pharmacyID int64, pharmacyName, pharmacySlug, kind string, config []byte, schedule string, enabled bool, lastRun *time.Time, lastStatus, lastError *string) Source {
	return Source{
		ID: id, PharmacyID: pharmacyID, PharmacyName: pharmacyName, PharmacySlug: pharmacySlug,
		Kind: kind, Config: json.RawMessage(config), ScheduleInterval: schedule, Enabled: enabled,
		LastRunAt: lastRun, LastStatus: deref(lastStatus), LastError: deref(lastError),
	}
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func uniqueOpsSlug(base, external string, exists func(string) bool) string {
	if base == "" {
		base = "item-" + external
	}
	if !exists(base) {
		return base
	}
	for i := 2; i < 50; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return fmt.Sprintf("%s-%d", base, 99)
}

type parsedRaw struct {
	ExternalID string
	NameFa     string
	NameEn     string
	BrandName  string
	ImageURL   string
	ProductURL string
	PriceToman int64
	InStock    bool
	SlugHint   string
}

func parseRawItem(raw json.RawMessage, fallbackID string) parsedRaw {
	var loose map[string]any
	_ = json.Unmarshal(raw, &loose)
	out := parsedRaw{ExternalID: fallbackID}
	out.NameFa = firstString(loose, "NameFa", "name_fa", "faName", "name")
	out.NameEn = firstString(loose, "NameEn", "name_en")
	out.BrandName = firstString(loose, "BrandName", "brand_name", "brandEnName")
	out.ImageURL = firstString(loose, "ImageURL", "image_url", "image")
	out.ProductURL = firstString(loose, "ProductURL", "product_url", "productURL")
	out.SlugHint = firstString(loose, "SlugHint", "slug")
	out.InStock = firstBool(loose, "InStock", "in_stock", "storeState")
	out.PriceToman = firstInt(loose, "PriceToman", "price_toman", "discountCost", "orginalCost")
	if out.NameFa == "" {
		out.NameFa = "کالای بدون نام"
	}
	if out.ExternalID == "" {
		out.ExternalID = firstString(loose, "ExternalID", "external_id", "id")
	}
	return out
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			case float64:
				return strconv.FormatInt(int64(t), 10)
			}
		}
	}
	return ""
}

func firstBool(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if v, ok := m[key].(bool); ok {
			return v
		}
	}
	return false
}

func firstInt(m map[string]any, keys ...string) int64 {
	for _, key := range keys {
		switch v := m[key].(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case json.Number:
			n, _ := v.Int64()
			return n
		}
	}
	return 0
}
