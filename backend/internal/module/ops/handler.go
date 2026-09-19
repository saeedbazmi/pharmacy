package ops

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

// ErrorWriter renders a domain error as the shared HTTP envelope.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

// Handler is the transport layer for /api/v1/ops.
type Handler struct {
	svc        *Service
	writeError ErrorWriter
	secure     bool
}

func NewHandler(svc *Service, writeError ErrorWriter, secureCookie bool) *Handler {
	return &Handler{svc: svc, writeError: writeError, secure: secureCookie}
}

// Register mounts public login/logout and the protected ops tree.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/ops/login", h.login)
	mux.HandleFunc("POST /api/v1/ops/logout", h.logout)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/ops/me", h.me)
	protected.HandleFunc("GET /api/v1/ops/pharmacies", h.listPharmacies)
	protected.HandleFunc("POST /api/v1/ops/pharmacies", h.createPharmacy)
	protected.HandleFunc("PATCH /api/v1/ops/pharmacies/{id}", h.updatePharmacy)
	protected.HandleFunc("POST /api/v1/ops/pharmacies/{id}/disable", h.disablePharmacy)
	protected.HandleFunc("GET /api/v1/ops/sources", h.listSources)
	protected.HandleFunc("POST /api/v1/ops/sources", h.createSource)
	protected.HandleFunc("POST /api/v1/ops/sources/test", h.testDraft)
	protected.HandleFunc("GET /api/v1/ops/sources/{id}", h.getSource)
	protected.HandleFunc("PATCH /api/v1/ops/sources/{id}", h.updateSource)
	protected.HandleFunc("POST /api/v1/ops/sources/{id}/disable", h.disableSource)
	protected.HandleFunc("POST /api/v1/ops/sources/{id}/test", h.testSource)
	protected.HandleFunc("POST /api/v1/ops/sources/{id}/sync", h.syncSource)
	protected.HandleFunc("GET /api/v1/ops/sources/{id}/jobs", h.sourceJobs)
	protected.HandleFunc("GET /api/v1/ops/matches", h.listMatches)
	protected.HandleFunc("POST /api/v1/ops/matches/{id}/approve", h.approveMatch)
	protected.HandleFunc("POST /api/v1/ops/matches/{id}/reject", h.rejectMatch)
	protected.HandleFunc("POST /api/v1/ops/matches/{id}/link", h.linkMatch)
	protected.HandleFunc("POST /api/v1/ops/matches/{id}/create-product", h.createFromMatch)
	protected.HandleFunc("POST /api/v1/ops/matches/{id}/undo", h.undoMatch)
	protected.HandleFunc("GET /api/v1/ops/products", h.listProducts)
	protected.HandleFunc("GET /api/v1/ops/products/{id}", h.getProduct)
	protected.HandleFunc("PATCH /api/v1/ops/products/{id}", h.updateProduct)
	protected.HandleFunc("GET /api/v1/ops/categories", h.listCategories)
	protected.HandleFunc("POST /api/v1/ops/categories", h.createCategory)
	protected.HandleFunc("PATCH /api/v1/ops/categories/{id}", h.updateCategory)
	protected.HandleFunc("POST /api/v1/ops/categories/{id}/disable", h.disableCategory)
	protected.HandleFunc("GET /api/v1/ops/brands", h.listBrands)
	protected.HandleFunc("POST /api/v1/ops/brands", h.createBrand)
	protected.HandleFunc("PATCH /api/v1/ops/brands/{id}", h.updateBrand)
	protected.HandleFunc("POST /api/v1/ops/brands/{id}/merge", h.mergeBrand)
	protected.HandleFunc("GET /api/v1/ops/audit", h.listAudit)
	protected.HandleFunc("GET /api/v1/ops/health", h.listHealth)
	protected.HandleFunc("GET /api/v1/ops/health/sources/{id}/runs", h.sourceRuns)
	protected.HandleFunc("GET /api/v1/ops/prices", h.listPrices)
	protected.HandleFunc("POST /api/v1/ops/prices/{id}/approve", h.approvePrice)
	protected.HandleFunc("POST /api/v1/ops/prices/{id}/reject", h.rejectPrice)
	protected.HandleFunc("GET /api/v1/ops/reports/stale", h.listStale)
	protected.HandleFunc("GET /api/v1/ops/reports/clicks.csv", h.clicksCSV)
	protected.HandleFunc("GET /api/v1/ops/reports/clicks", h.listClicks)
	protected.HandleFunc("/api/v1/ops/", h.opsNotFound)

	mux.Handle("/api/v1/ops/", h.requireOps(protected))
}

func (h *Handler) requireOps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/ops/login" || r.URL.Path == "/api/v1/ops/logout" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(SessionCookie)
		if err != nil || cookie.Value == "" {
			h.writeError(w, r, ErrUnauthorized)
			return
		}
		actor, err := h.svc.SessionActor(r.Context(), cookie.Value)
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		ctx := reqctx.WithActor(r.Context(), reqctx.Actor{
			ID: actor.ID, Username: actor.Username, Role: actor.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) actor(r *http.Request) Actor {
	a, ok := reqctx.ActorFrom(r.Context())
	if !ok {
		return Actor{}
	}
	return Actor{ID: a.ID, Username: a.Username, Role: a.Role}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<16); err != nil {
		h.writeError(w, r, err)
		return
	}
	session, err := h.svc.Login(r.Context(), body.Username, body.Password, clientIP(r))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
	httpx.WriteJSON(w, http.StatusOK, actorDTO(session.Actor))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var token string
	if c, err := r.Cookie(SessionCookie); err == nil {
		token = c.Value
	}
	_ = h.svc.Logout(r.Context(), token)
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode,
	})
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, actorDTO(h.actor(r)))
}

func (h *Handler) opsNotFound(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, r, ErrNotFound)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h *Handler) pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}
	return id, nil
}

func actorDTO(a Actor) map[string]any {
	return map[string]any{"id": a.ID, "username": a.Username, "role": a.Role}
}

func pageQuery(r *http.Request) (page, size int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	size, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	return page, size
}

func writePage[T any](w http.ResponseWriter, p Page[T], wrap string) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		wrap: p.Items, "page": p.Page, "page_size": p.Size, "total": p.Total,
	})
}

func decode[T any](w http.ResponseWriter, r *http.Request, dst *T) error {
	return httpx.ReadJSON(w, r, dst, 1<<20)
}

func rawJSON(v json.RawMessage) json.RawMessage {
	if len(v) == 0 {
		return json.RawMessage(`{}`)
	}
	return v
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func queryInt64(r *http.Request, key string) int64 {
	v, _ := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	return v
}

func queryFloat(r *http.Request, key string) float64 {
	v, _ := strconv.ParseFloat(r.URL.Query().Get(key), 64)
	return v
}

func queryTime(r *http.Request, key string) time.Time {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t
	}
	return time.Time{}
}
