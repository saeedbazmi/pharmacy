package catalog

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/search"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

// ErrorWriter renders a domain error as the shared HTTP error envelope. The
// implementation lives in the http layer; the handler only calls it.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

// Handler exposes the catalog over HTTP. It validates shape, calls the service
// and maps the result; no business rules live here.
type Handler struct {
	svc        *Service
	search     *search.Handler
	writeError ErrorWriter
}

// NewHandler builds the transport layer for the catalog module.
func NewHandler(svc *Service, searcher *search.Handler, writeError ErrorWriter) *Handler {
	return &Handler{svc: svc, search: searcher, writeError: writeError}
}

// Register mounts the public catalog routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/products", h.listProducts)
	mux.HandleFunc("GET /api/v1/products/{slug}", h.getProduct)
	mux.HandleFunc("GET /api/v1/sitemap", h.sitemap)
}

type productResponse struct {
	ID           int64           `json:"id"`
	Slug         string          `json:"slug"`
	NameFa       string          `json:"name_fa"`
	NameEn       string          `json:"name_en,omitempty"`
	GenericName  string          `json:"generic_name,omitempty"`
	DosageForm   string          `json:"dosage_form,omitempty"`
	Strength     string          `json:"strength,omitempty"`
	ImageURL     string          `json:"image_url,omitempty"`
	Description  string          `json:"description,omitempty"`
	BrandName    string          `json:"brand_name,omitempty"`
	CategoryName string          `json:"category_name,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Offers       []offerResponse `json:"offers"`
}

type offerResponse struct {
	ID           int64     `json:"id"`
	PriceRial    int64     `json:"price_rial"`
	InStock      bool      `json:"in_stock"`
	ProductURL   string    `json:"product_url"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	PharmacyName string    `json:"pharmacy_name"`
	PharmacySlug string    `json:"pharmacy_slug"`
	BestPrice    bool      `json:"best_price"`
	Stale        bool      `json:"stale"`
}

type productSummaryResponse struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	NameFa          string `json:"name_fa"`
	ImageURL        string `json:"image_url,omitempty"`
	BrandName       string `json:"brand_name,omitempty"`
	LowestPriceRial int64  `json:"lowest_price_rial"`
	OfferCount      int    `json:"offer_count"`
	InStockCount    int    `json:"in_stock_count"`
}

type productListResponse struct {
	Products []productSummaryResponse `json:"products"`
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	filter := ParseOfferFilter(r.URL.Query())
	product, err := h.svc.ProductBySlug(r.Context(), r.PathValue("slug"), filter)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	offers := make([]offerResponse, 0, len(product.Offers))
	for _, o := range product.Offers {
		offers = append(offers, offerResponse{
			ID:           o.ID,
			PriceRial:    o.PriceRial,
			InStock:      o.InStock,
			ProductURL:   o.ProductURL,
			LastSeenAt:   o.LastSeenAt,
			PharmacyName: o.PharmacyName,
			PharmacySlug: o.PharmacySlug,
			BestPrice:    o.BestPrice,
			Stale:        o.Stale,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, productResponse{
		ID:           product.ID,
		Slug:         product.Slug,
		NameFa:       product.NameFa,
		NameEn:       product.NameEn,
		GenericName:  product.GenericName,
		DosageForm:   product.DosageForm,
		Strength:     product.Strength,
		ImageURL:     product.ImageURL,
		Description:  product.Description,
		BrandName:    product.BrandName,
		CategoryName: product.CategoryName,
		UpdatedAt:    product.UpdatedAt,
		Offers:       offers,
	})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if strings.TrimSpace(q.Get("q")) != "" {
		if h.search == nil {
			h.writeError(w, r, search.ErrQueryEmpty)
			return
		}
		h.search.Search(w, r)
		return
	}
	if cat := strings.TrimSpace(q.Get("category")); cat != "" {
		h.listByCategory(w, r, cat)
		return
	}
	var (
		items []ProductSummary
		err   error
	)
	if raw := strings.TrimSpace(q.Get("slugs")); raw != "" {
		items, err = h.svc.ProductsBySlugs(r.Context(), strings.Split(raw, ","))
	} else {
		limit := 48
		if v := q.Get("limit"); v != "" {
			if n, convErr := strconv.Atoi(v); convErr == nil {
				limit = n
			}
		}
		items, err = h.svc.ListPublished(r.Context(), limit)
	}
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]productSummaryResponse, 0, len(items))
	for _, p := range items {
		out = append(out, productSummaryResponse{
			ID:              p.ID,
			Slug:            p.Slug,
			NameFa:          p.NameFa,
			ImageURL:        p.ImageURL,
			BrandName:       p.BrandName,
			LowestPriceRial: p.LowestPriceRial,
			OfferCount:      p.OfferCount,
			InStockCount:    p.InStockCount,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, productListResponse{Products: out})
}

func (h *Handler) listByCategory(w http.ResponseWriter, r *http.Request, slug string) {
	limit := 48
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, convErr := strconv.Atoi(v); convErr == nil {
			limit = n
		}
	}
	category, items, err := h.svc.ListByCategory(r.Context(), slug, limit)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]productSummaryResponse, 0, len(items))
	for _, p := range items {
		out = append(out, productSummaryResponse{
			ID:              p.ID,
			Slug:            p.Slug,
			NameFa:          p.NameFa,
			ImageURL:        p.ImageURL,
			BrandName:       p.BrandName,
			LowestPriceRial: p.LowestPriceRial,
			OfferCount:      p.OfferCount,
			InStockCount:    p.InStockCount,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"products": out,
		"category": map[string]any{"slug": category.Slug, "name_fa": category.NameFa},
	})
}

func (h *Handler) sitemap(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSitemap(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"slug": item.Slug, "updated_at": item.UpdatedAt})
	}
	cats, err := h.svc.ListCategorySitemap(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	catOut := make([]map[string]any, 0, len(cats))
	for _, item := range cats {
		catOut = append(catOut, map[string]any{"slug": item.Slug, "updated_at": item.UpdatedAt})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"products": out, "categories": catOut})
}
