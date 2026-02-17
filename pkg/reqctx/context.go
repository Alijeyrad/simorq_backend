package reqctx

import (
	"context"
	"time"
)

// ctxKey is a private type for context keys to prevent collisions.
type ctxKey int

const (
	keyRequestMeta ctxKey = iota
	keyClaims
	keyTrace
	keyGraphQL
)

// RequestMeta holds per-request metadata set by HTTP middleware.
type RequestMeta struct {
	// RequestID is a unique identifier for this request.
	// Format: UUID v4 string.
	RequestID string

	// ClientIP is the client's IP address.
	// May be from X-Forwarded-For or direct connection.
	ClientIP string

	// UserAgent is the client's User-Agent header value.
	UserAgent string

	// RequestedAt is when the request was received.
	RequestedAt time.Time
}

// WithRequestMeta stores RequestMeta in the context.
func WithRequestMeta(ctx context.Context, meta *RequestMeta) context.Context {
	return context.WithValue(ctx, keyRequestMeta, meta)
}

// RequestMetaFromContext retrieves RequestMeta from the context.
// Returns nil, false if not set.
func RequestMetaFromContext(ctx context.Context) (*RequestMeta, bool) {
	v := ctx.Value(keyRequestMeta)
	if v == nil {
		return nil, false
	}
	meta, ok := v.(*RequestMeta)
	return meta, ok
}

// MustRequestMeta retrieves RequestMeta from the context.
// Panics if not set. Use only when middleware guarantees it's present.
func MustRequestMeta(ctx context.Context) *RequestMeta {
	meta, ok := RequestMetaFromContext(ctx)
	if !ok || meta == nil {
		panic("reqctx: RequestMeta not found in context")
	}
	return meta
}

// RequestIDFromContext is a convenience function to get just the request ID.
// Returns empty string if RequestMeta is not set.
func RequestIDFromContext(ctx context.Context) string {
	meta, ok := RequestMetaFromContext(ctx)
	if !ok || meta == nil {
		return ""
	}
	return meta.RequestID
}
