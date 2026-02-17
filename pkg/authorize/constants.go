package authorize

import (
	"fmt"
	"regexp"
)

type Action string
type Resource string
type Role string
type Domain string

// ----------------------------
// Actions
// ----------------------------

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionList   Action = "list"

	// Power actions
	ActionManage  Action = "manage"  // CRUD + list
	ActionExecute Action = "execute" // run, trigger, start, stop, etc.

	// lifecycle actions
	ActionArchive Action = "archive"
	ActionClose   Action = "close"

	// RBAC-specific actions
	ActionGrant  Action = "grant"
	ActionRevoke Action = "revoke"
)

const (
	WildcardAction Action = "*"
)

var KnownActions = map[Action]struct{}{
	ActionCreate: {}, ActionRead: {}, ActionUpdate: {}, ActionDelete: {}, ActionList: {},
	ActionManage: {}, ActionExecute: {},
	ActionArchive: {}, ActionClose: {},
	ActionGrant: {}, ActionRevoke: {},
}

// ----------------------------
// Resources
// ----------------------------
//

const (
	WildcardResource Resource = "*"

	// Identity / auth
	ResourceUser          Resource = "user"
	ResourceProfile       Resource = "profile"
	ResourceAuthSession   Resource = "auth_session"
	ResourceRefreshToken  Resource = "refresh_token"
	ResourceOTP           Resource = "otp"
	ResourceOAuthIdentity Resource = "oauth_identity"
	ResourceWaitlist      Resource = "waitlist"

	// Core product
	ResourceProject     Resource = "project"
	ResourceChat        Resource = "chat"
	ResourceInteraction Resource = "interaction"

	// Feature management
	ResourceFeatureFlag     Resource = "feature_flag"
	ResourceUserFeatureFlag Resource = "user_feature_flag"

	// System
	ResourceSystem Resource = "system"
	ResourceAudit  Resource = "audit"
	ResourceRBAC   Resource = "rbac"
)

var KnownResources = map[Resource]struct{}{
	ResourceUser: {}, ResourceProfile: {}, ResourceAuthSession: {}, ResourceRefreshToken: {}, ResourceOTP: {}, ResourceOAuthIdentity: {},
	ResourceProject: {}, ResourceChat: {}, ResourceInteraction: {},
	ResourceFeatureFlag: {}, ResourceUserFeatureFlag: {},
	ResourceSystem: {}, ResourceAudit: {}, ResourceRBAC: {}, ResourceWaitlist: {},
}

// ----------------------------
// Roles
// ----------------------------
//
// These are the “policy subjects” we assign to users via grouping policies.

const (
	WildcardRole Role = "*"

	// Platform / system roles (domain = sys)
	RoleSysSuperAdmin Role = "role:sys:superadmin"
	RoleSysAdmin      Role = "role:sys:admin"
	RoleSysSupport    Role = "role:sys:support"

	// Project roles (domain = project:<uuid>)
	RoleProjectOwner  Role = "role:project:owner"
	RoleProjectAdmin  Role = "role:project:admin"
	RoleProjectMember Role = "role:project:member"
	RoleProjectViewer Role = "role:project:viewer"

	// Private user scope (domain = user:<uuid>)
	RoleUserSelf Role = "role:user:self"
)

var KnownRoles = map[Role]struct{}{
	RoleSysSuperAdmin: {}, RoleSysAdmin: {}, RoleSysSupport: {},
	RoleProjectOwner: {}, RoleProjectAdmin: {}, RoleProjectMember: {}, RoleProjectViewer: {},
	RoleUserSelf: {},
}

// Persian display names
var RoleDisplayNamesFA = map[Role]string{
	RoleSysSuperAdmin: "سوپرادمین سیستم",
	RoleSysAdmin:      "ادمین سیستم",
	RoleSysSupport:    "پشتیبانی",

	RoleProjectOwner:  "مالک پروژه",
	RoleProjectAdmin:  "ادمین پروژه",
	RoleProjectMember: "عضو پروژه",
	RoleProjectViewer: "بیننده پروژه",

	RoleUserSelf: "مالک (حریم شخصی)",
}

// ----------------------------
// Domains
// ----------------------------
//

const (
	DomainSys Domain = "sys"
)

// Domain prefixes (for exact domains we generate per entity)
const (
	DomainPrefixProject Domain = "project:"
	DomainPrefixUser    Domain = "user:"
)

const (
	WildcardDomain Domain = "*"
)

var (
	reUUID = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)
)

// Domain builders (typed, safe)
func ProjectDomain(projectID string) Domain {
	return Domain(fmt.Sprintf("%s%s", DomainPrefixProject, projectID))
}

func UserDomain(userID string) Domain {
	return Domain(fmt.Sprintf("%s%s", DomainPrefixUser, userID))
}

// Optional strict validators
func IsValidDomain(d Domain) bool {
	if d == DomainSys || d == WildcardDomain {
		return true
	}

	s := string(d)
	switch {
	case len(s) > len(DomainPrefixProject) && s[:len(DomainPrefixProject)] == string(DomainPrefixProject):
		return reUUID.MatchString(s[len(DomainPrefixProject):])
	case len(s) > len(DomainPrefixUser) && s[:len(DomainPrefixUser)] == string(DomainPrefixUser):
		return reUUID.MatchString(s[len(DomainPrefixUser):])
	default:
		return false
	}
}

// ----------------------------
// Casbin tuple helpers (optional)
// ----------------------------

type PolicyEffect string

const (
	EffectAllow PolicyEffect = "allow"
	EffectDeny  PolicyEffect = "deny"
)

// PolicySubject is the p.sub in Casbin: either a role (preferred) or a user/service id.
type PolicySubject string

// GroupSubject is the g.sub in Casbin: a concrete principal id (user_id or service_id).
type GroupSubject string

// Grouping rows: g, user_id, role, domain
type GroupingPolicy struct {
	Subject GroupSubject
	Role    Role
	Domain  Domain
}

// Permission rows: p, role, domain, resource, action, eft
type PermissionPolicy struct {
	Subject Role
	Domain  Domain
	Object  Resource
	Action  Action
	Effect  PolicyEffect
}
