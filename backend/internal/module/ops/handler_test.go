package ops

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testWriteError(w http.ResponseWriter, _ *http.Request, err error) {
	switch err {
	case ErrUnauthorized:
		w.WriteHeader(http.StatusUnauthorized)
	case ErrForbidden:
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestOpsRoutesRequireSession(t *testing.T) {
	store := newMem()
	h := NewHandler(testService(store), testWriteError, false)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ops/sources", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestWrongRoleIsForbidden(t *testing.T) {
	store := newMem()
	store.sessions[hashToken("viewer-token")] = Actor{ID: 9, Username: "v", Role: "viewer"}
	h := NewHandler(testService(store), testWriteError, false)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ops/me", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "viewer-token"})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestPublicUserCookieDoesNotUnlockOps(t *testing.T) {
	h := NewHandler(testService(newMem()), testWriteError, false)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ops/sources", nil)
	req.AddCookie(&http.Cookie{Name: "user_session", Value: "public-user-token"})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: role user must not reach ops", rec.Code)
	}
}
