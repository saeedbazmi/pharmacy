package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/httpx"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

// errorEnvelope is the single error shape every endpoint returns.
type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// mapping translates a domain error into a status, a stable machine code and a
// Persian message safe to show a user. Adding a domain error means adding one
// row here, nothing else.
type mapping struct {
	err     error
	status  int
	code    string
	message string
}

var mappings = []mapping{
	{catalog.ErrProductNotFound, http.StatusNotFound, "product_not_found", "کالای مورد نظر یافت نشد."},
	{catalog.ErrInvalidSlug, http.StatusBadRequest, "invalid_product_slug", "شناسه کالا معتبر نیست."},
	{httpx.ErrBodyTooLarge, http.StatusRequestEntityTooLarge, "body_too_large", "حجم درخواست بیش از حد مجاز است."},
	{context.DeadlineExceeded, http.StatusGatewayTimeout, "timeout", "پاسخ‌گویی بیش از حد انتظار طول کشید. دوباره تلاش کنید."},
	{context.Canceled, 499, "client_closed_request", "درخواست پیش از تکمیل لغو شد."},
}

// ErrorWriter renders errors using the shared envelope and logs the ones that
// need human attention. This is the only place errors become HTTP responses.
func ErrorWriter(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		ctx := r.Context()

		status, code, message := http.StatusInternalServerError, "internal_error", "خطای غیرمنتظره‌ای رخ داد. لطفاً دوباره تلاش کنید."
		for _, m := range mappings {
			if errors.Is(err, m.err) {
				status, code, message = m.status, m.code, m.message
				break
			}
		}

		// Unmapped errors are ours to fix, so they are logged with detail.
		// Mapped client errors are expected traffic and stay at debug level.
		if status >= http.StatusInternalServerError {
			log.ErrorContext(ctx, "http.error",
				"code", code, "status", status, "method", r.Method, "path", r.URL.Path, "error", err.Error())
		} else {
			log.DebugContext(ctx, "http.client_error",
				"code", code, "status", status, "method", r.Method, "path", r.URL.Path, "error", err.Error())
		}

		httpx.WriteJSON(w, status, errorEnvelope{Error: errorPayload{
			Code:      code,
			Message:   message,
			RequestID: reqctx.RequestID(ctx),
		}})
	}
}

// writeProblem renders an error that has no domain error behind it.
func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	httpx.WriteJSON(w, status, errorEnvelope{Error: errorPayload{
		Code:      code,
		Message:   message,
		RequestID: reqctx.RequestID(r.Context()),
	}})
}
