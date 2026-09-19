package ops

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
)

func (h *Handler) listHealth(w http.ResponseWriter, r *http.Request) {
	window := queryDays(r)
	items, err := h.svc.ListHealth(r.Context(), window)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if window <= 0 {
		window = defaultHealthWindow
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"sources":      toHealthList(items),
		"window_hours": int(window.Hours()),
	})
}

func (h *Handler) sourceRuns(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	runs, err := h.svc.SourceRuns(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	out := make([]runDTO, 0, len(runs))
	for _, run := range runs {
		out = append(out, runDTO{
			ID: run.ID, Status: run.Status, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt,
			OkCount: run.OkCount, FailCount: run.FailCount, Error: run.Error,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"runs": out})
}

func (h *Handler) listPrices(w http.ResponseWriter, r *http.Request) {
	page, size := pageQuery(r)
	got, err := h.svc.ListSuspicious(r.Context(), page, size)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"offers": toSuspiciousList(got.Items), "page": got.Page, "page_size": got.Size, "total": got.Total,
	})
}

func (h *Handler) approvePrice(w http.ResponseWriter, r *http.Request) {
	h.runPrice(w, r, h.svc.ApprovePrice)
}

func (h *Handler) rejectPrice(w http.ResponseWriter, r *http.Request) {
	h.runPrice(w, r, h.svc.RejectPrice)
}

func (h *Handler) runPrice(w http.ResponseWriter, r *http.Request, fn func(context.Context, Actor, int64) error) {
	id, err := h.pathID(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if err := fn(r.Context(), h.actor(r), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listStale(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListStaleSources(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sources": toStaleList(items)})
}

func (h *Handler) listClicks(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.ClickReport(r.Context(), queryTime(r, "from"), queryTime(r, "to"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, clickReportDTO(report))
}

func (h *Handler) clicksCSV(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.ClickReport(r.Context(), queryTime(r, "from"), queryTime(r, "to"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="redirect-clicks.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"pharmacy_id", "slug", "name", "clicks", "from", "to"})
	from := report.Since.Format(time.RFC3339)
	to := report.Until.Format(time.RFC3339)
	for _, ph := range report.Pharmacies {
		_ = cw.Write([]string{
			strconv.FormatInt(ph.ID, 10), ph.Slug, ph.Name,
			strconv.FormatInt(ph.Clicks, 10), from, to,
		})
	}
	cw.Flush()
}

func clickReportDTO(report ClickReport) map[string]any {
	pharmacies := make([]clickPharmacyDTO, 0, len(report.Pharmacies))
	for _, ph := range report.Pharmacies {
		pharmacies = append(pharmacies, clickPharmacyDTO(ph))
	}
	products := make([]clickProductDTO, 0, len(report.TopProducts))
	for _, p := range report.TopProducts {
		products = append(products, clickProductDTO(p))
	}
	var top *clickPharmacyDTO
	if report.TopPharmacy != nil {
		v := clickPharmacyDTO(*report.TopPharmacy)
		top = &v
	}
	return map[string]any{
		"from": report.Since, "to": report.Until, "total": report.Total,
		"pharmacies": pharmacies, "top_products": products, "top_pharmacy": top,
	}
}

func queryDays(r *http.Request) time.Duration {
	n, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil || n < 1 || n > 365 {
		return 0
	}
	return time.Duration(n) * 24 * time.Hour
}
