// Package reqctx carries per-request identifiers through context so that any
// layer can log them without threading extra arguments.
package reqctx

import "context"

type contextKey int

const (
	requestIDKey contextKey = iota
	traceIDKey
	actorIDKey
)

// WithRequestID returns a context carrying the given request id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request id, or an empty string when unset.
func RequestID(ctx context.Context) string {
	return stringValue(ctx, requestIDKey)
}

// WithTraceID returns a context carrying the given trace id.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// TraceID returns the trace id, or an empty string when unset.
func TraceID(ctx context.Context) string {
	return stringValue(ctx, traceIDKey)
}

// WithActorID returns a context carrying the authenticated actor id.
func WithActorID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, actorIDKey, id)
}

// ActorID returns the authenticated actor id, or an empty string when unset.
func ActorID(ctx context.Context) string {
	return stringValue(ctx, actorIDKey)
}

func stringValue(ctx context.Context, key contextKey) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(key).(string)
	return v
}
