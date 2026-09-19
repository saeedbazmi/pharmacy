package redirect

import "net/http"

// Handler serves GET /go/{offerId}.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /go/{offerId}", h.goOffer)
}

func (h *Handler) goOffer(w http.ResponseWriter, r *http.Request) {
	target, err := h.svc.GoTarget(r.Context(), r.PathValue("offerId"), r.Referer())
	if err != nil {
		writeGone(w)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	http.Redirect(w, r, target.ProductURL, http.StatusFound)
}

func writeGone(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="fa" dir="rtl">
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>پیشنهاد نامعتبر</title>
<body>
<main>
<h1>این پیشنهاد دیگر معتبر نیست</h1>
<p>لینک خرید منقضی شده یا این کالا در داروخانه موجود نیست.</p>
<p><a href="/">بازگشت به صفحه اصلی</a></p>
</main>
</body>
</html>`))
}
