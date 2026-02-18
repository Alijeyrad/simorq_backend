package authorize

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/pkg/reqctx"
)

var (
	ErrNoSubjectInContext = errors.New("no subject found in context")
)

// ClaimsProvider is an interface that any claims type can implement
// to provide user identification for authorization.
// Note: New code should use reqctx.AuthClaims which extends this interface.
type ClaimsProvider interface {
	GetUserID() uuid.UUID
}

// ctxKeyClaimsProvider is the context key for storing claims.
// Deprecated: Claims are now stored via reqctx.WithClaims.
// This is kept for backward compatibility.
type ctxKeyClaimsProvider struct{}

// WithClaimsProvider stores a ClaimsProvider in the context.
// Deprecated: Use reqctx.WithClaims for new code.
// This function is maintained for backward compatibility.
func WithClaimsProvider(ctx context.Context, cp ClaimsProvider) context.Context {
	return context.WithValue(ctx, ctxKeyClaimsProvider{}, cp)
}

// SubjectFromContext extracts the GroupSubject (user ID) from context.
// It first checks reqctx.AuthClaims, then falls back to the legacy ClaimsProvider.
func SubjectFromContext(ctx context.Context) (GroupSubject, error) {
	// First try the new reqctx way
	claims := reqctx.ClaimsFromContext(ctx)
	if claims != nil {
		userID := claims.GetUserID()
		if userID != uuid.Nil {
			return GroupSubject(userID.String()), nil
		}
	}

	// Fall back to legacy ClaimsProvider
	v := ctx.Value(ctxKeyClaimsProvider{})
	if v == nil {
		return "", ErrNoSubjectInContext
	}

	cp, ok := v.(ClaimsProvider)
	if !ok {
		return "", ErrNoSubjectInContext
	}

	userID := cp.GetUserID()
	if userID == uuid.Nil {
		return "", ErrNoSubjectInContext
	}

	return GroupSubject(userID.String()), nil
}

// MustSubjectFromContext extracts the GroupSubject from context or panics.
// Use only when you're certain the subject exists (e.g., after @authenticated directive).
func MustSubjectFromContext(ctx context.Context) GroupSubject {
	subject, err := SubjectFromContext(ctx)
	if err != nil {
		panic(err)
	}
	return subject
}

// UserIDFromContext extracts the user ID as uuid.UUID from context.
// Returns uuid.Nil and error if not found.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// First try the new reqctx way
	claims := reqctx.ClaimsFromContext(ctx)
	if claims != nil {
		userID := claims.GetUserID()
		if userID != uuid.Nil {
			return userID, nil
		}
	}

	// Fall back to legacy ClaimsProvider
	v := ctx.Value(ctxKeyClaimsProvider{})
	if v == nil {
		return uuid.Nil, ErrNoSubjectInContext
	}

	cp, ok := v.(ClaimsProvider)
	if !ok {
		return uuid.Nil, ErrNoSubjectInContext
	}

	userID := cp.GetUserID()
	if userID == uuid.Nil {
		return uuid.Nil, ErrNoSubjectInContext
	}

	return userID, nil
}

// DomainFromResource determines the appropriate domain based on resource ownership.
// - If clinicID is provided, returns clinic:<uuid> domain
// - If userID is provided, returns user:<uuid> domain
// - Otherwise returns sys domain
func DomainFromResource(clinicID, userID *string) Domain {
	if clinicID != nil && *clinicID != "" {
		return ClinicDomain(*clinicID)
	}
	if userID != nil && *userID != "" {
		return UserDomain(*userID)
	}
	return DomainSys
}

// UserSelfDomain returns the user's private domain for self-owned resources.
func UserSelfDomain(userID string) Domain {
	return UserDomain(userID)
}

// DomainFromContext determines the domain based on the current user in context.
// Useful for user-scoped operations where the domain is the user's own domain.
func DomainFromContext(ctx context.Context) (Domain, error) {
	subject, err := SubjectFromContext(ctx)
	if err != nil {
		return "", err
	}
	return UserDomain(string(subject)), nil
}
