package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

// Store is the persistence the admin service needs.
type Store interface {
	Coverage(ctx context.Context) (Coverage, error)
	CountClicks(ctx context.Context, from, to time.Time) (int64, error)
	CountSearches(ctx context.Context, from, to time.Time) (int64, error)
	ClicksByPharmacy(ctx context.Context, from, to time.Time) ([]PharmacyClicks, error)
	SearchVolumeByDay(ctx context.Context, from, to time.Time) ([]DayPoint, error)
	ClickVolumeByDay(ctx context.Context, from, to time.Time) ([]DayPoint, error)
	FrequentSearches(ctx context.Context, from, to time.Time) ([]SearchTerm, error)
	ZeroSearches(ctx context.Context, from, to time.Time) ([]SearchTerm, error)

	ListUsers(ctx context.Context) ([]InternalUser, error)
	GetUser(ctx context.Context, id int64) (InternalUser, error)
	UserExists(ctx context.Context, username string) (bool, error)
	InsertUser(ctx context.Context, username, passwordHash, role string) (int64, error)
	UpdateUser(ctx context.Context, id int64, role string, active bool, passwordHash *string) error
	CountActiveSuperAdmins(ctx context.Context) (int64, error)

	ListSettings(ctx context.Context) (map[string]json.RawMessage, error)
	UpsertSetting(ctx context.Context, key string, value json.RawMessage) error

	ListFlags(ctx context.Context) ([]Flag, error)
	UpsertGlobalFlag(ctx context.Context, key string, enabled bool) error
	UpsertPharmacyFlag(ctx context.Context, key string, pharmacyID int64, enabled bool) error
	ListPharmacies(ctx context.Context) ([]PharmacyBrief, error)

	WriteAudit(ctx context.Context, actor Actor, entity, entityID, action string, before, after any) error
}

type repository struct {
	q *store.Queries
}

// NewStore wires admin queries to a pool.
func NewStore(db *pgxpool.Pool) Store {
	return &repository{q: store.New(db)}
}

func (r *repository) Coverage(ctx context.Context) (Coverage, error) {
	row, err := r.q.DashboardCoverage(ctx)
	if err != nil {
		return Coverage{}, fmt.Errorf("dashboard coverage: %w", err)
	}
	return Coverage{
		ActivePharmacies:       row.ActivePharmacies,
		PublishedProducts:      row.PublishedProducts,
		MedianFreshnessSeconds: row.MedianFreshnessSeconds,
	}, nil
}

func (r *repository) CountClicks(ctx context.Context, from, to time.Time) (int64, error) {
	n, err := r.q.CountClicksInRange(ctx, store.CountClicksInRangeParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return 0, fmt.Errorf("count clicks: %w", err)
	}
	return n, nil
}

func (r *repository) CountSearches(ctx context.Context, from, to time.Time) (int64, error) {
	n, err := r.q.CountSearchesInRange(ctx, store.CountSearchesInRangeParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return 0, fmt.Errorf("count searches: %w", err)
	}
	return n, nil
}

func (r *repository) ClicksByPharmacy(ctx context.Context, from, to time.Time) ([]PharmacyClicks, error) {
	rows, err := r.q.ListClicksByPharmacy(ctx, store.ListClicksByPharmacyParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return nil, fmt.Errorf("clicks by pharmacy: %w", err)
	}
	out := make([]PharmacyClicks, 0, len(rows))
	for _, row := range rows {
		out = append(out, PharmacyClicks{ID: row.ID, Slug: row.Slug, Name: row.Name, Clicks: row.Clicks})
	}
	return out, nil
}

func (r *repository) SearchVolumeByDay(ctx context.Context, from, to time.Time) ([]DayPoint, error) {
	rows, err := r.q.ListSearchVolumeByDay(ctx, store.ListSearchVolumeByDayParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return nil, fmt.Errorf("search volume: %w", err)
	}
	out := make([]DayPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, DayPoint{Day: row.Day, Searches: row.Searches})
	}
	return out, nil
}

func (r *repository) ClickVolumeByDay(ctx context.Context, from, to time.Time) ([]DayPoint, error) {
	rows, err := r.q.ListClickVolumeByDay(ctx, store.ListClickVolumeByDayParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return nil, fmt.Errorf("click volume: %w", err)
	}
	out := make([]DayPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, DayPoint{Day: row.Day, Clicks: row.Clicks})
	}
	return out, nil
}

func (r *repository) FrequentSearches(ctx context.Context, from, to time.Time) ([]SearchTerm, error) {
	rows, err := r.q.ListFrequentSearchesInRange(ctx, store.ListFrequentSearchesInRangeParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return nil, fmt.Errorf("frequent searches: %w", err)
	}
	out := make([]SearchTerm, 0, len(rows))
	for _, row := range rows {
		out = append(out, SearchTerm{Query: row.QueryNormalized, Hits: row.Hits, LastSeen: asTime(row.LastSeen)})
	}
	return out, nil
}

func (r *repository) ZeroSearches(ctx context.Context, from, to time.Time) ([]SearchTerm, error) {
	rows, err := r.q.ListZeroResultSearchesInRange(ctx, store.ListZeroResultSearchesInRangeParams{RangeStart: from, RangeEnd: to})
	if err != nil {
		return nil, fmt.Errorf("zero searches: %w", err)
	}
	out := make([]SearchTerm, 0, len(rows))
	for _, row := range rows {
		out = append(out, SearchTerm{Query: row.QueryNormalized, Hits: row.Hits, LastSeen: asTime(row.LastSeen)})
	}
	return out, nil
}

func (r *repository) ListUsers(ctx context.Context) ([]InternalUser, error) {
	rows, err := r.q.ListInternalUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list internal users: %w", err)
	}
	out := make([]InternalUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, InternalUser{ID: row.ID, Username: row.Username, Role: row.Role, IsActive: row.IsActive, CreatedAt: row.CreatedAt})
	}
	return out, nil
}

func (r *repository) GetUser(ctx context.Context, id int64) (InternalUser, error) {
	row, err := r.q.GetInternalUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InternalUser{}, ErrNotFound
		}
		return InternalUser{}, fmt.Errorf("get internal user %d: %w", id, err)
	}
	return InternalUser{ID: row.ID, Username: row.Username, Role: row.Role, IsActive: row.IsActive}, nil
}

func (r *repository) UserExists(ctx context.Context, username string) (bool, error) {
	_, err := r.q.GetInternalUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lookup username: %w", err)
	}
	return true, nil
}

func (r *repository) InsertUser(ctx context.Context, username, passwordHash, role string) (int64, error) {
	id, err := r.q.InsertInternalUser(ctx, store.InsertInternalUserParams{
		Username: username, PasswordHash: passwordHash, Role: role,
	})
	if err != nil {
		return 0, fmt.Errorf("insert internal user: %w", err)
	}
	return id, nil
}

func (r *repository) UpdateUser(ctx context.Context, id int64, role string, active bool, passwordHash *string) error {
	if err := r.q.UpdateInternalUser(ctx, store.UpdateInternalUserParams{
		ID: id, Role: role, IsActive: active, PasswordHash: passwordHash,
	}); err != nil {
		return fmt.Errorf("update internal user %d: %w", id, err)
	}
	return nil
}

func (r *repository) CountActiveSuperAdmins(ctx context.Context) (int64, error) {
	n, err := r.q.CountActiveSuperAdmins(ctx)
	if err != nil {
		return 0, fmt.Errorf("count super admins: %w", err)
	}
	return n, nil
}

func (r *repository) ListSettings(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := r.q.ListSiteSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("list site settings: %w", err)
	}
	out := make(map[string]json.RawMessage, len(rows))
	for _, row := range rows {
		out[row.Key] = json.RawMessage(row.Value)
	}
	return out, nil
}

func (r *repository) UpsertSetting(ctx context.Context, key string, value json.RawMessage) error {
	if err := r.q.UpsertSiteSetting(ctx, store.UpsertSiteSettingParams{Key: key, Value: value}); err != nil {
		return fmt.Errorf("upsert setting %s: %w", key, err)
	}
	return nil
}

func (r *repository) ListFlags(ctx context.Context) ([]Flag, error) {
	rows, err := r.q.ListFeatureFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("list flags: %w", err)
	}
	out := make([]Flag, 0, len(rows))
	for _, row := range rows {
		f := Flag{Key: row.Key, Enabled: row.Enabled, UpdatedAt: row.UpdatedAt}
		if row.PharmacyID != nil {
			f.PharmacyID = *row.PharmacyID
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *repository) UpsertGlobalFlag(ctx context.Context, key string, enabled bool) error {
	if err := r.q.UpsertGlobalFlag(ctx, store.UpsertGlobalFlagParams{Key: key, Enabled: enabled}); err != nil {
		return fmt.Errorf("upsert global flag %s: %w", key, err)
	}
	return nil
}

func (r *repository) UpsertPharmacyFlag(ctx context.Context, key string, pharmacyID int64, enabled bool) error {
	id := pharmacyID
	if err := r.q.UpsertPharmacyFlag(ctx, store.UpsertPharmacyFlagParams{
		Key: key, PharmacyID: &id, Enabled: enabled,
	}); err != nil {
		return fmt.Errorf("upsert pharmacy flag %s/%d: %w", key, pharmacyID, err)
	}
	return nil
}

func (r *repository) ListPharmacies(ctx context.Context) ([]PharmacyBrief, error) {
	rows, err := r.q.ListActivePharmaciesBrief(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pharmacies: %w", err)
	}
	out := make([]PharmacyBrief, 0, len(rows))
	for _, row := range rows {
		out = append(out, PharmacyBrief{ID: row.ID, Slug: row.Slug, Name: row.Name})
	}
	return out, nil
}

func (r *repository) WriteAudit(ctx context.Context, actor Actor, entity, entityID, action string, before, after any) error {
	var actorID *int64
	if actor.ID > 0 {
		actorID = &actor.ID
	}
	return r.q.InsertAuditLog(ctx, store.InsertAuditLogParams{
		ActorID:    actorID,
		ActorName:  actor.Username,
		Entity:     entity,
		EntityID:   entityID,
		Action:     action,
		BeforeData: mustJSON(before),
		AfterData:  mustJSON(after),
	})
}

func asTime(v any) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case *time.Time:
		if t != nil {
			return *t
		}
	}
	return time.Time{}
}
