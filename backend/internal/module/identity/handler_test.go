package identity

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testWriteError(w http.ResponseWriter, _ *http.Request, err error) {
	switch err {
	case ErrUnauthorized:
		w.WriteHeader(http.StatusUnauthorized)
	case ErrInvalidPhone, ErrInvalidOTP, ErrOTPExpired, ErrOTPUsed:
		w.WriteHeader(http.StatusBadRequest)
	case ErrRateLimited:
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestMeRequiresSession(t *testing.T) {
	h := NewHandler(testService(newMem(), &captureSMS{}, nil), testWriteError, false)
	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://x/api/v1/auth/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestVerifySetsHttpOnlyCookie(t *testing.T) {
	store := newMem()
	sms := &captureSMS{}
	h := NewHandler(testService(store, sms, nil), testWriteError, false)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "http://x/api/v1/auth/otp/request", strings.NewReader(`{"phone":"09121234567"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("request status = %d body=%s", rec.Code, rec.Body.String())
	}

	verify := httptest.NewRequest(http.MethodPost, "http://x/api/v1/auth/otp/verify", strings.NewReader(`{"phone":"09121234567","code":"`+sms.lastCode+`"}`))
	verify.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, verify)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookie {
			found = c
		}
	}
	if found == nil || found.Value == "" || !found.HttpOnly {
		t.Fatalf("cookie = %+v", found)
	}
	if found.SameSite != http.SameSiteLaxMode {
		t.Fatalf("samesite = %v", found.SameSite)
	}
}
