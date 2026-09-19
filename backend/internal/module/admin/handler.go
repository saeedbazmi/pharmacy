package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

// ErrorWriter renders a domain error as the shared HTTP envelope.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

// Handler is the transport layer for /api/v1/admin and the public site payload.
type Handler struct {
	svc        *Service
	sessions   *ops.Service
	writeError ErrorWriter
}

func NewHandler(svc *Service, sessions *ops.Service, writeError ErrorWriter) *Handler {
	return &Handler{svc: svc, sessions: sessions, writeError: writeError}
}

// Register mounts the public site payload and the super_admin tree.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/site", h.publicSite)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/admin/dashboard", h.dashboard)
	protected.HandleFunc("GET /api/v1/admin/users", h.listUsers)
	protected.HandleFunc("POST /api/v1/admin/users", h.createUser)
	protected.HandleFunc("PATCH /api/v1/admin/users/{id}", h.updateUser)
	protected.HandleFunc("GET /api/v1/admin/settings", h.getSettings)
	protected.HandleFunc("PUT /api/v1/admin/settings", h.putSettings)
	protected.HandleFunc("GET /api/v1/admin/flags", h.listFlags)
	protected.HandleFunc("PUT /api/v1/admin/flags", h.putFlag)
	protected.HandleFunc("/api/v1/admin/", h.adminNotFound)
	mux.Handle("/api/v1/admin/", h.requireSuperAdmin(protected))
}

func (h *Handler) requireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(ops.SessionCookie)
		if err != nil || cookie.Value == "" {
			h.writeError(w, r, ErrUnauthorized)
			return
		}
		actor, err := h.sessions.SessionActor(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, ops.ErrUnauthorized) {
				h.writeError(w, r, ErrUnauthorized)
				return
			}
			if errors.Is(err, ops.ErrForbidden) {
				h.writeError(w, r, ErrForbidden)
				return
			}
			h.writeError(w, r, err)
			return
		}
		if actor.Role != RoleSuperAdmin {
			h.writeError(w, r, ErrForbidden)
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

func (h *Handler) adminNotFound(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, r, ErrNotFound)
}

func (h *Handler) publicSite(w http.ResponseWriter, r *http.Request) {
	got, err := h.svc.PublicSite(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, got)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	got, err := h.svc.Dashboard(r.Context(), days)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"from": got.From, "to": got.To, "days": got.Days,
		"searches": got.Searches, "clicks": got.Clicks, "ctr": got.CTR,
		"coverage": map[string]any{
			"active_pharmacies":        got.Coverage.ActivePharmacies,
			"published_products":       got.Coverage.PublishedProducts,
			"median_freshness_seconds": got.Coverage.MedianFreshnessSeconds,
		},
		"pharmacies": toPharmacyClicks(got.Pharmacies),
		"frequent":   toTerms(got.Frequent),
		"zero":       toTerms(got.Zero),
		"series":     toSeries(got.Series),
	})
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListUsers(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": toUsers(items)})
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<16); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreateUser(r.Context(), h.actor(r), body.Username, body.Password, body.Role)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toUser(got))
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, r, ErrInvalidInput)
		return
	}
	var body struct {
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
		Password string `json:"password"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<16); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateUser(r.Context(), h.actor(r), id, UserPatch{
		Role: body.Role, IsActive: body.IsActive, Password: body.Password,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toUser(got))
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	got, err := h.svc.GetSettings(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, got)
}

func (h *Handler) putSettings(w http.ResponseWriter, r *http.Request) {
	var body Settings
	if err := httpx.ReadJSON(w, r, &body, 1<<20); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateSettings(r.Context(), h.actor(r), body)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, got)
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	flags, pharmacies, err := h.svc.ListFlags(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"flags":      toFlags(flags),
		"pharmacies": toPharmacies(pharmacies),
	})
}

func (h *Handler) putFlag(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key        string `json:"key"`
		PharmacyID int64  `json:"pharmacy_id"`
		Enabled    bool   `json:"enabled"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<16); err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.SetFlag(r.Context(), h.actor(r), body.Key, body.PharmacyID, body.Enabled); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func toUser(u InternalUser) map[string]any {
	return map[string]any{
		"id": u.ID, "username": u.Username, "role": u.Role,
		"is_active": u.IsActive, "created_at": u.CreatedAt,
	}
}

func toUsers(items []InternalUser) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, u := range items {
		out = append(out, toUser(u))
	}
	return out
}

func toTerms(items []SearchTerm) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, t := range items {
		out = append(out, map[string]any{"query": t.Query, "hits": t.Hits, "last_seen": t.LastSeen})
	}
	return out
}

func toPharmacyClicks(items []PharmacyClicks) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{"id": p.ID, "slug": p.Slug, "name": p.Name, "clicks": p.Clicks})
	}
	return out
}

func toSeries(items []DayPoint) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{"day": p.Day, "searches": p.Searches, "clicks": p.Clicks})
	}
	return out
}

func toFlags(items []Flag) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, f := range items {
		row := map[string]any{"key": f.Key, "enabled": f.Enabled, "updated_at": f.UpdatedAt}
		if f.PharmacyID > 0 {
			row["pharmacy_id"] = f.PharmacyID
		}
		out = append(out, row)
	}
	return out
}

func toPharmacies(items []PharmacyBrief) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{"id": p.ID, "slug": p.Slug, "name": p.Name})
	}
	return out
}
