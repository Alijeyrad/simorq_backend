package authorize

import (
	"context"
	"log/slog"
)

// SeedDefaultPolicies sets up the baseline RBAC policies for the system.
func SeedDefaultPolicies(ctx context.Context, auth IAuthorization) error {
	logger := slog.Default()

	// System-level policies (domain: sys)
	sysPolicies := []PermissionPolicy{
		// SuperAdmin: god mode
		{RoleSysSuperAdmin, DomainSys, WildcardResource, WildcardAction, EffectAllow},

		// SysAdmin: manage most things except RBAC
		{RoleSysAdmin, DomainSys, ResourceUser, ActionManage, EffectAllow},
		{RoleSysAdmin, DomainSys, ResourceProject, ActionManage, EffectAllow},
		{RoleSysAdmin, DomainSys, ResourceFeatureFlag, ActionManage, EffectAllow},
		{RoleSysAdmin, DomainSys, ResourceAudit, ActionRead, EffectAllow},

		// SysSupport: read-only for troubleshooting
		{RoleSysSupport, DomainSys, ResourceUser, ActionRead, EffectAllow},
		{RoleSysSupport, DomainSys, ResourceChat, ActionRead, EffectAllow},
		{RoleSysSupport, DomainSys, ResourceProject, ActionRead, EffectAllow},
	}

	// Project-level policies (domain: project:*)
	projectPolicies := []PermissionPolicy{
		// ProjectOwner: full control within project
		{RoleProjectOwner, WildcardDomain, ResourceProject, ActionManage, EffectAllow},
		{RoleProjectOwner, WildcardDomain, ResourceChat, ActionManage, EffectAllow},
		{RoleProjectOwner, WildcardDomain, ResourceInteraction, ActionManage, EffectAllow},
		{RoleProjectOwner, WildcardDomain, ResourceRBAC, ActionGrant, EffectAllow},

		// ProjectAdmin: manage content but not RBAC
		{RoleProjectAdmin, WildcardDomain, ResourceProject, ActionUpdate, EffectAllow},
		{RoleProjectAdmin, WildcardDomain, ResourceChat, ActionManage, EffectAllow},
		{RoleProjectAdmin, WildcardDomain, ResourceInteraction, ActionManage, EffectAllow},

		// ProjectMember: create and read
		{RoleProjectMember, WildcardDomain, ResourceChat, ActionCreate, EffectAllow},
		{RoleProjectMember, WildcardDomain, ResourceChat, ActionRead, EffectAllow},
		{RoleProjectMember, WildcardDomain, ResourceInteraction, ActionCreate, EffectAllow},
		{RoleProjectMember, WildcardDomain, ResourceInteraction, ActionRead, EffectAllow},

		// ProjectViewer: read-only
		{RoleProjectViewer, WildcardDomain, ResourceChat, ActionRead, EffectAllow},
		{RoleProjectViewer, WildcardDomain, ResourceInteraction, ActionRead, EffectAllow},
	}

	// User-level policies (domain: user:*)
	userPolicies := []PermissionPolicy{
		// UserSelf: full control over own resources
		{RoleUserSelf, WildcardDomain, ResourceProfile, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceAuthSession, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceRefreshToken, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceOAuthIdentity, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceProject, ActionCreate, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceChat, ActionCreate, EffectAllow},
	}

	allPolicies := append(append(sysPolicies, projectPolicies...), userPolicies...)

	for _, p := range allPolicies {
		added, err := auth.AddPermission(ctx, p.Subject, p.Domain, p.Object, p.Action, p.Effect)
		if err != nil {
			logger.Error("failed to add policy", "policy", p, "error", err)
			return err
		}
		if added {
			logger.Debug("added policy", "role", p.Subject, "domain", p.Domain, "resource", p.Object, "action", p.Action)
		}
	}

	logger.Info("seeded default RBAC policies", "count", len(allPolicies))
	return nil
}

// AssignUserSelfRole assigns the user:self role in the user's private domain.
// Call this when creating a new user.
func AssignUserSelfRole(ctx context.Context, auth IAuthorization, userID string) error {
	domain := UserDomain(userID)
	subject := GroupSubject(userID)

	_, err := auth.AddRoleForUserInDomain(ctx, subject, RoleUserSelf, domain)
	return err
}

// AssignProjectOwnerRole assigns the project:owner role to a user for a specific project.
// Call this when creating a new project.
func AssignProjectOwnerRole(ctx context.Context, auth IAuthorization, userID, projectID string) error {
	domain := ProjectDomain(projectID)
	subject := GroupSubject(userID)

	_, err := auth.AddRoleForUserInDomain(ctx, subject, RoleProjectOwner, domain)
	return err
}

// AssignProjectRole assigns a project role to a user for a specific project.
// Use this when adding members to a project with a specific role.
// Valid roles: RoleProjectAdmin, RoleProjectMember, RoleProjectViewer
func AssignProjectRole(ctx context.Context, auth IAuthorization, userID, projectID string, role Role) error {
	// Validate role is a project role
	switch role {
	case RoleProjectOwner, RoleProjectAdmin, RoleProjectMember, RoleProjectViewer:
		// valid project roles
	default:
		return ErrInvalidArgs
	}

	domain := ProjectDomain(projectID)
	subject := GroupSubject(userID)

	_, err := auth.AddRoleForUserInDomain(ctx, subject, role, domain)
	return err
}

// RemoveProjectRole removes a project role from a user for a specific project.
func RemoveProjectRole(ctx context.Context, auth IAuthorization, userID, projectID string, role Role) error {
	domain := ProjectDomain(projectID)
	subject := GroupSubject(userID)

	_, err := auth.RemoveRoleForUserInDomain(ctx, subject, role, domain)
	return err
}

// GetProjectRoles returns all roles a user has in a specific project.
func GetProjectRoles(ctx context.Context, auth IAuthorization, userID, projectID string) ([]Role, error) {
	domain := ProjectDomain(projectID)
	subject := GroupSubject(userID)

	return auth.GetRolesForUserInDomain(ctx, subject, domain)
}

// AssignSystemRole assigns a system-level role to a user.
// Valid roles: RoleSysAdmin, RoleSysSupport
// Note: RoleSysSuperAdmin should be assigned manually/carefully.
func AssignSystemRole(ctx context.Context, auth IAuthorization, userID string, role Role) error {
	switch role {
	case RoleSysAdmin, RoleSysSupport:
		// valid system roles that can be assigned programmatically
	case RoleSysSuperAdmin:
		// superadmin is valid but should be assigned with caution
	default:
		return ErrInvalidArgs
	}

	subject := GroupSubject(userID)
	_, err := auth.AddRoleForUserInDomain(ctx, subject, role, DomainSys)
	return err
}

// RemoveSystemRole removes a system-level role from a user.
func RemoveSystemRole(ctx context.Context, auth IAuthorization, userID string, role Role) error {
	subject := GroupSubject(userID)
	_, err := auth.RemoveRoleForUserInDomain(ctx, subject, role, DomainSys)
	return err
}
