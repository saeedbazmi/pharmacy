package search

import (
	"net/http"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

type Handler struct {
	svc        *Service
	writeError ErrorWriter
}

func NewHandler(svc *Service, writeError ErrorWriter) *Handler {
	return &Handler{svc: svc, writeError: writeError}
}

type hitResponse struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	NameFa          string `json:"name_fa"`
	ImageURL        string `json:"image_url,omitempty"`
	BrandName       string `json:"brand_name,omitempty"`
	LowestPriceRial int64  `json:"lowest_price_rial"`
	OfferCount      int    `json:"offer_count"`
	InStockCount    int    `json:"in_stock_count"`
}

type facetResponse struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type listResponse struct {
	Products   []hitResponse   `json:"products"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int             `json:"total"`
	Brands     []facetResponse `json:"brands,omitempty"`
	Categories []facetResponse `json:"categories,omitempty"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q, err := parseQuery(r.URL.Query())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	res, err := h.svc.Search(r.Context(), q)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := listResponse{
		Products:   make([]hitResponse, 0, len(res.Hits)),
		Page:       res.Page,
		PageSize:   res.PageSize,
		Total:      res.Total,
		Brands:     make([]facetResponse, 0, len(res.Brands)),
		Categories: make([]facetResponse, 0, len(res.Categories)),
	}
	for _, hit := range res.Hits {
		out.Products = append(out.Products, hitResponse{
			ID:              hit.ID,
			Slug:            hit.Slug,
			NameFa:          hit.NameFa,
			ImageURL:        hit.ImageURL,
			BrandName:       hit.BrandName,
			LowestPriceRial: hit.LowestPriceRial,
			OfferCount:      hit.OfferCount,
			InStockCount:    hit.InStockCount,
		})
	}
	for _, f := range res.Brands {
		out.Brands = append(out.Brands, facetResponse{Slug: f.Slug, Name: f.Name})
	}
	for _, f := range res.Categories {
		out.Categories = append(out.Categories, facetResponse{Slug: f.Slug, Name: f.Name})
	}
	w.Header().Set("Cache-Control", "public, max-age=30")
	httpx.WriteJSON(w, http.StatusOK, out)
}
