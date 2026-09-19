package identity

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

// ErrorWriter renders a domain error as the shared HTTP envelope.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

// Handler is the transport layer for public auth and account search history.
type Handler struct {
	svc        *Service
	writeError ErrorWriter
	secure     bool
}

func NewHandler(svc *Service, writeError ErrorWriter, secureCookie bool) *Handler {
	return &Handler{svc: svc, writeError: writeError, secure: secureCookie}
}

// Register mounts public OTP routes and the authenticated search-history routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/otp/request", h.requestOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/verify", h.verifyOTP)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.me)
	mux.HandleFunc("GET /api/v1/account/searches", h.listSearches)
	mux.HandleFunc("DELETE /api/v1/account/searches", h.deleteSearches)
}

// Attach loads the public user from the session cookie when present. Missing
// or invalid cookies leave the request as a guest so core flows stay open.
func (h *Handler) Attach(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(SessionCookie)
		if err != nil || c.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		user, err := h.svc.UserFromSession(r.Context(), c.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := reqctx.WithUser(r.Context(), reqctx.User{
			ID: user.ID, Phone: user.Phone, Role: user.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone string `json:"phone"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<12); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.RequestOTP(r.Context(), body.Phone, clientIP(r))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":                   true,
		"phone_masked":         got.PhoneMasked,
		"resend_after_seconds": int(got.ResendAfter.Seconds()),
	})
}

func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<12); err != nil {
		h.writeError(w, r, err)
		return
	}
	session, err := h.svc.VerifyOTP(r.Context(), body.Phone, body.Code, clientIP(r))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	h.setSessionCookie(w, session)
	httpx.WriteJSON(w, http.StatusOK, userDTO(session.User))
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
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, userDTO(user))
}

func (h *Handler) listSearches(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListSearches(r.Context(), user.ID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"id":      item.ID,
			"at":      item.At,
			"query":   item.Query,
			"results": item.Results,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"searches": out})
}

func (h *Handler) deleteSearches(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteSearches(r.Context(), user.ID); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) currentUser(w http.ResponseWriter, r *http.Request) (User, bool) {
	if u, ok := reqctx.UserFrom(r.Context()); ok {
		return User{ID: u.ID, Phone: u.Phone, Role: u.Role}, true
	}
	h.writeError(w, r, ErrUnauthorized)
	return User{}, false
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, session Session) {
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
}

func userDTO(u User) map[string]any {
	return map[string]any{
		"id":           u.ID,
		"phone_masked": MaskPhone(u.Phone),
		"role":         u.Role,
	}
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
