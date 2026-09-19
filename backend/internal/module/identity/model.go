package identity

import (
	"errors"
	"time"
)

const (
	RoleUser       = "user"
	SessionCookie  = "user_session"
	sessionTTL     = 30 * 24 * time.Hour
	otpTTL         = 2 * time.Minute
	otpSendWindow  = 15 * time.Minute
	otpMaxPerPhone = 5
	otpMaxPerIP    = 10
	otpMaxAttempts = 5
	otpResendAfter = 60 * time.Second
	otpCodeLen     = 6
)

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidPhone   = errors.New("invalid phone")
	ErrInvalidOTP     = errors.New("invalid otp")
	ErrOTPExpired     = errors.New("otp expired")
	ErrOTPUsed        = errors.New("otp already used")
	ErrRateLimited    = errors.New("otp rate limited")
	ErrSMSUnavailable = errors.New("sms unavailable")
	ErrNotFound       = errors.New("not found")
)

// User is a registered public account.
type User struct {
	ID    int64
	Phone string
	Role  string
}

// Session is a live login cookie.
type Session struct {
	Token     string
	ExpiresAt time.Time
	User      User
}

// OTPRequest is returned after a code is accepted for delivery.
type OTPRequest struct {
	PhoneMasked string
	ResendAfter time.Duration
}

type otpRecord struct {
	ID        int64
	Phone     string
	Hash      string
	ExpiresAt time.Time
	Consumed  bool
	Attempts  int32
	CreatedAt time.Time
}
