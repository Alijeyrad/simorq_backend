package reqctx

import (
	"context"

	"github.com/google/uuid"
)

// AuthClaims defines the interface for authentication claims.
// This interface allows different token implementations (PASETO, JWT, etc.)
// to be used interchangeably.
type AuthClaims interface {
	// GetUserID returns the authenticated user's ID.
	GetUserID() uuid.UUID

	// GetSessionID returns the session ID, if available.
	GetSessionID() *uuid.UUID

	// GetTokenType returns the token type (e.g., "access", "refresh").
	GetTokenType() string

	// IsExpired returns true if the token has expired.
	IsExpired() bool
}

// WithClaims stores authentication claims in the context.
func WithClaims(ctx context.Context, claims AuthClaims) context.Context {
	return context.WithValue(ctx, keyClaims, claims)
}

// ClaimsFromContext retrieves authentication claims from the context.
// Returns nil if not set or if the request is not authenticated.
func ClaimsFromContext(ctx context.Context) AuthClaims {
	v := ctx.Value(keyClaims)
	if v == nil {
		return nil
	}
	claims, ok := v.(AuthClaims)
	if !ok {
		return nil
	}
	return claims
}

// MustClaims retrieves claims from the context.
// Panics if claims are not present.
func MustClaims(ctx context.Context) AuthClaims {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		panic("reqctx: claims not found in context")
	}
	return claims
}

// IsAuthenticated returns true if valid claims exist in the context.
func IsAuthenticated(ctx context.Context) bool {
	claims := ClaimsFromContext(ctx)
	return claims != nil && !claims.IsExpired()
}

// UserIDFromContext extracts the user ID from claims.
// Returns uuid.Nil and false if not authenticated.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return uuid.Nil, false
	}
	return claims.GetUserID(), true
}
