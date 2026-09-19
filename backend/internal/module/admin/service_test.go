package admin

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

type memStore struct {
	users      map[int64]InternalUser
	nextUser   int64
	superCount int64
	flags      []Flag
	flagLoads  int
	settings   map[string]json.RawMessage
	audits     int
	searches   int64
	clicks     int64
}

func newMem() *memStore {
	return &memStore{
		users:      map[int64]InternalUser{},
		nextUser:   1,
		superCount: 1,
		flags:      []Flag{{Key: FlagDirectPurchase, Enabled: false}},
		settings:   map[string]json.RawMessage{},
	}
}

func (m *memStore) Coverage(context.Context) (Coverage, error) {
	return Coverage{ActivePharmacies: 2, PublishedProducts: 10, MedianFreshnessSeconds: 3600}, nil
}
func (m *memStore) CountClicks(context.Context, time.Time, time.Time) (int64, error) {
	return m.clicks, nil
}
func (m *memStore) CountSearches(context.Context, time.Time, time.Time) (int64, error) {
	return m.searches, nil
}
func (m *memStore) ClicksByPharmacy(context.Context, time.Time, time.Time) ([]PharmacyClicks, error) {
	return nil, nil
}
func (m *memStore) SearchVolumeByDay(context.Context, time.Time, time.Time) ([]DayPoint, error) {
	return nil, nil
}
func (m *memStore) ClickVolumeByDay(context.Context, time.Time, time.Time) ([]DayPoint, error) {
	return nil, nil
}
func (m *memStore) FrequentSearches(context.Context, time.Time, time.Time) ([]SearchTerm, error) {
	return nil, nil
}
func (m *memStore) ZeroSearches(context.Context, time.Time, time.Time) ([]SearchTerm, error) {
	return nil, nil
}
func (m *memStore) ListUsers(context.Context) ([]InternalUser, error) {
	out := make([]InternalUser, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, u)
	}
	return out, nil
}
func (m *memStore) GetUser(_ context.Context, id int64) (InternalUser, error) {
	u, ok := m.users[id]
	if !ok {
		return InternalUser{}, ErrNotFound
	}
	return u, nil
}
func (m *memStore) UserExists(_ context.Context, username string) (bool, error) {
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}
func (m *memStore) InsertUser(_ context.Context, username, _, role string) (int64, error) {
	id := m.nextUser
	m.nextUser++
	m.users[id] = InternalUser{ID: id, Username: username, Role: role, IsActive: true}
	if role == RoleSuperAdmin {
		m.superCount++
	}
	return id, nil
}
func (m *memStore) UpdateUser(_ context.Context, id int64, role string, active bool, _ *string) error {
	u, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	if u.Role == RoleSuperAdmin && u.IsActive && (role != RoleSuperAdmin || !active) {
		m.superCount--
	}
	if u.Role != RoleSuperAdmin && role == RoleSuperAdmin && active {
		m.superCount++
	}
	u.Role = role
	u.IsActive = active
	m.users[id] = u
	return nil
}
func (m *memStore) CountActiveSuperAdmins(context.Context) (int64, error) {
	return m.superCount, nil
}
func (m *memStore) ListSettings(context.Context) (map[string]json.RawMessage, error) {
	return m.settings, nil
}
func (m *memStore) UpsertSetting(_ context.Context, key string, value json.RawMessage) error {
	m.settings[key] = value
	return nil
}
func (m *memStore) ListFlags(context.Context) ([]Flag, error) {
	m.flagLoads++
	out := append([]Flag{}, m.flags...)
	return out, nil
}
func (m *memStore) UpsertGlobalFlag(_ context.Context, key string, enabled bool) error {
	for i, f := range m.flags {
		if f.Key == key && f.PharmacyID == 0 {
			m.flags[i].Enabled = enabled
			return nil
		}
	}
	m.flags = append(m.flags, Flag{Key: key, Enabled: enabled})
	return nil
}
func (m *memStore) UpsertPharmacyFlag(_ context.Context, key string, pharmacyID int64, enabled bool) error {
	for i, f := range m.flags {
		if f.Key == key && f.PharmacyID == pharmacyID {
			m.flags[i].Enabled = enabled
			return nil
		}
	}
	m.flags = append(m.flags, Flag{Key: key, PharmacyID: pharmacyID, Enabled: enabled})
	return nil
}
func (m *memStore) ListPharmacies(context.Context) ([]PharmacyBrief, error) {
	return []PharmacyBrief{{ID: 1, Slug: "darukade", Name: "داروکده"}}, nil
}
func (m *memStore) WriteAudit(context.Context, Actor, string, string, string, any, any) error {
	m.audits++
	return nil
}

func testService(store *memStore) *Service {
	return NewService(store, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func TestParseRangeClamps(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	from, to, days := parseRange(0, now)
	if days != 7 {
		t.Fatalf("days = %d, want 7", days)
	}
	if to != now || !from.Equal(now.Add(-7*24*time.Hour)) {
		t.Fatalf("range = %s .. %s", from, to)
	}
	_, _, days = parseRange(900, now)
	if days != 366 {
		t.Fatalf("days = %d, want 366", days)
	}
}

func TestClickThroughRateZeroSearches(t *testing.T) {
	if got := clickThroughRate(12, 0); got != 0 {
		t.Fatalf("ctr = %v, want 0", got)
	}
	if got := clickThroughRate(1, 4); got != 0.25 {
		t.Fatalf("ctr = %v, want 0.25", got)
	}
}

func TestLastSuperAdminCannotDemoteOrDisableSelf(t *testing.T) {
	store := newMem()
	store.users[1] = InternalUser{ID: 1, Username: "root", Role: RoleSuperAdmin, IsActive: true}
	store.superCount = 1
	svc := testService(store)
	actor := Actor{ID: 1, Username: "root", Role: RoleSuperAdmin}

	_, err := svc.UpdateUser(context.Background(), actor, 1, UserPatch{Role: RoleDataOps})
	if err != ErrLastSuperAdmin {
		t.Fatalf("demote last admin: %v", err)
	}
	off := false
	_, err = svc.UpdateUser(context.Background(), actor, 1, UserPatch{Role: RoleSuperAdmin, IsActive: &off})
	if err != ErrLastSuperAdmin {
		t.Fatalf("disable last admin: %v", err)
	}
}

func TestSecondSuperAdminCanBeDemoted(t *testing.T) {
	store := newMem()
	store.users[1] = InternalUser{ID: 1, Username: "root", Role: RoleSuperAdmin, IsActive: true}
	store.users[2] = InternalUser{ID: 2, Username: "other", Role: RoleSuperAdmin, IsActive: true}
	store.superCount = 2
	svc := testService(store)

	got, err := svc.UpdateUser(context.Background(), Actor{ID: 1, Username: "root", Role: RoleSuperAdmin}, 2, UserPatch{
		Role: RoleDataOps,
	})
	if err != nil {
		t.Fatalf("demote second admin: %v", err)
	}
	if got.Role != RoleDataOps {
		t.Fatalf("role = %s", got.Role)
	}
	if store.audits == 0 {
		t.Fatal("expected audit write")
	}
}

func TestFlagCacheDoesNotHitStoreTwice(t *testing.T) {
	store := newMem()
	svc := testService(store)
	ctx := context.Background()

	on, err := svc.FlagEnabled(ctx, FlagDirectPurchase, 0)
	if err != nil || on {
		t.Fatalf("default flag = %v %v, want false", on, err)
	}
	if store.flagLoads != 1 {
		t.Fatalf("loads = %d, want 1", store.flagLoads)
	}
	on, err = svc.FlagEnabled(ctx, FlagDirectPurchase, 0)
	if err != nil || on {
		t.Fatalf("cached flag = %v %v", on, err)
	}
	if store.flagLoads != 1 {
		t.Fatalf("second read hit store (%d)", store.flagLoads)
	}
}

func TestPharmacyFlagOverridesGlobal(t *testing.T) {
	rows := []Flag{
		{Key: FlagDirectPurchase, Enabled: false},
		{Key: FlagDirectPurchase, PharmacyID: 7, Enabled: true},
	}
	if !resolveFlag(rows, FlagDirectPurchase, 7) {
		t.Fatal("pharmacy override should win")
	}
	if resolveFlag(rows, FlagDirectPurchase, 3) {
		t.Fatal("other pharmacy should see global off")
	}
}

func TestSetFlagInvalidatesCacheAndRejectsUnknownKey(t *testing.T) {
	store := newMem()
	svc := testService(store)
	ctx := context.Background()
	actor := Actor{ID: 1, Username: "root", Role: RoleSuperAdmin}

	if _, err := svc.FlagEnabled(ctx, FlagDirectPurchase, 0); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetFlag(ctx, actor, "checkout", 0, true); err != ErrInvalidInput {
		t.Fatalf("unknown key: %v", err)
	}
	if err := svc.SetFlag(ctx, actor, FlagDirectPurchase, 0, true); err != nil {
		t.Fatal(err)
	}
	on, err := svc.FlagEnabled(ctx, FlagDirectPurchase, 0)
	if err != nil || !on {
		t.Fatalf("after set = %v %v", on, err)
	}
	if store.flagLoads < 2 {
		t.Fatalf("cache was not invalidated, loads=%d", store.flagLoads)
	}
}

func TestUpdateSettingsInvalidatesCache(t *testing.T) {
	store := newMem()
	svc := testService(store)
	ctx := context.Background()
	first, err := svc.PublicSite(ctx)
	if err != nil {
		t.Fatal(err)
	}
	next := first
	next.SEO.Title = "عنوان تازه"
	if _, err := svc.UpdateSettings(ctx, Actor{ID: 1, Username: "root"}, next); err != nil {
		t.Fatal(err)
	}
	got, err := svc.PublicSite(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.SEO.Title != "عنوان تازه" {
		t.Fatalf("title = %q", got.SEO.Title)
	}
	if store.audits == 0 {
		t.Fatal("settings change must be audited")
	}
}

func TestDashboardCTR(t *testing.T) {
	store := newMem()
	store.searches = 20
	store.clicks = 5
	got, err := testService(store).Dashboard(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if got.CTR != 0.25 {
		t.Fatalf("ctr = %v", got.CTR)
	}
	if got.Coverage.ActivePharmacies != 2 {
		t.Fatalf("pharmacies = %d", got.Coverage.ActivePharmacies)
	}
}
