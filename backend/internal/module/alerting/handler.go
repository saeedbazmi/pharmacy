package alerting

import (
	"net/http"
	"strconv"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

type Handler struct {
	svc        *Service
	writeError ErrorWriter
}

func NewHandler(svc *Service, writeError ErrorWriter) *Handler {
	return &Handler{svc: svc, writeError: writeError}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/account/favorites", h.listFavorites)
	mux.HandleFunc("POST /api/v1/account/favorites", h.addFavorite)
	mux.HandleFunc("GET /api/v1/account/favorites/{productId}", h.favoriteExists)
	mux.HandleFunc("DELETE /api/v1/account/favorites/{productId}", h.removeFavorite)
	mux.HandleFunc("GET /api/v1/account/alerts", h.listAlerts)
	mux.HandleFunc("POST /api/v1/account/alerts", h.createAlert)
	mux.HandleFunc("DELETE /api/v1/account/alerts/{id}", h.deleteAlert)
}

func (h *Handler) listFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListFavorites(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row := map[string]any{
			"product_id":        item.ProductID,
			"slug":              item.Slug,
			"name_fa":           item.NameFa,
			"lowest_price_rial": item.LowestPriceRial,
			"offer_count":       item.OfferCount,
			"created_at":        item.CreatedAt,
		}
		if item.ImageURL != "" {
			row["image_url"] = item.ImageURL
		}
		out = append(out, row)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"favorites": out})
}

func (h *Handler) addFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	var body struct {
		ProductID int64 `json:"product_id"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<12); err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.AddFavorite(r.Context(), userID, body.ProductID); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) favoriteExists(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	id, err := pathID(r, "productId")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	found, err := h.svc.FavoriteExists(r.Context(), userID, id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"favorite": found})
}

func (h *Handler) removeFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	id, err := pathID(r, "productId")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.RemoveFavorite(r.Context(), userID, id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListAlerts(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row := map[string]any{
			"id":                  item.ID,
			"product_id":          item.ProductID,
			"kind":                item.Kind,
			"baseline_price_rial": item.BaselinePriceRial,
			"status":              item.Status,
			"created_at":          item.CreatedAt,
			"product_slug":        item.ProductSlug,
			"product_name":        item.ProductName,
		}
		if item.TargetPriceRial != nil {
			row["target_price_rial"] = *item.TargetPriceRial
		}
		out = append(out, row)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"alerts": out})
}

func (h *Handler) createAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	var body struct {
		ProductID       int64  `json:"product_id"`
		Kind            string `json:"kind"`
		TargetPriceRial *int64 `json:"target_price_rial"`
	}
	if err := httpx.ReadJSON(w, r, &body, 1<<12); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreateAlert(r.Context(), userID, AlertInput{
		ProductID: body.ProductID, Kind: body.Kind, TargetPriceRial: body.TargetPriceRial,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"id": got.ID, "product_id": got.ProductID, "kind": got.Kind, "status": got.Status,
	})
}

func (h *Handler) deleteAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.DeleteAlert(r.Context(), userID, id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	u, ok := reqctx.UserFrom(r.Context())
	if !ok {
		h.writeError(w, r, ErrUnauthorized)
		return 0, false
	}
	return u.ID, true
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}
	return id, nil
}
