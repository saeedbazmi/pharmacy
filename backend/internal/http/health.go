package http

import (
	"context"
	"net/http"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

// Pinger is the readiness dependency: anything that can be checked cheaply.
type Pinger interface {
	Ping(ctx context.Context) error
}

// liveness answers whether the process is up. It must not touch dependencies,
// otherwise a database blip would make the orchestrator restart a healthy app.
func liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// readiness answers whether the process can serve traffic right now.
func readiness(db Pinger, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":   "unavailable",
				"database": "unreachable",
			})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":   "ok",
			"database": "ok",
		})
	}
}
