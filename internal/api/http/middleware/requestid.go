package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/pkg/reqctx"
)

const (
	HeaderRequestID = "X-Request-Id"
	LocalRequestID  = "request_id"
)

// RequestID middleware generates or preserves request IDs and captures request metadata.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		// prefer incoming, else generate
		rid := c.Get(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}

		c.Locals(LocalRequestID, rid)
		c.Set(HeaderRequestID, rid) // send back to client
		// set it on the request headers so adaptor/http handlers can read it
		c.Request().Header.Set(HeaderRequestID, rid)

		// Store full request metadata in locals for later context attachment
		meta := &reqctx.RequestMeta{
			RequestID:   rid,
			ClientIP:    c.IP(),
			UserAgent:   c.Get("User-Agent"),
			RequestedAt: time.Now(),
		}
		c.Locals("request_meta", meta)

		return c.Next()
	}
}

// RequestIDFromFiber retrieves the request ID from Fiber locals.
func RequestIDFromFiber(c fiber.Ctx) (string, bool) {
	v := c.Locals(LocalRequestID)
	s, ok := v.(string)
	return s, ok && s != ""
}

// RequestMetaFromFiber retrieves the full request metadata from Fiber locals.
func RequestMetaFromFiber(c fiber.Ctx) (*reqctx.RequestMeta, bool) {
	v := c.Locals("request_meta")
	meta, ok := v.(*reqctx.RequestMeta)
	return meta, ok && meta != nil
}

// WithRequestMeta attaches request metadata to the context.
// This is the preferred way to store request info in context.
func WithRequestMeta(ctx context.Context, meta *reqctx.RequestMeta) context.Context {
	return reqctx.WithRequestMeta(ctx, meta)
}

// RequestMetaFromContext retrieves request metadata from context.
func RequestMetaFromContext(ctx context.Context) (*reqctx.RequestMeta, bool) {
	return reqctx.RequestMetaFromContext(ctx)
}

// Deprecated: WithRequestID stores only the request ID.
// Use WithRequestMeta for full request metadata.
func WithRequestID(ctx context.Context, rid string) context.Context {
	// For backward compatibility, wrap in RequestMeta
	meta, ok := reqctx.RequestMetaFromContext(ctx)
	if ok && meta != nil {
		// Update existing meta
		meta.RequestID = rid
		return ctx
	}
	// Create minimal meta with just request ID
	return reqctx.WithRequestMeta(ctx, &reqctx.RequestMeta{
		RequestID:   rid,
		RequestedAt: time.Now(),
	})
}

// Deprecated: RequestIDFromContext retrieves the request ID from context.
// Use RequestMetaFromContext and access RequestMeta.RequestID instead.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	rid := reqctx.RequestIDFromContext(ctx)
	return rid, rid != ""
}
