// Package logger builds the structured JSON logger used by every service.
//
// Event names follow the domain.action convention (search.query, redirect.click)
// so logs can be aggregated and alerted on. Never log OTP codes, tokens,
// cookies or full phone numbers; use MaskPhone for phone numbers.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

// New returns a JSON logger writing to stdout, tagged with the service name.
func New(service string, level slog.Level) *slog.Logger {
	return NewWithWriter(os.Stdout, service, level)
}

// NewWithWriter is New with an explicit destination, used by tests.
func NewWithWriter(w io.Writer, service string, level slog.Level) *slog.Logger {
	base := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(contextHandler{Handler: base}).With(slog.String("service", service))
}

// ParseLevel maps a configuration string to a slog level.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "", "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q", s)
	}
}

// contextHandler attaches request-scoped identifiers to every record, so any
// *Context log call carries them without the caller repeating the fields.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := reqctx.RequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if id := reqctx.TraceID(ctx); id != "" {
		r.AddAttrs(slog.String("trace_id", id))
	}
	if id := reqctx.ActorID(ctx); id != "" {
		r.AddAttrs(slog.String("actor_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}

// MaskPhone hides the middle digits of a phone number so it can appear in logs
// without exposing the subscriber, e.g. 09121234567 becomes 0912***4567.
func MaskPhone(phone string) string {
	digits := make([]rune, 0, len(phone))
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) < 7 {
		return "***"
	}
	return string(digits[:4]) + "***" + string(digits[len(digits)-4:])
}
