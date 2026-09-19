package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/cache"
)

// Service holds launch-admin rules: dashboard, users, settings and flags.
type Service struct {
	store    Store
	log      *slog.Logger
	now      func() time.Time
	flags    *cache.TTL[[]Flag]
	settings *cache.TTL[Settings]
}

func NewService(store Store, log *slog.Logger) *Service {
	return &Service{
		store:    store,
		log:      log,
		now:      time.Now,
		flags:    cache.NewTTL[[]Flag](30*time.Second, 8),
		settings: cache.NewTTL[Settings](30*time.Second, 8),
	}
}

func parseRange(days int, now time.Time) (from, to time.Time, n int) {
	if days < 1 {
		days = defaultRangeDays
	}
	if days > maxRangeDays {
		days = maxRangeDays
	}
	to = now.UTC()
	from = to.Add(-time.Duration(days) * 24 * time.Hour)
	return from, to, days
}

func clickThroughRate(clicks, searches int64) float64 {
	if searches <= 0 {
		return 0
	}
	return float64(clicks) / float64(searches)
}

func mergeSeries(searches, clicks []DayPoint) []DayPoint {
	byDay := map[int64]*DayPoint{}
	order := make([]int64, 0)
	add := func(day time.Time) *DayPoint {
		key := day.UTC().Truncate(24 * time.Hour).Unix()
		if p, ok := byDay[key]; ok {
			return p
		}
		p := &DayPoint{Day: time.Unix(key, 0).UTC()}
		byDay[key] = p
		order = append(order, key)
		return p
	}
	for _, p := range searches {
		add(p.Day).Searches = p.Searches
	}
	for _, p := range clicks {
		add(p.Day).Clicks = p.Clicks
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	out := make([]DayPoint, 0, len(order))
	for _, key := range order {
		out = append(out, *byDay[key])
	}
	return out
}

func resolveFlag(rows []Flag, key string, pharmacyID int64) bool {
	var global bool
	var override *bool
	for _, f := range rows {
		if f.Key != key {
			continue
		}
		if f.PharmacyID == 0 {
			global = f.Enabled
			continue
		}
		if pharmacyID > 0 && f.PharmacyID == pharmacyID {
			v := f.Enabled
			override = &v
		}
	}
	if override != nil {
		return *override
	}
	return global
}

func (s *Service) Dashboard(ctx context.Context, days int) (Dashboard, error) {
	from, to, days := parseRange(days, s.now())
	cov, err := s.store.Coverage(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	searches, err := s.store.CountSearches(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	clicks, err := s.store.CountClicks(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	pharmacies, err := s.store.ClicksByPharmacy(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	frequent, err := s.store.FrequentSearches(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	zero, err := s.store.ZeroSearches(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	searchDays, err := s.store.SearchVolumeByDay(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	clickDays, err := s.store.ClickVolumeByDay(ctx, from, to)
	if err != nil {
		return Dashboard{}, err
	}
	return Dashboard{
		From: from, To: to, Days: days,
		Searches: searches, Clicks: clicks, CTR: clickThroughRate(clicks, searches),
		Coverage: cov, Pharmacies: pharmacies, Frequent: frequent, Zero: zero,
		Series: mergeSeries(searchDays, clickDays),
	}, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]InternalUser, error) {
	return s.store.ListUsers(ctx)
}

func (s *Service) CreateUser(ctx context.Context, actor Actor, username, password, role string) (InternalUser, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if !validUsername(username) {
		return InternalUser{}, ErrInvalidInput
	}
	if len(password) < minPasswordLen {
		return InternalUser{}, ErrInvalidInput
	}
	if role != RoleDataOps && role != RoleSuperAdmin {
		return InternalUser{}, ErrInvalidInput
	}
	exists, err := s.store.UserExists(ctx, username)
	if err != nil {
		return InternalUser{}, err
	}
	if exists {
		return InternalUser{}, ErrConflict
	}
	hash, err := ops.HashPassword(password)
	if err != nil {
		return InternalUser{}, fmt.Errorf("hash password: %w", err)
	}
	id, err := s.store.InsertUser(ctx, username, hash, role)
	if err != nil {
		return InternalUser{}, err
	}
	got, err := s.store.GetUser(ctx, id)
	if err != nil {
		return InternalUser{}, err
	}
	s.audit(ctx, actor, "internal_user", fmt.Sprint(id), "create", nil, got)
	return got, nil
}

func (s *Service) UpdateUser(ctx context.Context, actor Actor, id int64, patch UserPatch) (InternalUser, error) {
	if id <= 0 {
		return InternalUser{}, ErrInvalidInput
	}
	if patch.Role != RoleDataOps && patch.Role != RoleSuperAdmin {
		return InternalUser{}, ErrInvalidInput
	}
	if patch.Password != "" && len(patch.Password) < minPasswordLen {
		return InternalUser{}, ErrInvalidInput
	}
	before, err := s.store.GetUser(ctx, id)
	if err != nil {
		return InternalUser{}, err
	}
	active := before.IsActive
	if patch.IsActive != nil {
		active = *patch.IsActive
	}
	if losesSuperAdmin(before, patch.Role, active) {
		n, err := s.store.CountActiveSuperAdmins(ctx)
		if err != nil {
			return InternalUser{}, err
		}
		if n <= 1 {
			return InternalUser{}, ErrLastSuperAdmin
		}
	}
	var hash *string
	if patch.Password != "" {
		encoded, err := ops.HashPassword(patch.Password)
		if err != nil {
			return InternalUser{}, fmt.Errorf("hash password: %w", err)
		}
		hash = &encoded
	}
	if err := s.store.UpdateUser(ctx, id, patch.Role, active, hash); err != nil {
		return InternalUser{}, err
	}
	got, err := s.store.GetUser(ctx, id)
	if err != nil {
		return InternalUser{}, err
	}
	s.audit(ctx, actor, "internal_user", fmt.Sprint(id), "update", before, got)
	return got, nil
}

func losesSuperAdmin(before InternalUser, role string, active bool) bool {
	if before.Role != RoleSuperAdmin || !before.IsActive {
		return false
	}
	return role != RoleSuperAdmin || !active
}

func validUsername(s string) bool {
	if len(s) < 3 || len(s) > 32 {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func (s *Service) PublicSite(ctx context.Context) (Settings, error) {
	return s.settingsFromCache(ctx)
}

func (s *Service) GetSettings(ctx context.Context) (Settings, error) {
	return s.settingsFromCache(ctx)
}

func (s *Service) settingsFromCache(ctx context.Context) (Settings, error) {
	if got, ok := s.settings.Get(settingsCacheKey); ok {
		return got, nil
	}
	got, err := s.loadSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	s.settings.Set(settingsCacheKey, got)
	return got, nil
}

func (s *Service) loadSettings(ctx context.Context) (Settings, error) {
	got := defaultSettings()
	rows, err := s.store.ListSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	if raw, ok := rows[settingSEO]; ok {
		_ = json.Unmarshal(raw, &got.SEO)
	}
	if raw, ok := rows[settingBanner]; ok {
		_ = json.Unmarshal(raw, &got.Banner)
	}
	if raw, ok := rows[settingFeatured]; ok {
		_ = json.Unmarshal(raw, &got.Featured)
	}
	if raw, ok := rows[settingPages]; ok {
		_ = json.Unmarshal(raw, &got.Pages)
	}
	if got.Featured.Slugs == nil {
		got.Featured.Slugs = []string{}
	}
	return got, nil
}

func (s *Service) UpdateSettings(ctx context.Context, actor Actor, next Settings) (Settings, error) {
	next = normalizeSettings(next)
	before, err := s.loadSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	if err := s.store.UpsertSetting(ctx, settingSEO, mustJSON(next.SEO)); err != nil {
		return Settings{}, err
	}
	if err := s.store.UpsertSetting(ctx, settingBanner, mustJSON(next.Banner)); err != nil {
		return Settings{}, err
	}
	if err := s.store.UpsertSetting(ctx, settingFeatured, mustJSON(next.Featured)); err != nil {
		return Settings{}, err
	}
	if err := s.store.UpsertSetting(ctx, settingPages, mustJSON(next.Pages)); err != nil {
		return Settings{}, err
	}
	s.settings.Delete(settingsCacheKey)
	s.audit(ctx, actor, "site_settings", "all", "update", before, next)
	return next, nil
}

func normalizeSettings(in Settings) Settings {
	in.SEO.Title = strings.TrimSpace(in.SEO.Title)
	in.SEO.Description = strings.TrimSpace(in.SEO.Description)
	if in.SEO.Title == "" {
		in.SEO.Title = defaultSettings().SEO.Title
	}
	in.Banner.Title = strings.TrimSpace(in.Banner.Title)
	in.Banner.Body = strings.TrimSpace(in.Banner.Body)
	clean := make([]string, 0, len(in.Featured.Slugs))
	seen := map[string]bool{}
	for _, slug := range in.Featured.Slugs {
		slug = strings.ToLower(strings.TrimSpace(slug))
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		clean = append(clean, slug)
	}
	in.Featured.Slugs = clean
	in.Pages.About = strings.TrimSpace(in.Pages.About)
	in.Pages.Contact = strings.TrimSpace(in.Pages.Contact)
	in.Pages.Terms = strings.TrimSpace(in.Pages.Terms)
	in.Pages.Disclaimer = strings.TrimSpace(in.Pages.Disclaimer)
	return in
}

func (s *Service) ListFlags(ctx context.Context) ([]Flag, []PharmacyBrief, error) {
	flags, err := s.flagsFromCache(ctx)
	if err != nil {
		return nil, nil, err
	}
	pharmacies, err := s.store.ListPharmacies(ctx)
	if err != nil {
		return nil, nil, err
	}
	return flags, pharmacies, nil
}

func (s *Service) SetFlag(ctx context.Context, actor Actor, key string, pharmacyID int64, enabled bool) error {
	if key != FlagDirectPurchase {
		return ErrInvalidInput
	}
	if pharmacyID < 0 {
		return ErrInvalidInput
	}
	before, err := s.flagsFromCache(ctx)
	if err != nil {
		return err
	}
	if pharmacyID == 0 {
		if err := s.store.UpsertGlobalFlag(ctx, key, enabled); err != nil {
			return err
		}
	} else {
		if err := s.store.UpsertPharmacyFlag(ctx, key, pharmacyID, enabled); err != nil {
			return err
		}
	}
	s.flags.Delete(flagCacheKey)
	after := Flag{Key: key, PharmacyID: pharmacyID, Enabled: enabled}
	s.audit(ctx, actor, "feature_flag", key, "update", before, after)
	s.log.InfoContext(ctx, "flag.updated", "key", key, "pharmacy_id", pharmacyID, "enabled", enabled)
	return nil
}

// FlagEnabled reports a cached flag. Public pages must not branch on this in phase 1.
func (s *Service) FlagEnabled(ctx context.Context, key string, pharmacyID int64) (bool, error) {
	rows, err := s.flagsFromCache(ctx)
	if err != nil {
		return false, err
	}
	return resolveFlag(rows, key, pharmacyID), nil
}

func (s *Service) flagsFromCache(ctx context.Context) ([]Flag, error) {
	if got, ok := s.flags.Get(flagCacheKey); ok {
		return got, nil
	}
	rows, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, err
	}
	s.flags.Set(flagCacheKey, rows)
	return rows, nil
}

func (s *Service) audit(ctx context.Context, actor Actor, entity, entityID, action string, before, after any) {
	if err := s.store.WriteAudit(ctx, actor, entity, entityID, action, before, after); err != nil {
		s.log.ErrorContext(ctx, "audit.failed", "entity", entity, "action", action, "error", err.Error())
	}
}
