package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
)

type stubPinger struct{ err error }

func (s stubPinger) Ping(context.Context) error { return s.err }

type stubProductReader struct {
	product catalog.Product
	err     error
}

func (s stubProductReader) ProductBySlug(context.Context, string) (catalog.Product, error) {
	return s.product, s.err
}

func newTestRouter(t *testing.T, pinger Pinger, reader catalog.ProductReader) http.Handler {
	t.Helper()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return NewRouter(Deps{
		Log:            log,
		DB:             pinger,
		Catalog:        catalog.NewService(reader, log),
		RequestTimeout: 5 * time.Second,
		DBTimeout:      time.Second,
		MaxBodyBytes:   1 << 20,
	})
}

func TestHealthzIsIndependentOfDatabase(t *testing.T) {
	router := newTestRouter(t, stubPinger{err: errors.New("database down")}, stubProductReader{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: liveness must not depend on the database", rec.Code)
	}
}

func TestReadyzReflectsDatabaseState(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		router := newTestRouter(t, stubPinger{}, stubProductReader{})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("database unreachable", func(t *testing.T) {
		router := newTestRouter(t, stubPinger{err: errors.New("no connection")}, stubProductReader{})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})
}

func TestRequestIDIsGeneratedAndEchoed(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if got := rec.Header().Get(requestIDHeader); got == "" {
		t.Fatal("response is missing the request id header")
	}
}

func TestRequestIDFromClientIsReusedWhenSafe(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "trace-abc-123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got != "trace-abc-123" {
		t.Fatalf("request id = %q, want the client value", got)
	}
}

func TestUnsafeClientRequestIDIsReplaced(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "bad value\nwith newline")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got == "bad value\nwith newline" {
		t.Fatal("unsafe request id was echoed back")
	}
}

func TestNotFoundUsesSharedErrorEnvelope(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-such-path", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	var body errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not the shared envelope: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Errorf("code = %q, want not_found", body.Error.Code)
	}
	if body.Error.RequestID == "" {
		t.Error("envelope should carry the request id for support")
	}
}

func TestProductNotFoundIsMappedTo404(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{err: catalog.ErrProductNotFound})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/products/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	var body errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not the shared envelope: %v", err)
	}
	if body.Error.Code != "product_not_found" {
		t.Errorf("code = %q, want product_not_found", body.Error.Code)
	}
}

func TestInvalidSlugIsMappedTo400(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/products/%20", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestProductIsServedAsJSON(t *testing.T) {
	updated := time.Date(2026, 9, 17, 8, 30, 0, 0, time.UTC)
	router := newTestRouter(t, stubPinger{}, stubProductReader{product: catalog.Product{
		ID:        42,
		Slug:      "acetaminophen-500",
		NameFa:    "استامینوفن ۵۰۰",
		UpdatedAt: updated,
	}})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/products/acetaminophen-500", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("content type = %q", ct)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["slug"] != "acetaminophen-500" {
		t.Errorf("slug = %v", body["slug"])
	}
	if body["name_fa"] != "استامینوفن ۵۰۰" {
		t.Errorf("name_fa = %v: persian text must survive encoding", body["name_fa"])
	}
}

func TestReservedRouteGroupsAnswerNotImplemented(t *testing.T) {
	router := newTestRouter(t, stubPinger{}, stubProductReader{})

	for _, path := range []string{"/api/v1/ops/sources", "/api/v1/admin/users"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s: status = %d, want 501", path, rec.Code)
		}
	}
}

func TestPanicBecomesLoggedInternalError(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) {
		panic("handler exploded")
	})
	handler := chain(mux, withRequestID(), withRecover(log))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !bytes.Contains(logs.Bytes(), []byte("http.panic")) {
		t.Error("panic was not logged as http.panic")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("handler exploded")) {
		t.Error("internal panic detail leaked into the response body")
	}
}
