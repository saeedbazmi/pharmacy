package ops

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

type memStore struct {
	users    map[string]userRecord
	sessions map[string]Actor
	attempts int64
	audits   []AuditEntry
	sources  map[int64]Source
	matches  map[int64]matchRecord
	products map[int64]Product
	nextID   int64
	applied  []applyMatchInput
}

func newMem() *memStore {
	return &memStore{
		users:    map[string]userRecord{},
		sessions: map[string]Actor{},
		sources:  map[int64]Source{},
		matches:  map[int64]matchRecord{},
		products: map[int64]Product{},
		nextID:   1,
	}
}

func (m *memStore) GetUserByUsername(_ context.Context, username string) (userRecord, error) {
	u, ok := m.users[username]
	if !ok {
		return userRecord{}, ErrInvalidLogin
	}
	return u, nil
}
func (m *memStore) GetUserByID(context.Context, int64) (Actor, error) { return Actor{}, ErrNotFound }
func (m *memStore) InsertUser(_ context.Context, username, hash, role string) (int64, error) {
	id := m.nextID
	m.nextID++
	m.users[username] = userRecord{Actor: Actor{ID: id, Username: username, Role: role}, PasswordHash: hash, Active: true}
	return id, nil
}
func (m *memStore) InsertSession(_ context.Context, _ int64, tokenHash string, _ time.Time) error {
	return nil
}
func (m *memStore) SessionUser(_ context.Context, tokenHash string) (Actor, time.Time, error) {
	a, ok := m.sessions[tokenHash]
	if !ok {
		return Actor{}, time.Time{}, ErrUnauthorized
	}
	return a, time.Now().Add(time.Hour), nil
}
func (m *memStore) DeleteSession(context.Context, string) error { return nil }
func (m *memStore) InsertLoginAttempt(context.Context, string, string) error {
	m.attempts++
	return nil
}
func (m *memStore) CountRecentLogins(context.Context, string, string, time.Time) (int64, error) {
	return m.attempts, nil
}
func (m *memStore) WriteAudit(_ context.Context, entry AuditEntry) error {
	m.audits = append(m.audits, entry)
	return nil
}
func (m *memStore) ListAudit(context.Context, int64, time.Time, time.Time, int32, int32) ([]AuditEntry, int64, error) {
	return m.audits, int64(len(m.audits)), nil
}
func (m *memStore) ListPharmacies(context.Context) ([]Pharmacy, error) { return nil, nil }
func (m *memStore) GetPharmacy(context.Context, int64) (Pharmacy, error) {
	return Pharmacy{ID: 1, Slug: "p", Name: "P", SiteDomain: "p.ir", Status: "active"}, nil
}
func (m *memStore) InsertPharmacy(context.Context, PharmacyInput) (int64, error) { return 1, nil }
func (m *memStore) UpdatePharmacy(context.Context, int64, PharmacyInput) error   { return nil }
func (m *memStore) DisablePharmacy(context.Context, int64) error                 { return nil }
func (m *memStore) ListSources(context.Context) ([]Source, error)                { return nil, nil }
func (m *memStore) GetSource(_ context.Context, id int64) (Source, error) {
	s, ok := m.sources[id]
	if !ok {
		return Source{}, ErrNotFound
	}
	return s, nil
}
func (m *memStore) InsertSource(context.Context, SourceInput) (int64, error) { return 1, nil }
func (m *memStore) UpdateSource(context.Context, int64, SourceInput) error   { return nil }
func (m *memStore) DisableSource(context.Context, int64) error               { return nil }
func (m *memStore) ListSourceJobs(context.Context, int64) ([]JobStatus, error) {
	return nil, nil
}
func (m *memStore) LatestSyncRun(context.Context, int64) (*SyncRun, error) { return nil, nil }
func (m *memStore) ListMatches(context.Context, int64, float64, int32, int32) ([]Match, int64, error) {
	return nil, 0, nil
}
func (m *memStore) GetMatch(_ context.Context, id int64) (matchRecord, error) {
	row, ok := m.matches[id]
	if !ok {
		return matchRecord{}, ErrNotFound
	}
	return row, nil
}
func (m *memStore) ApplyMatch(_ context.Context, in applyMatchInput) error {
	row := m.matches[in.ID]
	row.Status = in.Status
	row.LinkedProductID = in.ProductID
	m.matches[in.ID] = row
	m.applied = append(m.applied, in)
	return nil
}
func (m *memStore) ReopenMatch(_ context.Context, id int64, _ int64) error {
	row := m.matches[id]
	row.Status = "pending"
	m.matches[id] = row
	return nil
}
func (m *memStore) GetProduct(_ context.Context, id int64) (Product, error) {
	p, ok := m.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return p, nil
}
func (m *memStore) ListProducts(context.Context, string, string, int32, int32) ([]Product, int64, error) {
	return nil, 0, nil
}
func (m *memStore) UpdateProduct(_ context.Context, p Product) error {
	m.products[p.ID] = p
	return nil
}
func (m *memStore) CreateProductFromItem(context.Context, matchRecord, int64) (int64, error) {
	return 9, nil
}
func (m *memStore) ListCategories(context.Context) ([]Category, error) { return nil, nil }
func (m *memStore) GetCategory(context.Context, int64) (Category, error) {
	return Category{ID: 1, Slug: "c", NameFa: "دسته"}, nil
}
func (m *memStore) InsertCategory(context.Context, Category) (int64, error) { return 1, nil }
func (m *memStore) UpdateCategory(context.Context, Category) error          { return nil }
func (m *memStore) DeleteCategory(context.Context, int64, int64) error      { return nil }
func (m *memStore) CategoryRedirect(context.Context, string) (int64, error) { return 1, nil }
func (m *memStore) CountCategoryProducts(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *memStore) CountCategoryChildren(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *memStore) ListBrands(context.Context) ([]Brand, error) { return nil, nil }
func (m *memStore) GetBrand(context.Context, int64) (Brand, error) {
	return Brand{ID: 1, Slug: "b", NameFa: "برند"}, nil
}
func (m *memStore) InsertBrand(context.Context, Brand) (int64, error) { return 1, nil }
func (m *memStore) UpdateBrand(context.Context, Brand) error          { return nil }
func (m *memStore) MergeBrands(context.Context, int64, int64) error   { return nil }
func (m *memStore) CountBrandProducts(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *memStore) PendingMatchCount(context.Context) (int64, error) { return 0, nil }
func (m *memStore) ListSourceHealth(context.Context) ([]SourceHealth, error) {
	return nil, nil
}
func (m *memStore) ListSyncSuccessRates(context.Context, time.Time, int64) (map[int64]successRate, error) {
	return map[int64]successRate{}, nil
}
func (m *memStore) ListSourceRuns(context.Context, int64, int32) ([]SyncRun, error) {
	return nil, nil
}
func (m *memStore) ListStaleSources(context.Context, time.Time) ([]StaleSource, error) {
	return nil, nil
}
func (m *memStore) ListSuspicious(context.Context, int32, int32) ([]SuspiciousOffer, int64, error) {
	return nil, 0, nil
}
func (m *memStore) GetSuspicious(_ context.Context, id int64) (SuspiciousOffer, error) {
	if id != 1 {
		return SuspiciousOffer{}, ErrNotFound
	}
	return SuspiciousOffer{ID: 1, CurrentPriceRial: 1000, ProposedPriceRial: 9000, SourceID: 2, Status: "suspicious"}, nil
}
func (m *memStore) ApproveSuspicious(_ context.Context, id int64) (SuspiciousOffer, error) {
	return SuspiciousOffer{ID: id, CurrentPriceRial: 9000, ProposedPriceRial: 9000}, nil
}
func (m *memStore) RejectSuspicious(context.Context, int64, int64) error { return nil }
func (m *memStore) ClickReport(context.Context, time.Time, time.Time) (ClickReport, error) {
	return ClickReport{}, nil
}

type nopQueue struct{}

func (nopQueue) EnqueueSyncSource(context.Context, int64) (bool, error) { return true, nil }

func testService(store *memStore) *Service {
	return NewService(store, nil, nopQueue{}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func TestLoginWrongPasswordFails(t *testing.T) {
	store := newMem()
	hash, _ := HashPassword("correct")
	store.users["ops"] = userRecord{Actor: Actor{ID: 1, Username: "ops", Role: RoleDataOps}, PasswordHash: hash, Active: true}
	svc := testService(store)
	if _, err := svc.Login(context.Background(), "ops", "nope", "1.1.1.1"); err != ErrInvalidLogin {
		t.Fatalf("err = %v", err)
	}
}

func TestLoginRateLimit(t *testing.T) {
	store := newMem()
	store.attempts = loginMaxAttempt
	svc := testService(store)
	if _, err := svc.Login(context.Background(), "ops", "x", "1.1.1.1"); err != ErrRateLimited {
		t.Fatalf("err = %v", err)
	}
}

func TestSessionWrongRoleIsForbidden(t *testing.T) {
	store := newMem()
	store.sessions[hashToken("tok")] = Actor{ID: 2, Username: "viewer", Role: "viewer"}
	svc := testService(store)
	if _, err := svc.SessionActor(context.Background(), "tok"); err != ErrForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestSessionMissingIsUnauthorized(t *testing.T) {
	svc := testService(newMem())
	if _, err := svc.SessionActor(context.Background(), ""); err != ErrUnauthorized {
		t.Fatalf("err = %v", err)
	}
}

func TestApproveMatchWritesAudit(t *testing.T) {
	store := newMem()
	store.matches[3] = matchRecord{Match: Match{ID: 3, Status: "pending", SuggestedProductID: 10, SourceItemID: 4}}
	svc := testService(store)
	actor := Actor{ID: 1, Username: "ops", Role: RoleDataOps}
	if err := svc.ApproveMatch(context.Background(), actor, 3); err != nil {
		t.Fatal(err)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "approve" {
		t.Fatalf("audits = %+v", store.audits)
	}
	if store.matches[3].Status != "approved" {
		t.Fatalf("status = %s", store.matches[3].Status)
	}
}

func TestUpdateProductLocksEditedFields(t *testing.T) {
	store := newMem()
	store.products[5] = Product{ID: 5, NameFa: "قدیم", Status: "published", SourceSnapshot: json.RawMessage(`{"name_fa":"منبع"}`)}
	svc := testService(store)
	name := "جدید"
	status := "hidden"
	got, err := svc.UpdateProduct(context.Background(), Actor{ID: 1, Username: "ops"}, 5, ProductPatch{NameFa: &name, Status: &status})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "hidden" {
		t.Fatalf("status = %s", got.Status)
	}
	if !contains(got.LockedFields, "name_fa") {
		t.Fatalf("locked_fields = %v", got.LockedFields)
	}
}

func TestStatusOnlyPatchDoesNotLockImage(t *testing.T) {
	store := newMem()
	store.products[5] = Product{ID: 5, NameFa: "قدیم", ImageURL: "https://x/a.jpg", Status: "published"}
	svc := testService(store)
	hidden := "hidden"
	got, err := svc.UpdateProduct(context.Background(), Actor{ID: 1, Username: "ops"}, 5, ProductPatch{Status: &hidden})
	if err != nil {
		t.Fatal(err)
	}
	if got.ImageURL != "https://x/a.jpg" {
		t.Fatalf("image wiped: %s", got.ImageURL)
	}
	if contains(got.LockedFields, "image_url") {
		t.Fatalf("status patch locked image: %v", got.LockedFields)
	}
}

func TestCreatePharmacyWritesAudit(t *testing.T) {
	store := newMem()
	svc := testService(store)
	if _, err := svc.CreatePharmacy(context.Background(), Actor{ID: 1, Username: "ops"}, PharmacyInput{
		Name: "داروخانه آزمایشی", Slug: "test-pharm", SiteDomain: "test.ir",
	}); err != nil {
		t.Fatal(err)
	}
	if len(store.audits) != 1 {
		t.Fatalf("expected audit row, got %d", len(store.audits))
	}
}

func TestApprovePriceWritesAudit(t *testing.T) {
	store := newMem()
	svc := testService(store)
	actor := Actor{ID: 1, Username: "ops", Role: RoleDataOps}
	if err := svc.ApprovePrice(context.Background(), actor, 1); err != nil {
		t.Fatal(err)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "approve_price" {
		t.Fatalf("audits = %+v", store.audits)
	}
}

func TestRejectPriceWritesAudit(t *testing.T) {
	store := newMem()
	svc := testService(store)
	if err := svc.RejectPrice(context.Background(), Actor{ID: 1, Username: "ops"}, 1); err != nil {
		t.Fatal(err)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "reject_price" {
		t.Fatalf("audits = %+v", store.audits)
	}
}

func TestSourceOverdue(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	last := now.Add(-2 * time.Hour)
	if !sourceOverdue(&last, "01:00:00", true, now) {
		t.Fatal("source two hours late should be overdue")
	}
	fresh := now.Add(-30 * time.Minute)
	if sourceOverdue(&fresh, "01:00:00", true, now) {
		t.Fatal("source inside schedule should not be overdue")
	}
	if sourceOverdue(nil, "1 hour", false, now) {
		t.Fatal("disabled source is not overdue")
	}
}
