// Package reqctx carries per-request identifiers through context so that any
// layer can log them without threading extra arguments.
package reqctx

import (
	"context"
	"strconv"
)

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

// Actor is the authenticated internal operator for this request.
type Actor struct {
	ID       int64
	Username string
	Role     string
}

type actorValueKey struct{}

// WithActor stores the operator on the request context.
func WithActor(ctx context.Context, actor Actor) context.Context {
	ctx = context.WithValue(ctx, actorValueKey{}, actor)
	return WithActorID(ctx, actor.Username)
}

// ActorFrom returns the operator when the request is authenticated.
func ActorFrom(ctx context.Context) (Actor, bool) {
	if ctx == nil {
		return Actor{}, false
	}
	v, ok := ctx.Value(actorValueKey{}).(Actor)
	return v, ok
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

// User is the authenticated public account for this request.
type User struct {
	ID    int64
	Phone string
	Role  string
}

type userValueKey struct{}

// WithUser stores the public account and sets actor_id for logs.
func WithUser(ctx context.Context, user User) context.Context {
	ctx = context.WithValue(ctx, userValueKey{}, user)
	return WithActorID(ctx, "user:"+strconv.FormatInt(user.ID, 10))
}

// UserFrom returns the public account when the request is authenticated.
func UserFrom(ctx context.Context) (User, bool) {
	if ctx == nil {
		return User{}, false
	}
	v, ok := ctx.Value(userValueKey{}).(User)
	return v, ok
}
