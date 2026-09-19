package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/saeedbazmi/pharmacy/backend/internal/module/admin"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/alerting"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/catalog"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/identity"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/ops"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/redirect"
	"github.com/saeedbazmi/pharmacy/backend/internal/module/search"
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
	{catalog.ErrCategoryNotFound, http.StatusNotFound, "category_not_found", "دسته‌بندی یافت نشد."},
	{search.ErrQueryEmpty, http.StatusBadRequest, "search_query_empty", "عبارت جست‌وجو را بنویسید."},
	{search.ErrQueryTooLong, http.StatusBadRequest, "search_query_too_long", "عبارت جست‌وجو بیش از حد طولانی است."},
	{redirect.ErrOfferNotFound, http.StatusNotFound, "offer_not_found", "این پیشنهاد دیگر معتبر نیست."},
	{redirect.ErrInvalidOfferID, http.StatusNotFound, "offer_not_found", "این پیشنهاد دیگر معتبر نیست."},
	{ops.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", "برای ادامه وارد شوید."},
	{ops.ErrForbidden, http.StatusForbidden, "forbidden", "به این بخش دسترسی ندارید."},
	{ops.ErrInvalidLogin, http.StatusUnauthorized, "invalid_login", "نام کاربری یا گذرواژه نادرست است."},
	{ops.ErrRateLimited, http.StatusTooManyRequests, "rate_limited", "تعداد تلاش ورود بیش از حد مجاز است. کمی بعد دوباره تلاش کنید."},
	{ops.ErrInactiveUser, http.StatusForbidden, "inactive_user", "حساب کاربری غیرفعال است."},
	{ops.ErrInvalidInput, http.StatusBadRequest, "invalid_input", "ورودی معتبر نیست."},
	{ops.ErrNotFound, http.StatusNotFound, "not_found", "مورد نظر یافت نشد."},
	{ops.ErrConflict, http.StatusConflict, "conflict", "این تغییر با دادهٔ موجود سازگار نیست."},
	{ops.ErrHasChildren, http.StatusConflict, "has_children", "این دسته زیردسته دارد و حذف نمی‌شود."},
	{ops.ErrHasProducts, http.StatusConflict, "has_products", "این دسته کالا دارد؛ ابتدا کالاها را منتقل کنید."},
	{ops.ErrUnknownFetcher, http.StatusBadRequest, "unknown_fetcher", "این نوع منبع پشتیبانی نمی‌شود."},
	{ops.ErrUnsafeURL, http.StatusBadRequest, "unsafe_url", "نشانی منبع مجاز نیست."},
	{ops.ErrInvalidSchedule, http.StatusBadRequest, "invalid_schedule", "بازهٔ زمان‌بندی معتبر نیست."},
	{ops.ErrInvalidConfig, http.StatusBadRequest, "invalid_config", "پیکربندی منبع معتبر نیست."},
	{ops.ErrMatchNotPending, http.StatusConflict, "match_not_pending", "این تطبیق دیگر در صف نیست."},
	{ops.ErrMatchNotDecided, http.StatusConflict, "match_not_decided", "این تطبیق هنوز تایید یا رد نشده است."},
	{ops.ErrSameBrand, http.StatusBadRequest, "same_brand", "نمی‌توان یک برند را با خودش ادغام کرد."},
	{ops.ErrNotSuspicious, http.StatusConflict, "not_suspicious", "این قیمت دیگر در صف بازبینی نیست."},
	{identity.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", "برای ادامه وارد شوید."},
	{identity.ErrInvalidPhone, http.StatusBadRequest, "invalid_phone", "شماره موبایل معتبر نیست."},
	{identity.ErrInvalidOTP, http.StatusUnauthorized, "invalid_otp", "کد تایید نادرست است."},
	{identity.ErrOTPExpired, http.StatusBadRequest, "otp_expired", "کد تایید منقضی شده است. دوباره درخواست کنید."},
	{identity.ErrOTPUsed, http.StatusBadRequest, "otp_used", "این کد قبلاً استفاده شده است."},
	{identity.ErrRateLimited, http.StatusTooManyRequests, "otp_rate_limited", "تعداد ارسال کد بیش از حد مجاز است. کمی بعد دوباره تلاش کنید."},
	{identity.ErrSMSUnavailable, http.StatusServiceUnavailable, "sms_unavailable", "ارسال پیامک الان ممکن نیست. بعداً تلاش کنید."},
	{alerting.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", "برای ادامه وارد شوید."},
	{alerting.ErrInvalidInput, http.StatusBadRequest, "invalid_input", "ورودی معتبر نیست."},
	{alerting.ErrProductNotFound, http.StatusNotFound, "product_not_found", "کالای مورد نظر یافت نشد."},
	{alerting.ErrAlertLimit, http.StatusConflict, "alert_limit", "به سقف تعداد هشدار فعال رسیده‌اید."},
	{alerting.ErrNotFound, http.StatusNotFound, "not_found", "مورد نظر یافت نشد."},
	{alerting.ErrInvalidAlertKind, http.StatusBadRequest, "invalid_alert_kind", "نوع هشدار معتبر نیست."},
	{admin.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", "برای ادامه وارد شوید."},
	{admin.ErrForbidden, http.StatusForbidden, "forbidden", "به این بخش دسترسی ندارید."},
	{admin.ErrInvalidInput, http.StatusBadRequest, "invalid_input", "ورودی معتبر نیست."},
	{admin.ErrNotFound, http.StatusNotFound, "not_found", "مورد نظر یافت نشد."},
	{admin.ErrConflict, http.StatusConflict, "conflict", "این تغییر با دادهٔ موجود سازگار نیست."},
	{admin.ErrLastSuperAdmin, http.StatusConflict, "last_super_admin", "آخرین مدیر کل را نمی‌توان حذف یا تنزل داد."},
	{httpx.ErrBodyTooLarge, http.StatusRequestEntityTooLarge, "body_too_large", "حجم درخواست بیش از حد مجاز است."},
	{httpx.ErrInvalidJSON, http.StatusBadRequest, "invalid_input", "ورودی معتبر نیست."},
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
