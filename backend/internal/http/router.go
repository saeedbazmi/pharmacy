// Package http wires transport concerns: routing, middleware and error mapping.
// No business rule lives in this package.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
)

// Deps are everything the router needs, passed in explicitly.
type Deps struct {
	Log            *slog.Logger
	DB             Pinger
	Catalog        *catalog.Service
	RequestTimeout time.Duration
	DBTimeout      time.Duration
	MaxBodyBytes   int64
}

// NewRouter builds the HTTP handler for the api service.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	writeError := ErrorWriter(d.Log)

	// Operational endpoints stay outside /api/v1 and outside the access log
	// budget of public traffic.
	mux.Handle("GET /healthz", liveness())
	mux.Handle("GET /readyz", readiness(d.DB, d.DBTimeout))

	// Public API.
	catalog.NewHandler(d.Catalog, writeError).Register(mux)

	// Route groups reserved for later milestones. They answer explicitly instead
	// of falling through to a confusing 404 once the panel work starts.
	mux.HandleFunc("/api/v1/ops/", notImplemented)
	mux.HandleFunc("/api/v1/admin/", notImplemented)

	mux.HandleFunc("/", notFound)

	return chain(mux,
		withRequestID(),
		withRecover(d.Log),
		withAccessLog(d.Log),
		withTimeout(d.RequestTimeout),
		withMaxBody(d.MaxBodyBytes),
	)
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusNotImplemented,
		"not_implemented", "این بخش هنوز پیاده‌سازی نشده است.")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusNotFound,
		"not_found", "نشانی مورد نظر یافت نشد.")
}
