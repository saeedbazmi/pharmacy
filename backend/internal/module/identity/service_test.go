package identity

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type memStore struct {
	users    map[string]User
	otps     map[string]otpRecord
	sends    []struct{ phone, ip string }
	sessions map[string]User
	expiry   map[string]time.Time
	searches []SearchEntry
	nextID   int64
}

func newMem() *memStore {
	return &memStore{
		users:    map[string]User{},
		otps:     map[string]otpRecord{},
		sessions: map[string]User{},
		expiry:   map[string]time.Time{},
		nextID:   1,
	}
}

func (m *memStore) GetUserByPhone(_ context.Context, phone string) (User, error) {
	u, ok := m.users[phone]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}
func (m *memStore) GetUserByID(_ context.Context, id int64) (User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}
func (m *memStore) InsertUser(_ context.Context, phone string) (User, error) {
	if u, ok := m.users[phone]; ok {
		return u, nil
	}
	u := User{ID: m.nextID, Phone: phone, Role: RoleUser}
	m.nextID++
	m.users[phone] = u
	return u, nil
}
func (m *memStore) InsertOTP(_ context.Context, phone, hash, _ string, expires time.Time) (int64, error) {
	id := m.nextID
	m.nextID++
	m.otps[phone] = otpRecord{ID: id, Phone: phone, Hash: hash, ExpiresAt: expires, CreatedAt: expires.Add(-otpTTL)}
	return id, nil
}
func (m *memStore) LatestOTP(_ context.Context, phone string) (otpRecord, error) {
	rec, ok := m.otps[phone]
	if !ok {
		return otpRecord{}, ErrNotFound
	}
	return rec, nil
}
func (m *memStore) ConsumeOTP(_ context.Context, id int64) (bool, error) {
	for phone, rec := range m.otps {
		if rec.ID == id {
			if rec.Consumed {
				return false, nil
			}
			rec.Consumed = true
			m.otps[phone] = rec
			return true, nil
		}
	}
	return false, nil
}
func (m *memStore) BumpOTPAttempts(_ context.Context, id int64) error {
	for phone, rec := range m.otps {
		if rec.ID == id {
			rec.Attempts++
			m.otps[phone] = rec
			return nil
		}
	}
	return nil
}
func (m *memStore) InsertOTPSend(_ context.Context, phone, ip string) error {
	m.sends = append(m.sends, struct{ phone, ip string }{phone, ip})
	return nil
}
func (m *memStore) CountOTPSendsPhone(_ context.Context, phone string, _ time.Time) (int64, error) {
	var n int64
	for _, s := range m.sends {
		if s.phone == phone {
			n++
		}
	}
	return n, nil
}
func (m *memStore) CountOTPSendsIP(_ context.Context, ip string, _ time.Time) (int64, error) {
	var n int64
	for _, s := range m.sends {
		if s.ip == ip {
			n++
		}
	}
	return n, nil
}
func (m *memStore) InsertSession(_ context.Context, userID int64, tokenHash string, expires time.Time) error {
	for _, u := range m.users {
		if u.ID == userID {
			m.sessions[tokenHash] = u
			m.expiry[tokenHash] = expires
			return nil
		}
	}
	return ErrNotFound
}
func (m *memStore) UserBySession(_ context.Context, tokenHash string) (User, time.Time, error) {
	u, ok := m.sessions[tokenHash]
	if !ok {
		return User{}, time.Time{}, ErrUnauthorized
	}
	return u, m.expiry[tokenHash], nil
}
func (m *memStore) DeleteSession(_ context.Context, tokenHash string) error {
	delete(m.sessions, tokenHash)
	return nil
}
func (m *memStore) ListSearches(context.Context, int64) ([]SearchEntry, error) {
	return m.searches, nil
}
func (m *memStore) DeleteSearches(context.Context, int64) error {
	m.searches = nil
	return nil
}

type captureSMS struct {
	lastPhone string
	lastCode  string
}

func (c *captureSMS) SendOTP(phone, code string) error {
	c.lastPhone, c.lastCode = phone, code
	return nil
}
func (c *captureSMS) Notify(string, string) error { return nil }

func testService(store Store, sms SMSSender, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	svc := NewService(store, sms, "test-pepper", log)
	svc.now = func() time.Time { return time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC) }
	return svc
}

func TestRequestOTPRejectsInvalidPhone(t *testing.T) {
	svc := testService(newMem(), &captureSMS{}, nil)
	if _, err := svc.RequestOTP(context.Background(), "021", "1.1.1.1"); err != ErrInvalidPhone {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyOTPSuccessAndSingleUse(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	svc := testService(store, sms, nil)
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	session, err := svc.VerifyOTP(context.Background(), "09121234567", sms.lastCode, "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.Role != RoleUser || session.Token == "" {
		t.Fatalf("session = %+v", session)
	}
	if _, err := svc.VerifyOTP(context.Background(), "09121234567", sms.lastCode, "1.1.1.1"); err != ErrOTPUsed {
		t.Fatalf("second verify err = %v, want used", err)
	}
}

func TestVerifyOTPWrongCode(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	svc := testService(store, sms, nil)
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.VerifyOTP(context.Background(), "09121234567", "000000", "1.1.1.1"); err != ErrInvalidOTP {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyOTPExpired(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	svc := testService(store, sms, nil)
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	svc.now = func() time.Time { return now.Add(otpTTL + time.Second) }
	if _, err := svc.VerifyOTP(context.Background(), "09121234567", sms.lastCode, "1.1.1.1"); err != ErrOTPExpired {
		t.Fatalf("err = %v", err)
	}
}

func TestRequestOTPRateLimitByPhone(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	svc := testService(store, sms, nil)
	base := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	for i := 0; i < otpMaxPerPhone; i++ {
		svc.now = func() time.Time { return base.Add(time.Duration(i) * 2 * time.Minute) }
		if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
	svc.now = func() time.Time { return base.Add(20 * time.Minute) }
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != ErrRateLimited {
		t.Fatalf("err = %v, want rate limited", err)
	}
}

func TestRequestOTPDoesNotLogCodeOrFullPhone(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	svc := testService(store, sms, log)
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	body := logs.String()
	if !strings.Contains(body, "otp.sent") {
		t.Fatalf("missing otp.sent: %s", body)
	}
	if strings.Contains(body, sms.lastCode) {
		t.Fatal("otp code appeared in structured logs")
	}
	if strings.Contains(body, "09121234567") {
		t.Fatal("full phone appeared in structured logs")
	}
	if !strings.Contains(body, "0912***4567") {
		t.Fatalf("masked phone missing: %s", body)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	svc := testService(store, sms, nil)
	if _, err := svc.RequestOTP(context.Background(), "09121234567", "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	session, err := svc.VerifyOTP(context.Background(), "09121234567", sms.lastCode, "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), session.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UserFromSession(context.Background(), session.Token); err != ErrUnauthorized {
		t.Fatalf("err = %v", err)
	}
}
