package catalog

import (
	"net/http"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

// ErrorWriter renders a domain error as the shared HTTP error envelope. The
// implementation lives in the http layer; the handler only calls it.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

// Handler exposes the catalog over HTTP. It validates shape, calls the service
// and maps the result; no business rules live here.
type Handler struct {
	svc        *Service
	writeError ErrorWriter
}

// NewHandler builds the transport layer for the catalog module.
func NewHandler(svc *Service, writeError ErrorWriter) *Handler {
	return &Handler{svc: svc, writeError: writeError}
}

// Register mounts the public catalog routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/products/{slug}", h.getProduct)
}

// productResponse is the wire shape. Domain types carry no json tags, so the
// transport format can change without touching the domain.
type productResponse struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	NameFa       string    `json:"name_fa"`
	NameEn       string    `json:"name_en,omitempty"`
	GenericName  string    `json:"generic_name,omitempty"`
	DosageForm   string    `json:"dosage_form,omitempty"`
	Strength     string    `json:"strength,omitempty"`
	ImageURL     string    `json:"image_url,omitempty"`
	Description  string    `json:"description,omitempty"`
	BrandName    string    `json:"brand_name,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.ProductBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		h.writeError(w, r, err)
		return
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
	})
}
