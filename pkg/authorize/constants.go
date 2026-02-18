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

	// Lifecycle actions
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

const (
	WildcardResource Resource = "*"

	// Identity / auth
	ResourceUser         Resource = "user"
	ResourceAuthSession  Resource = "auth_session"
	ResourceRefreshToken Resource = "refresh_token"
	ResourceOTP          Resource = "otp"

	// Clinic (tenant management)
	ResourceClinic           Resource = "clinic"
	ResourceClinicMember     Resource = "clinic_member"
	ResourceClinicSettings   Resource = "clinic_settings"
	ResourceClinicInvitation Resource = "clinic_invitation"

	// Clinical records
	ResourcePatient             Resource = "patient"
	ResourcePatientFile         Resource = "patient_file"
	ResourcePatientReport       Resource = "patient_report"
	ResourcePatientPrescription Resource = "patient_prescription"
	ResourcePatientTest         Resource = "patient_test"
	ResourcePatientIntakeForm   Resource = "patient_intake_form"

	// Scheduling
	ResourceTimeSlot      Resource = "time_slot"
	ResourceRecurringRule Resource = "recurring_rule"
	ResourceAppointment   Resource = "appointment"

	// Financial
	ResourceWallet      Resource = "wallet"
	ResourceTransaction Resource = "transaction"
	ResourcePayment     Resource = "payment"
	ResourceWithdrawal  Resource = "withdrawal"
	ResourceCommission  Resource = "commission"

	// Communication
	ResourceConversation Resource = "conversation"
	ResourceMessage      Resource = "message"
	ResourceTicket       Resource = "ticket"
	ResourceNotification Resource = "notification"

	// Intern module
	ResourceInternTask   Resource = "intern_task"
	ResourceInternAccess Resource = "intern_access"

	// System / platform admin
	ResourceSystem Resource = "system"
	ResourceAudit  Resource = "audit"
	ResourceRBAC   Resource = "rbac"
)

var KnownResources = map[Resource]struct{}{
	ResourceUser: {}, ResourceAuthSession: {}, ResourceRefreshToken: {}, ResourceOTP: {},
	ResourceClinic: {}, ResourceClinicMember: {}, ResourceClinicSettings: {}, ResourceClinicInvitation: {},
	ResourcePatient: {}, ResourcePatientFile: {}, ResourcePatientReport: {},
	ResourcePatientPrescription: {}, ResourcePatientTest: {}, ResourcePatientIntakeForm: {},
	ResourceTimeSlot: {}, ResourceRecurringRule: {}, ResourceAppointment: {},
	ResourceWallet: {}, ResourceTransaction: {}, ResourcePayment: {}, ResourceWithdrawal: {}, ResourceCommission: {},
	ResourceConversation: {}, ResourceMessage: {}, ResourceTicket: {}, ResourceNotification: {},
	ResourceInternTask: {}, ResourceInternAccess: {},
	ResourceSystem: {}, ResourceAudit: {}, ResourceRBAC: {},
}

// ----------------------------
// Roles
// ----------------------------
//
// These are the "policy subjects" we assign to users via grouping policies.

const (
	WildcardRole Role = "*"

	// Platform role (domain = sys)
	RolePlatformSuperAdmin Role = "role:platform:superadmin"

	// Clinic roles (domain = clinic:<uuid>)
	RoleClinicOwner     Role = "role:clinic:owner"
	RoleClinicAdmin     Role = "role:clinic:admin"
	RoleClinicTherapist Role = "role:clinic:therapist"
	RoleClinicIntern    Role = "role:clinic:intern"
	RoleClinicClient    Role = "role:clinic:client" // patient / مراجع

	// Private user scope (domain = user:<uuid>)
	RoleUserSelf Role = "role:user:self"
)

var KnownRoles = map[Role]struct{}{
	RolePlatformSuperAdmin: {},
	RoleClinicOwner:        {},
	RoleClinicAdmin:        {},
	RoleClinicTherapist:    {},
	RoleClinicIntern:       {},
	RoleClinicClient:       {},
	RoleUserSelf:           {},
}

// Persian display names
var RoleDisplayNamesFA = map[Role]string{
	RolePlatformSuperAdmin: "سوپرادمین پلتفرم",
	RoleClinicOwner:        "مالک کلینیک",
	RoleClinicAdmin:        "ادمین کلینیک",
	RoleClinicTherapist:    "درمانگر",
	RoleClinicIntern:       "کارورز",
	RoleClinicClient:       "مراجع",
	RoleUserSelf:           "خود کاربر",
}

// Clinic member role strings (stored in DB clinic_members.role column)
const (
	ClinicMemberRoleOwner     = "owner"
	ClinicMemberRoleAdmin     = "admin"
	ClinicMemberRoleTherapist = "therapist"
	ClinicMemberRoleIntern    = "intern"
)

// ClinicMemberRoleToRBACRole maps DB role values to Casbin roles
var ClinicMemberRoleToRBACRole = map[string]Role{
	ClinicMemberRoleOwner:     RoleClinicOwner,
	ClinicMemberRoleAdmin:     RoleClinicAdmin,
	ClinicMemberRoleTherapist: RoleClinicTherapist,
	ClinicMemberRoleIntern:    RoleClinicIntern,
}

// ----------------------------
// Domains
// ----------------------------

const (
	DomainSys Domain = "sys"
)

// Domain prefixes (for exact domains we generate per entity)
const (
	DomainPrefixClinic Domain = "clinic:"
	DomainPrefixUser   Domain = "user:"
)

const (
	WildcardDomain Domain = "*"
)

var (
	reUUID = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)
)

// Domain builders (typed, safe)
func ClinicDomain(clinicID string) Domain {
	return Domain(fmt.Sprintf("%s%s", DomainPrefixClinic, clinicID))
}

func UserDomain(userID string) Domain {
	return Domain(fmt.Sprintf("%s%s", DomainPrefixUser, userID))
}

// IsValidDomain checks whether d is a recognised domain string.
func IsValidDomain(d Domain) bool {
	if d == DomainSys || d == WildcardDomain {
		return true
	}

	s := string(d)
	switch {
	case len(s) > len(DomainPrefixClinic) && s[:len(DomainPrefixClinic)] == string(DomainPrefixClinic):
		return reUUID.MatchString(s[len(DomainPrefixClinic):])
	case len(s) > len(DomainPrefixUser) && s[:len(DomainPrefixUser)] == string(DomainPrefixUser):
		return reUUID.MatchString(s[len(DomainPrefixUser):])
	default:
		return false
	}
}

// ----------------------------
// Casbin tuple helpers
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
