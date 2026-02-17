// pkg/authorize/casbin.go
package authorize

import (
	"context"
	"errors"
	"fmt"

	casbin "github.com/casbin/casbin/v2"
)

var (
	ErrForbidden   = errors.New("forbidden")
	ErrInvalidArgs = errors.New("invalid authorization arguments")
)

// IAuthorization is the only thing services/middleware should depend on.
type IAuthorization interface {
	// Enforce answers: "Is subject allowed to act on object inside domain?"
	Enforce(ctx context.Context, subject GroupSubject, domain Domain, object Resource, action Action) (bool, error)

	// MustEnforce is convenience for services: return ErrForbidden if not allowed.
	MustEnforce(ctx context.Context, subject GroupSubject, domain Domain, object Resource, action Action) error

	// Role management (grouping policies): g, user_id, role, domain
	AddRoleForUserInDomain(ctx context.Context, subject GroupSubject, role Role, domain Domain) (bool, error)
	RemoveRoleForUserInDomain(ctx context.Context, subject GroupSubject, role Role, domain Domain) (bool, error)
	GetRolesForUserInDomain(ctx context.Context, subject GroupSubject, domain Domain) ([]Role, error)

	// Permission management (policies): p, role, domain, object, action, eft
	AddPermission(ctx context.Context, role Role, domain Domain, object Resource, action Action, effect PolicyEffect) (bool, error)
	RemovePermission(ctx context.Context, role Role, domain Domain, object Resource, action Action, effect PolicyEffect) (bool, error)

	Raw() *casbin.DistributedEnforcer
}

// Authorization is a thin typed wrapper around casbin.Enforcer.
type Authorization struct {
	enforcer       *casbin.DistributedEnforcer
	superAdminRole Role
}

// NewAuthorization wraps an already-configured Enforcer
func NewAuthorization(e *casbin.DistributedEnforcer) (IAuthorization, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: enforcer is nil", ErrInvalidArgs)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}

	return &Authorization{
		enforcer:       e,
		superAdminRole: RoleSysSuperAdmin,
	}, nil
}

func (a *Authorization) Raw() *casbin.DistributedEnforcer { return a.enforcer }

func (a *Authorization) Enforce(ctx context.Context, subject GroupSubject, domain Domain, object Resource, action Action) (bool, error) {
	_ = ctx // reserved for tracing/logging later

	if subject == "" {
		return false, fmt.Errorf("%w: subject is empty", ErrInvalidArgs)
	}
	if domain == "" || !IsValidDomain(domain) {
		return false, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	if object == "" {
		return false, fmt.Errorf("%w: object is empty", ErrInvalidArgs)
	}
	if action == "" {
		return false, fmt.Errorf("%w: action is empty", ErrInvalidArgs)
	}

	// Guardrails: ensure you're only using known constants
	if _, ok := KnownResources[object]; !ok && object != WildcardResource {
		return false, fmt.Errorf("%w: unknown resource: %q", ErrInvalidArgs, object)
	}
	if _, ok := KnownActions[action]; !ok && action != WildcardAction {
		return false, fmt.Errorf("%w: unknown action: %q", ErrInvalidArgs, action)
	}

	// Optional bypass: If user has sys superadmin in sys domain, allow everything.
	if a.superAdminRole != "" {
		if ok := a.enforcer.HasGroupingPolicy(string(subject), string(a.superAdminRole), string(DomainSys)); ok {
			return true, nil
		}
	}

	allowed, err := a.enforcer.Enforce(string(subject), string(domain), string(object), string(action))
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func (a *Authorization) MustEnforce(ctx context.Context, subject GroupSubject, domain Domain, object Resource, action Action) error {
	ok, err := a.Enforce(ctx, subject, domain, object, action)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

// ---- Grouping (roles) ----

func (a *Authorization) AddRoleForUserInDomain(ctx context.Context, subject GroupSubject, role Role, domain Domain) (bool, error) {
	_ = ctx
	if subject == "" || role == "" {
		return false, fmt.Errorf("%w: empty subject/role", ErrInvalidArgs)
	}
	if _, ok := KnownRoles[role]; !ok && role != WildcardRole {
		return false, fmt.Errorf("%w: unknown role: %q", ErrInvalidArgs, role)
	}
	if domain == "" || !IsValidDomain(domain) {
		return false, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	return a.enforcer.AddGroupingPolicy(string(subject), string(role), string(domain))
}

func (a *Authorization) RemoveRoleForUserInDomain(ctx context.Context, subject GroupSubject, role Role, domain Domain) (bool, error) {
	_ = ctx
	if subject == "" || role == "" {
		return false, fmt.Errorf("%w: empty subject/role", ErrInvalidArgs)
	}
	if domain == "" || !IsValidDomain(domain) {
		return false, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	return a.enforcer.RemoveGroupingPolicy(string(subject), string(role), string(domain))
}

func (a *Authorization) GetRolesForUserInDomain(ctx context.Context, subject GroupSubject, domain Domain) ([]Role, error) {
	_ = ctx
	if subject == "" {
		return nil, fmt.Errorf("%w: subject is empty", ErrInvalidArgs)
	}
	if domain == "" || !IsValidDomain(domain) {
		return nil, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	roles := a.enforcer.GetRolesForUserInDomain(string(subject), string(domain))
	out := make([]Role, 0, len(roles))
	for _, r := range roles {
		out = append(out, Role(r))
	}
	return out, nil
}

// ---- Permissions (p rules) ----

func (a *Authorization) AddPermission(ctx context.Context, role Role, domain Domain, object Resource, action Action, effect PolicyEffect) (bool, error) {
	_ = ctx
	if role == "" || domain == "" || object == "" || action == "" || effect == "" {
		return false, fmt.Errorf("%w: empty permission fields", ErrInvalidArgs)
	}
	if _, ok := KnownRoles[role]; !ok && role != WildcardRole {
		return false, fmt.Errorf("%w: unknown role: %q", ErrInvalidArgs, role)
	}
	if !IsValidDomain(domain) {
		return false, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	if _, ok := KnownResources[object]; !ok && object != WildcardResource {
		return false, fmt.Errorf("%w: unknown resource: %q", ErrInvalidArgs, object)
	}
	if _, ok := KnownActions[action]; !ok && action != WildcardAction {
		return false, fmt.Errorf("%w: unknown action: %q", ErrInvalidArgs, action)
	}
	if effect != EffectAllow && effect != EffectDeny {
		return false, fmt.Errorf("%w: invalid effect: %q", ErrInvalidArgs, effect)
	}

	// p, sub(role), dom, obj, act, eft
	return a.enforcer.AddPolicy(string(role), string(domain), string(object), string(action), string(effect))
}

func (a *Authorization) RemovePermission(ctx context.Context, role Role, domain Domain, object Resource, action Action, effect PolicyEffect) (bool, error) {
	_ = ctx
	if role == "" || domain == "" || object == "" || action == "" || effect == "" {
		return false, fmt.Errorf("%w: empty permission fields", ErrInvalidArgs)
	}
	if !IsValidDomain(domain) {
		return false, fmt.Errorf("%w: invalid domain: %q", ErrInvalidArgs, domain)
	}
	return a.enforcer.RemovePolicy(string(role), string(domain), string(object), string(action), string(effect))
}
