package identity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Service is the identity business layer: OTP login, sessions, search history.
type Service struct {
	store  Store
	sms    SMSSender
	log    *slog.Logger
	pepper string
	now    func() time.Time
}

func NewService(store Store, sms SMSSender, pepper string, log *slog.Logger) *Service {
	if sms == nil {
		sms = NewUnavailableSender()
	}
	if log == nil {
		log = slog.Default()
	}
	return &Service{store: store, sms: sms, log: log, pepper: pepper, now: time.Now}
}

// RequestOTP normalises the phone, enforces rate limits, stores a hashed code
// and asks the sender to deliver it. The plaintext code never enters slog.
func (s *Service) RequestOTP(ctx context.Context, rawPhone, ip string) (OTPRequest, error) {
	phone, err := NormalizePhone(rawPhone)
	if err != nil {
		return OTPRequest{}, err
	}
	since := s.now().UTC().Add(-otpSendWindow)
	nPhone, err := s.store.CountOTPSendsPhone(ctx, phone, since)
	if err != nil {
		return OTPRequest{}, fmt.Errorf("count otp sends by phone: %w", err)
	}
	nIP, err := s.store.CountOTPSendsIP(ctx, ip, since)
	if err != nil {
		return OTPRequest{}, fmt.Errorf("count otp sends by ip: %w", err)
	}
	if nPhone >= otpMaxPerPhone || nIP >= otpMaxPerIP {
		s.log.InfoContext(ctx, "otp.rate_limited", "phone", MaskPhone(phone))
		return OTPRequest{}, ErrRateLimited
	}

	latest, err := s.store.LatestOTP(ctx, phone)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return OTPRequest{}, fmt.Errorf("latest otp: %w", err)
	}
	if err == nil && !latest.Consumed && s.now().UTC().Sub(latest.CreatedAt) < otpResendAfter {
		s.log.InfoContext(ctx, "otp.rate_limited", "phone", MaskPhone(phone))
		return OTPRequest{}, ErrRateLimited
	}

	code, err := randomDigits(otpCodeLen)
	if err != nil {
		return OTPRequest{}, fmt.Errorf("generate otp: %w", err)
	}
	expires := s.now().UTC().Add(otpTTL)
	if _, err := s.store.InsertOTP(ctx, phone, hashOTP(s.pepper, phone, code), ip, expires); err != nil {
		return OTPRequest{}, fmt.Errorf("insert otp: %w", err)
	}
	if err := s.store.InsertOTPSend(ctx, phone, ip); err != nil {
		return OTPRequest{}, fmt.Errorf("insert otp send: %w", err)
	}
	if err := s.sms.SendOTP(phone, code); err != nil {
		return OTPRequest{}, err
	}
	s.log.InfoContext(ctx, "otp.sent", "phone", MaskPhone(phone))
	return OTPRequest{PhoneMasked: MaskPhone(phone), ResendAfter: otpResendAfter}, nil
}

// VerifyOTP consumes a matching unexpired code and issues a session cookie token.
func (s *Service) VerifyOTP(ctx context.Context, rawPhone, code, _ string) (Session, error) {
	phone, err := NormalizePhone(rawPhone)
	if err != nil {
		return Session{}, err
	}
	code = strings.TrimSpace(code)
	if len(code) != otpCodeLen {
		return Session{}, ErrInvalidOTP
	}

	rec, err := s.store.LatestOTP(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrInvalidOTP
		}
		return Session{}, fmt.Errorf("latest otp: %w", err)
	}
	if rec.Consumed {
		return Session{}, ErrOTPUsed
	}
	if !rec.ExpiresAt.After(s.now().UTC()) {
		_, _ = s.store.ConsumeOTP(ctx, rec.ID)
		return Session{}, ErrOTPExpired
	}
	if rec.Attempts >= otpMaxAttempts {
		s.log.InfoContext(ctx, "otp.rate_limited", "phone", MaskPhone(phone))
		return Session{}, ErrRateLimited
	}
	if !compareOTP(s.pepper, phone, code, rec.Hash) {
		_ = s.store.BumpOTPAttempts(ctx, rec.ID)
		if rec.Attempts+1 >= otpMaxAttempts {
			s.log.InfoContext(ctx, "otp.rate_limited", "phone", MaskPhone(phone))
			return Session{}, ErrRateLimited
		}
		return Session{}, ErrInvalidOTP
	}
	ok, err := s.store.ConsumeOTP(ctx, rec.ID)
	if err != nil {
		return Session{}, fmt.Errorf("consume otp: %w", err)
	}
	if !ok {
		return Session{}, ErrOTPUsed
	}

	user, err := s.store.InsertUser(ctx, phone)
	if err != nil {
		return Session{}, fmt.Errorf("upsert user: %w", err)
	}
	token, err := randomToken()
	if err != nil {
		return Session{}, fmt.Errorf("session token: %w", err)
	}
	expires := s.now().UTC().Add(sessionTTL)
	if err := s.store.InsertSession(ctx, user.ID, hashToken(token), expires); err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}
	s.log.InfoContext(ctx, "otp.verified", "phone", MaskPhone(phone), "user_id", user.ID)
	return Session{Token: token, ExpiresAt: expires, User: user}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.store.DeleteSession(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	s.log.InfoContext(ctx, "auth.logout")
	return nil
}

func (s *Service) UserFromSession(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}
	user, expires, err := s.store.UserBySession(ctx, hashToken(token))
	if err != nil {
		return User{}, ErrUnauthorized
	}
	if !expires.After(s.now().UTC()) {
		_ = s.store.DeleteSession(ctx, hashToken(token))
		return User{}, ErrUnauthorized
	}
	if user.Role != RoleUser {
		return User{}, ErrUnauthorized
	}
	return user, nil
}

func (s *Service) ListSearches(ctx context.Context, userID int64) ([]SearchEntry, error) {
	items, err := s.store.ListSearches(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list searches: %w", err)
	}
	return items, nil
}

func (s *Service) DeleteSearches(ctx context.Context, userID int64) error {
	if err := s.store.DeleteSearches(ctx, userID); err != nil {
		return fmt.Errorf("delete searches: %w", err)
	}
	return nil
}
