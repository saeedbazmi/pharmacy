// Package http wires transport concerns: routing, middleware and error mapping.
// No business rule lives in this package.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/admin"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/alerting"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/identity"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/redirect"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/search"
)

// Deps are everything the router needs, passed in explicitly.
type Deps struct {
	Log            *slog.Logger
	DB             Pinger
	Catalog        *catalog.Service
	Search         *search.Handler
	Redirect       *redirect.Handler
	Ops            *ops.Handler
	Identity       *identity.Handler
	Alerting       *alerting.Handler
	Admin          *admin.Handler
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
	catalog.NewHandler(d.Catalog, d.Search, writeError).Register(mux)
	if d.Redirect != nil {
		d.Redirect.Register(mux)
	}

	if d.Ops != nil {
		d.Ops.Register(mux)
	} else {
		mux.HandleFunc("/api/v1/ops/", unauthorizedOps)
	}
	if d.Identity != nil {
		d.Identity.Register(mux)
	}
	if d.Alerting != nil {
		d.Alerting.Register(mux)
	}
	if d.Admin != nil {
		d.Admin.Register(mux)
	} else {
		mux.HandleFunc("/api/v1/admin/", unauthorizedOps)
	}

	mux.HandleFunc("/", notFound)

	return chain(mux,
		withRequestID(),
		withRecover(d.Log),
		withAccessLog(d.Log),
		withTimeout(d.RequestTimeout),
		withMaxBody(d.MaxBodyBytes),
		withUser(d.Identity),
	)
}

func unauthorizedOps(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusUnauthorized,
		"unauthorized", "برای ادامه وارد شوید.")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusNotFound,
		"not_found", "نشانی مورد نظر یافت نشد.")
}
