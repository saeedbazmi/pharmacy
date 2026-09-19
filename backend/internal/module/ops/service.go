package ops

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/ingest"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/jobq"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

// Enqueuer schedules a source sync without the service knowing jobq internals.
type Enqueuer interface {
	EnqueueSyncSource(ctx context.Context, sourceID int64) (bool, error)
}

// Service is the ops business layer.
type Service struct {
	store    Store
	fetchers *ingest.Registry
	queue    Enqueuer
	log      *slog.Logger
	now      func() time.Time
}

func NewService(store Store, fetchers *ingest.Registry, queue Enqueuer, log *slog.Logger) *Service {
	return &Service{store: store, fetchers: fetchers, queue: queue, log: log, now: time.Now}
}

func (s *Service) fetcherNames() []string {
	if s.fetchers == nil {
		return nil
	}
	return s.fetchers.Names()
}

func (s *Service) Login(ctx context.Context, username, password, ip string) (Session, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return Session{}, ErrInvalidLogin
	}
	since := s.now().UTC().Add(-loginWindow)
	n, err := s.store.CountRecentLogins(ctx, username, ip, since)
	if err != nil {
		return Session{}, err
	}
	if n >= loginMaxAttempt {
		s.log.WarnContext(ctx, "auth.rate_limited", "username", username)
		return Session{}, ErrRateLimited
	}
	_ = s.store.InsertLoginAttempt(ctx, username, ip)

	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.InfoContext(ctx, "auth.failed", "username", username)
		return Session{}, ErrInvalidLogin
	}
	if err := comparePassword(user.PasswordHash, password); err != nil {
		s.log.InfoContext(ctx, "auth.failed", "username", username)
		return Session{}, ErrInvalidLogin
	}
	if !user.Active {
		return Session{}, ErrInactiveUser
	}
	if !allowedRole(user.Role) {
		return Session{}, ErrForbidden
	}
	token, hash, err := newSessionToken()
	if err != nil {
		return Session{}, err
	}
	expires := s.now().UTC().Add(sessionTTL)
	if err := s.store.InsertSession(ctx, user.ID, hash, expires); err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}
	s.log.InfoContext(ctx, "auth.login", "actor_id", user.ID, "username", user.Username)
	return Session{Token: token, ExpiresAt: expires, Actor: user.Actor}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.store.DeleteSession(ctx, hashToken(token)); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "auth.logout")
	return nil
}

func (s *Service) SessionActor(ctx context.Context, token string) (Actor, error) {
	if token == "" {
		return Actor{}, ErrUnauthorized
	}
	actor, expires, err := s.store.SessionUser(ctx, hashToken(token))
	if err != nil {
		return Actor{}, err
	}
	if !expires.After(s.now().UTC()) {
		_ = s.store.DeleteSession(ctx, hashToken(token))
		return Actor{}, ErrUnauthorized
	}
	if !allowedRole(actor.Role) {
		return Actor{}, ErrForbidden
	}
	return actor, nil
}

func (s *Service) audit(ctx context.Context, actor Actor, entity, entityID, action string, before, after any) {
	entry := AuditEntry{
		ActorID:   actor.ID,
		ActorName: actor.Username,
		Entity:    entity,
		EntityID:  entityID,
		Action:    action,
		Before:    MaskSecrets(mustJSON(before)),
		After:     MaskSecrets(mustJSON(after)),
	}
	if err := s.store.WriteAudit(ctx, entry); err != nil {
		s.log.ErrorContext(ctx, "audit.failed", "entity", entity, "action", action, "error", err.Error())
	}
}

func (s *Service) ListAudit(ctx context.Context, actorID int64, since, until time.Time, page, size int) (Page[AuditEntry], error) {
	page, size = clampPage(page, size)
	if since.IsZero() {
		since = s.now().UTC().AddDate(0, -1, 0)
	}
	if until.IsZero() {
		until = s.now().UTC().Add(time.Second)
	}
	items, total, err := s.store.ListAudit(ctx, actorID, since, until, int32(size), int32((page-1)*size))
	if err != nil {
		return Page[AuditEntry]{}, err
	}
	return Page[AuditEntry]{Items: items, Total: total, Page: page, Size: size}, nil
}

func (s *Service) ListPharmacies(ctx context.Context) ([]Pharmacy, error) {
	return s.store.ListPharmacies(ctx)
}

func (s *Service) CreatePharmacy(ctx context.Context, actor Actor, in PharmacyInput) (Pharmacy, error) {
	in, err := normalizePharmacy(in)
	if err != nil {
		return Pharmacy{}, err
	}
	id, err := s.store.InsertPharmacy(ctx, in)
	if err != nil {
		return Pharmacy{}, err
	}
	got, err := s.store.GetPharmacy(ctx, id)
	if err != nil {
		return Pharmacy{}, err
	}
	s.audit(ctx, actor, "pharmacy", fmt.Sprint(id), "create", nil, got)
	return got, nil
}

func (s *Service) UpdatePharmacy(ctx context.Context, actor Actor, id int64, in PharmacyInput) (Pharmacy, error) {
	before, err := s.store.GetPharmacy(ctx, id)
	if err != nil {
		return Pharmacy{}, err
	}
	in, err = normalizePharmacy(in)
	if err != nil {
		return Pharmacy{}, err
	}
	if in.Status == "" {
		in.Status = before.Status
	}
	if err := s.store.UpdatePharmacy(ctx, id, in); err != nil {
		return Pharmacy{}, err
	}
	got, err := s.store.GetPharmacy(ctx, id)
	if err != nil {
		return Pharmacy{}, err
	}
	s.audit(ctx, actor, "pharmacy", fmt.Sprint(id), "update", before, got)
	return got, nil
}

func (s *Service) DisablePharmacy(ctx context.Context, actor Actor, id int64) error {
	before, err := s.store.GetPharmacy(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DisablePharmacy(ctx, id); err != nil {
		return err
	}
	s.audit(ctx, actor, "pharmacy", fmt.Sprint(id), "disable", before, map[string]string{"status": "disabled"})
	return nil
}

func (s *Service) ListSources(ctx context.Context) ([]Source, error) {
	items, err := s.store.ListSources(ctx)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Config = MaskSecrets(items[i].Config)
	}
	return items, nil
}

func (s *Service) GetSource(ctx context.Context, id int64) (Source, error) {
	src, err := s.store.GetSource(ctx, id)
	if err != nil {
		return Source{}, err
	}
	run, err := s.store.LatestSyncRun(ctx, id)
	if err != nil {
		return Source{}, err
	}
	if run != nil {
		src.LastStatus = run.Status
		src.LastError = run.Error
	}
	src.Config = MaskSecrets(src.Config)
	return src, nil
}

func (s *Service) CreateSource(ctx context.Context, actor Actor, in SourceInput) (Source, error) {
	norm, stored, err := s.normalizeSource(in, nil)
	if err != nil {
		return Source{}, err
	}
	norm.Config = stored
	id, err := s.store.InsertSource(ctx, norm)
	if err != nil {
		return Source{}, err
	}
	got, err := s.GetSource(ctx, id)
	if err != nil {
		return Source{}, err
	}
	s.audit(ctx, actor, "data_source", fmt.Sprint(id), "create", nil, got)
	return got, nil
}

func (s *Service) UpdateSource(ctx context.Context, actor Actor, id int64, in SourceInput) (Source, error) {
	before, err := s.store.GetSource(ctx, id)
	if err != nil {
		return Source{}, err
	}
	norm, stored, err := s.normalizeSource(in, before.Config)
	if err != nil {
		return Source{}, err
	}
	if in.Enabled {
		norm.Enabled = true
	} else {
		norm.Enabled = before.Enabled
	}
	// Enabled is explicit when the client sent the field; keep incoming value.
	norm.Enabled = in.Enabled
	norm.Config = stored
	if err := s.store.UpdateSource(ctx, id, norm); err != nil {
		return Source{}, err
	}
	got, err := s.GetSource(ctx, id)
	if err != nil {
		return Source{}, err
	}
	s.audit(ctx, actor, "data_source", fmt.Sprint(id), "update", map[string]any{
		"kind": before.Kind, "schedule": before.ScheduleInterval, "enabled": before.Enabled,
	}, got)
	return got, nil
}

func (s *Service) DisableSource(ctx context.Context, actor Actor, id int64) error {
	before, err := s.store.GetSource(ctx, id)
	if err != nil {
		return err
	}
	if err := s.store.DisableSource(ctx, id); err != nil {
		return err
	}
	s.audit(ctx, actor, "data_source", fmt.Sprint(id), "disable", before, map[string]bool{"enabled": false})
	return nil
}

func (s *Service) TestConnection(ctx context.Context, id int64) ([]PreviewItem, error) {
	src, err := s.store.GetSource(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.preview(ctx, src.Config)
}

func (s *Service) TestDraft(ctx context.Context, in SourceInput) ([]PreviewItem, error) {
	if in.PharmacyID == 0 {
		in.PharmacyID = 1
	}
	_, stored, err := s.normalizeSource(in, nil)
	if err != nil {
		return nil, err
	}
	return s.preview(ctx, stored)
}

func (s *Service) preview(ctx context.Context, config json.RawMessage) ([]PreviewItem, error) {
	cfg, err := parseSourceConfig(config, s.fetcherNames())
	if err != nil {
		return nil, err
	}
	fetcher, ok := s.fetchers.Get(cfg.Fetcher)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownFetcher, cfg.Fetcher)
	}
	items, err := fetcher.Fetch(ctx, ingest.Source{Config: mustJSON(cfg)})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	if len(items) > previewLimit {
		items = items[:previewLimit]
	}
	out := make([]PreviewItem, 0, len(items))
	for _, item := range items {
		out = append(out, PreviewItem{
			ExternalID: item.ExternalID, NameFa: item.NameFa, BrandName: item.BrandName,
			PriceToman: item.PriceToman, InStock: item.InStock, ProductURL: item.ProductURL, ImageURL: item.ImageURL,
		})
	}
	return out, nil
}

func (s *Service) SyncNow(ctx context.Context, actor Actor, id int64) (JobStatus, error) {
	if _, err := s.store.GetSource(ctx, id); err != nil {
		return JobStatus{}, err
	}
	enqueued, err := s.queue.EnqueueSyncSource(ctx, id)
	if err != nil {
		return JobStatus{}, err
	}
	jobs, err := s.store.ListSourceJobs(ctx, id)
	if err != nil {
		return JobStatus{}, err
	}
	run, err := s.store.LatestSyncRun(ctx, id)
	if err != nil {
		return JobStatus{}, err
	}
	status := JobStatus{Enqueued: enqueued, LatestRun: run}
	if len(jobs) > 0 {
		status.JobID = jobs[0].JobID
		status.Kind = jobs[0].Kind
		status.Status = jobs[0].Status
		status.Attempts = jobs[0].Attempts
		status.RunAt = jobs[0].RunAt
		status.LastError = jobs[0].LastError
		status.CreatedAt = jobs[0].CreatedAt
	}
	s.audit(ctx, actor, "data_source", fmt.Sprint(id), "sync", nil, map[string]any{"enqueued": enqueued})
	return status, nil
}

func (s *Service) SourceJobs(ctx context.Context, id int64) (JobStatus, error) {
	if _, err := s.store.GetSource(ctx, id); err != nil {
		return JobStatus{}, err
	}
	jobs, err := s.store.ListSourceJobs(ctx, id)
	if err != nil {
		return JobStatus{}, err
	}
	run, err := s.store.LatestSyncRun(ctx, id)
	if err != nil {
		return JobStatus{}, err
	}
	status := JobStatus{LatestRun: run}
	if len(jobs) > 0 {
		status.JobID = jobs[0].JobID
		status.Kind = jobs[0].Kind
		status.Status = jobs[0].Status
		status.Attempts = jobs[0].Attempts
		status.RunAt = jobs[0].RunAt
		status.LastError = jobs[0].LastError
		status.CreatedAt = jobs[0].CreatedAt
	}
	return status, nil
}

func (s *Service) normalizeSource(in SourceInput, previous json.RawMessage) (SourceInput, json.RawMessage, error) {
	kind, err := normalizeKind(in.Kind)
	if err != nil {
		return in, nil, err
	}
	schedule, err := normalizeSchedule(in.ScheduleInterval)
	if err != nil {
		return in, nil, err
	}
	merged := in.Config
	if previous != nil {
		merged = MergeSecrets(previous, in.Config)
	}
	cfg, err := parseSourceConfig(merged, s.fetcherNames())
	if err != nil {
		return in, nil, err
	}
	if in.PharmacyID <= 0 {
		return in, nil, fmt.Errorf("%w: pharmacy_id is required", ErrInvalidInput)
	}
	stored, err := json.Marshal(cfg)
	if err != nil {
		return in, nil, err
	}
	in.Kind = kind
	in.ScheduleInterval = schedule
	in.Config = stored
	return in, stored, nil
}

func normalizePharmacy(in PharmacyInput) (PharmacyInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Slug = textfa.Slugify(in.Slug)
	if in.Slug == "" {
		in.Slug = textfa.Slugify(in.Name)
	}
	in.SiteDomain = strings.TrimSpace(in.SiteDomain)
	if in.Name == "" || in.Slug == "" || in.SiteDomain == "" {
		return in, fmt.Errorf("%w: name, slug and site_domain are required", ErrInvalidInput)
	}
	if in.Status == "" {
		in.Status = "active"
	}
	switch in.Status {
	case "active", "paused", "disabled":
	default:
		return in, fmt.Errorf("%w: invalid pharmacy status", ErrInvalidInput)
	}
	return in, nil
}

func newSessionToken() (plain, hash string, err error) {
	var buf [32]byte
	if _, err = rand.Read(buf[:]); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(buf[:])
	return plain, hashToken(plain), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func mustJSON(v any) json.RawMessage {
	if v == nil {
		return json.RawMessage(`{}`)
	}
	if raw, ok := v.(json.RawMessage); ok {
		return raw
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func clampPage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

const passwordIters = 120000

// HashPassword stores a slow salted SHA-256 stretch. Used by the opsuser command.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return encodePassword(password, salt, passwordIters), nil
}

func encodePassword(password string, salt []byte, iters int) string {
	sum := sha256.Sum256(append(append([]byte{}, salt...), password...))
	buf := sum[:]
	for i := 1; i < iters; i++ {
		next := sha256.Sum256(append(append([]byte{}, salt...), buf...))
		buf = next[:]
	}
	return fmt.Sprintf("sha256i$%d$%s$%s", iters, hex.EncodeToString(salt), hex.EncodeToString(buf))
}

func comparePassword(encoded, password string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "sha256i" {
		return ErrInvalidLogin
	}
	iters, err := strconv.Atoi(parts[1])
	if err != nil || iters < 1000 {
		return ErrInvalidLogin
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return ErrInvalidLogin
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil {
		return ErrInvalidLogin
	}
	got := encodePassword(password, salt, iters)
	gotParts := strings.Split(got, "$")
	if len(gotParts) != 4 {
		return ErrInvalidLogin
	}
	have, err := hex.DecodeString(gotParts[3])
	if err != nil {
		return ErrInvalidLogin
	}
	if subtle.ConstantTimeCompare(want, have) != 1 {
		return ErrInvalidLogin
	}
	return nil
}

// Compile-time check that jobq.Queue satisfies Enqueuer.
var _ Enqueuer = (*jobq.Queue)(nil)
