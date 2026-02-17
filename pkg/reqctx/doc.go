// Package reqctx provides centralized request context management.
//
// This package is the single source of truth for all request-scoped data
// including authentication claims, request metadata, tracing information,
// and GraphQL operation context.
//
// # Context Keys
//
// All context keys are private unexported types to prevent collisions.
// Access is provided through type-safe getter and setter functions.
//
// # Usage
//
// Setting values (typically in middleware):
//
//	ctx = reqctx.WithRequestMeta(ctx, &reqctx.RequestMeta{
//	    RequestID:   "abc-123",
//	    ClientIP:    "192.168.1.1",
//	    UserAgent:   "Mozilla/5.0",
//	    RequestedAt: time.Now(),
//	})
//
//	ctx = reqctx.WithClaims(ctx, claims)
//
// Getting values (in resolvers, services, etc.):
//
//	meta, ok := reqctx.RequestMetaFromContext(ctx)
//	claims := reqctx.ClaimsFromContext(ctx)
//	if reqctx.IsAuthenticated(ctx) {
//	    userID, _ := reqctx.UserIDFromContext(ctx)
//	}
//
// # Tracing
//
// The package provides OpenTelemetry-compatible trace context:
//
//	ctx = reqctx.WithTrace(ctx, &reqctx.TraceInfo{
//	    TraceID: reqctx.GenerateTraceID(),
//	    SpanID:  reqctx.GenerateSpanID(),
//	})
//
// # Contracts
//
// The following contracts are guaranteed:
//
//   - RequestMeta is always set by HTTP middleware for all requests
//   - Claims is set only for authenticated requests (token present and valid)
//   - TraceInfo is set when distributed tracing is enabled
//   - GraphQLOperation is set for GraphQL requests only
package reqctx
