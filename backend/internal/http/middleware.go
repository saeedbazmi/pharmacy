package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/identity"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

const requestIDHeader = "X-Request-Id"

// middleware is a handler decorator.
type middleware func(http.Handler) http.Handler

// chain applies middlewares so the first argument is the outermost layer.
func chain(h http.Handler, ms ...middleware) http.Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

// withRequestID puts a request id in the context and echoes it back, so a user
// report can be traced to exact log lines.
func withRequestID() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestIDHeader)
			if !isSafeRequestID(id) {
				id = newRequestID()
			}
			w.Header().Set(requestIDHeader, id)
			next.ServeHTTP(w, r.WithContext(reqctx.WithRequestID(r.Context(), id)))
		})
	}
}

// withRecover turns a panic into a logged 500 instead of a dropped connection.
func withRecover(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				log.ErrorContext(r.Context(), "http.panic",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()))
				writeProblem(w, r, http.StatusInternalServerError,
					"internal_error", "خطای غیرمنتظره‌ای رخ داد. لطفاً دوباره تلاش کنید.")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// withAccessLog records one line per request at info level.
func withAccessLog(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rec := httpx.NewStatusRecorder(w)

			next.ServeHTTP(rec, r)

			log.InfoContext(r.Context(), "http.request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.Status(),
				"bytes", rec.Bytes(),
				"duration_ms", time.Since(started).Milliseconds())
		})
	}
}

// withTimeout bounds every request so a slow dependency cannot pile up
// connections. Handlers observe the deadline through the request context.
func withTimeout(d time.Duration) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// withMaxBody caps the request body before any handler reads it.
func withMaxBody(maxBytes int64) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func withUser(h *identity.Handler) middleware {
	return func(next http.Handler) http.Handler {
		if h == nil {
			return next
		}
		return h.Attach(next)
	}
}

func newRequestID() string {
	var buf [16]byte
	// rand.Read from crypto/rand never fails as of Go 1.24.
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

// isSafeRequestID accepts a caller supplied id only when it cannot pollute logs
// or response headers.
func isSafeRequestID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}
