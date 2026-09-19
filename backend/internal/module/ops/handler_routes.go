package ops

import (
	"encoding/json"
	"net/http"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

type pharmacyBody struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	SiteDomain string `json:"site_domain"`
	LogoURL    string `json:"logo_url"`
	Status     string `json:"status"`
}

type sourceBody struct {
	PharmacyID       int64           `json:"pharmacy_id"`
	Kind             string          `json:"kind"`
	Config           json.RawMessage `json:"config"`
	ScheduleInterval string          `json:"schedule_interval"`
	Enabled          *bool           `json:"enabled"`
}

type productBody struct {
	NameFa      *string `json:"name_fa"`
	NameEn      *string `json:"name_en"`
	GenericName *string `json:"generic_name"`
	BrandID     *int64  `json:"brand_id"`
	CategoryID  *int64  `json:"category_id"`
	ImageURL    *string `json:"image_url"`
	Status      *string `json:"status"`
}

type categoryBody struct {
	ParentID   int64  `json:"parent_id"`
	Slug       string `json:"slug"`
	NameFa     string `json:"name_fa"`
	Position   int32  `json:"position"`
	ReassignTo int64  `json:"reassign_to"`
}

type brandBody struct {
	Slug   string `json:"slug"`
	NameFa string `json:"name_fa"`
	NameEn string `json:"name_en"`
	IntoID int64  `json:"into_id"`
}

type linkBody struct {
	ProductID int64 `json:"product_id"`
}

func (h *Handler) listPharmacies(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListPharmacies(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"pharmacies": toPharmacies(items)})
}

func (h *Handler) createPharmacy(w http.ResponseWriter, r *http.Request) {
	var body pharmacyBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreatePharmacy(r.Context(), h.actor(r), PharmacyInput(body))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toPharmacy(got))
}

func (h *Handler) updatePharmacy(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body pharmacyBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdatePharmacy(r.Context(), h.actor(r), id, PharmacyInput(body))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toPharmacy(got))
}

func (h *Handler) disablePharmacy(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.DisablePharmacy(r.Context(), h.actor(r), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listSources(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSources(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sources": toSources(items)})
}

func (h *Handler) getSource(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.GetSource(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSource(got))
}

func (h *Handler) createSource(w http.ResponseWriter, r *http.Request) {
	in, err := h.readSource(w, r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreateSource(r.Context(), h.actor(r), in)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toSource(got))
}

func (h *Handler) updateSource(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	in, err := h.readSource(w, r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateSource(r.Context(), h.actor(r), id, in)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSource(got))
}

func (h *Handler) disableSource(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.DisableSource(r.Context(), h.actor(r), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) testSource(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	items, err := h.svc.TestConnection(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) testDraft(w http.ResponseWriter, r *http.Request) {
	in, err := h.readSource(w, r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	items, err := h.svc.TestDraft(r.Context(), in)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) syncSource(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.SyncNow(r.Context(), h.actor(r), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, toJob(got))
}

func (h *Handler) sourceJobs(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.SourceJobs(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toJob(got))
}

func (h *Handler) listMatches(w http.ResponseWriter, r *http.Request) {
	page, size := pageQuery(r)
	got, err := h.svc.ListMatches(r.Context(), queryInt64(r, "source_id"), queryFloat(r, "min_score"), page, size)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"matches": toMatches(got.Items), "page": got.Page, "page_size": got.Size, "total": got.Total,
	})
}

func (h *Handler) approveMatch(w http.ResponseWriter, r *http.Request) {
	h.runMatch(w, r, func(actor Actor, id int64) error {
		return h.svc.ApproveMatch(r.Context(), actor, id)
	})
}

func (h *Handler) rejectMatch(w http.ResponseWriter, r *http.Request) {
	h.runMatch(w, r, func(actor Actor, id int64) error {
		return h.svc.RejectMatch(r.Context(), actor, id)
	})
}

func (h *Handler) undoMatch(w http.ResponseWriter, r *http.Request) {
	h.runMatch(w, r, func(actor Actor, id int64) error {
		return h.svc.UndoMatch(r.Context(), actor, id)
	})
}

func (h *Handler) linkMatch(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body linkBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.LinkMatch(r.Context(), h.actor(r), id, body.ProductID); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) createFromMatch(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	productID, err := h.svc.CreateFromMatch(r.Context(), h.actor(r), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]int64{"product_id": productID})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	page, size := pageQuery(r)
	got, err := h.svc.ListProducts(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("q"), page, size)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"products": toProducts(got.Items), "page": got.Page, "page_size": got.Size, "total": got.Total,
	})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toProduct(got))
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body productBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateProduct(r.Context(), h.actor(r), id, ProductPatch{
		NameFa: body.NameFa, NameEn: body.NameEn, GenericName: body.GenericName,
		BrandID: body.BrandID, CategoryID: body.CategoryID, ImageURL: body.ImageURL, Status: body.Status,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toProduct(got))
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListCategories(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"categories": toCategories(items)})
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var body categoryBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreateCategory(r.Context(), h.actor(r), Category{ParentID: body.ParentID, Slug: body.Slug, NameFa: body.NameFa, Position: body.Position})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toCategory(got))
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body categoryBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateCategory(r.Context(), h.actor(r), id, Category{ParentID: body.ParentID, Slug: body.Slug, NameFa: body.NameFa, Position: body.Position})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toCategory(got))
}

func (h *Handler) disableCategory(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body categoryBody
	_ = decode(w, r, &body)
	if err := h.svc.DisableCategory(r.Context(), h.actor(r), id, body.ReassignTo); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listBrands(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBrands(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"brands": toBrands(items)})
}

func (h *Handler) createBrand(w http.ResponseWriter, r *http.Request) {
	var body brandBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.CreateBrand(r.Context(), h.actor(r), Brand{Slug: body.Slug, NameFa: body.NameFa, NameEn: body.NameEn})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toBrand(got))
}

func (h *Handler) updateBrand(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body brandBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	got, err := h.svc.UpdateBrand(r.Context(), h.actor(r), id, Brand{Slug: body.Slug, NameFa: body.NameFa, NameEn: body.NameEn})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toBrand(got))
}

func (h *Handler) mergeBrand(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	var body brandBody
	if err := decode(w, r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := h.svc.MergeBrands(r.Context(), h.actor(r), id, body.IntoID); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listAudit(w http.ResponseWriter, r *http.Request) {
	page, size := pageQuery(r)
	got, err := h.svc.ListAudit(r.Context(), queryInt64(r, "actor_id"), queryTime(r, "since"), queryTime(r, "until"), page, size)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"entries": toAudits(got.Items), "page": got.Page, "page_size": got.Size, "total": got.Total,
	})
}

func (h *Handler) readSource(w http.ResponseWriter, r *http.Request) (SourceInput, error) {
	var body sourceBody
	if err := decode(w, r, &body); err != nil {
		return SourceInput{}, err
	}
	in := SourceInput{
		PharmacyID: body.PharmacyID, Kind: body.Kind, Config: rawJSON(body.Config),
		ScheduleInterval: body.ScheduleInterval, Enabled: true,
	}
	if body.Enabled != nil {
		in.Enabled = *body.Enabled
	}
	return in, nil
}

func (h *Handler) runMatch(w http.ResponseWriter, r *http.Request, fn func(actor Actor, id int64) error) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := fn(h.actor(r), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
