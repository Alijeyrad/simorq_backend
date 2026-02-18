package authorize

import (
	"context"
	"log/slog"
)

// SeedDefaultPolicies sets up the baseline RBAC policies for Simorgh clinics.
func SeedDefaultPolicies(ctx context.Context, auth IAuthorization) error {
	logger := slog.Default()

	// ---------------------------------------------------------------------------
	// System / platform level (domain: sys)
	// ---------------------------------------------------------------------------
	sysPolicies := []PermissionPolicy{
		// Platform superadmin: god mode over everything
		{RolePlatformSuperAdmin, DomainSys, WildcardResource, WildcardAction, EffectAllow},
	}

	// ---------------------------------------------------------------------------
	// Clinic level (domain: clinic:*)
	// ---------------------------------------------------------------------------
	clinicPolicies := []PermissionPolicy{
		// Owner: full control inside the clinic
		{RoleClinicOwner, WildcardDomain, WildcardResource, ActionManage, EffectAllow},

		// Admin: manage most resources (not commission rules or platform RBAC)
		{RoleClinicAdmin, WildcardDomain, ResourceClinic, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceClinicMember, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceClinicSettings, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceClinicInvitation, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatient, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatientFile, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatientReport, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatientPrescription, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatientTest, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourcePatientIntakeForm, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceTimeSlot, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceRecurringRule, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceAppointment, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceWallet, ActionRead, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceTransaction, ActionRead, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceConversation, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceMessage, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceTicket, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceNotification, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceInternTask, ActionManage, EffectAllow},
		{RoleClinicAdmin, WildcardDomain, ResourceInternAccess, ActionManage, EffectAllow},

		// Therapist: manage own patients, reports, schedule, appointments
		{RoleClinicTherapist, WildcardDomain, ResourceClinic, ActionRead, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceClinicMember, ActionRead, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceClinicSettings, ActionRead, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatient, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatientFile, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatientReport, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatientPrescription, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatientTest, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourcePatientIntakeForm, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceTimeSlot, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceRecurringRule, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceAppointment, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceWallet, ActionRead, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceTransaction, ActionRead, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceConversation, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceMessage, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceTicket, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceInternTask, ActionManage, EffectAllow},
		{RoleClinicTherapist, WildcardDomain, ResourceInternAccess, ActionManage, EffectAllow},

		// Intern: read-only on patients/files/reports they are explicitly granted access to
		{RoleClinicIntern, WildcardDomain, ResourceClinic, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceClinicMember, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourcePatient, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourcePatientFile, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourcePatientReport, ActionCreate, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourcePatientReport, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceAppointment, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceTimeSlot, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceInternTask, ActionCreate, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceInternTask, ActionRead, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceInternTask, ActionUpdate, EffectAllow},
		{RoleClinicIntern, WildcardDomain, ResourceTicket, ActionManage, EffectAllow},

		// Client (patient): access to own data
		{RoleClinicClient, WildcardDomain, ResourceAppointment, ActionCreate, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceAppointment, ActionRead, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceAppointment, ActionDelete, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourcePatientReport, ActionRead, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourcePatientFile, ActionRead, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceConversation, ActionManage, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceMessage, ActionManage, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceTicket, ActionManage, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceWallet, ActionRead, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceTransaction, ActionRead, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceNotification, ActionManage, EffectAllow},
		{RoleClinicClient, WildcardDomain, ResourceTimeSlot, ActionRead, EffectAllow},
	}

	// ---------------------------------------------------------------------------
	// User scope (domain: user:*)
	// ---------------------------------------------------------------------------
	userPolicies := []PermissionPolicy{
		{RoleUserSelf, WildcardDomain, ResourceUser, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceAuthSession, ActionManage, EffectAllow},
		{RoleUserSelf, WildcardDomain, ResourceRefreshToken, ActionManage, EffectAllow},
	}

	allPolicies := append(append(sysPolicies, clinicPolicies...), userPolicies...)

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

// ---------------------------------------------------------------------------
// Assignment helpers
// ---------------------------------------------------------------------------

// AssignUserSelfRole assigns the user:self role in the user's private domain.
// Call this when creating a new user.
func AssignUserSelfRole(ctx context.Context, auth IAuthorization, userID string) error {
	domain := UserDomain(userID)
	subject := GroupSubject(userID)
	_, err := auth.AddRoleForUserInDomain(ctx, subject, RoleUserSelf, domain)
	return err
}

// AssignClinicOwnerRole assigns the clinic:owner role to a user for a specific clinic.
// Call this when a user creates a new clinic.
func AssignClinicOwnerRole(ctx context.Context, auth IAuthorization, userID, clinicID string) error {
	domain := ClinicDomain(clinicID)
	subject := GroupSubject(userID)
	_, err := auth.AddRoleForUserInDomain(ctx, subject, RoleClinicOwner, domain)
	return err
}

// AssignClinicRole assigns a clinic role to a user for a specific clinic.
// Valid roles: RoleClinicAdmin, RoleClinicTherapist, RoleClinicIntern, RoleClinicClient.
func AssignClinicRole(ctx context.Context, auth IAuthorization, userID, clinicID string, role Role) error {
	switch role {
	case RoleClinicOwner, RoleClinicAdmin, RoleClinicTherapist, RoleClinicIntern, RoleClinicClient:
		// valid clinic roles
	default:
		return ErrInvalidArgs
	}
	domain := ClinicDomain(clinicID)
	subject := GroupSubject(userID)
	_, err := auth.AddRoleForUserInDomain(ctx, subject, role, domain)
	return err
}

// RemoveClinicRole removes a clinic role from a user for a specific clinic.
func RemoveClinicRole(ctx context.Context, auth IAuthorization, userID, clinicID string, role Role) error {
	domain := ClinicDomain(clinicID)
	subject := GroupSubject(userID)
	_, err := auth.RemoveRoleForUserInDomain(ctx, subject, role, domain)
	return err
}

// GetClinicRoles returns all roles a user has in a specific clinic.
func GetClinicRoles(ctx context.Context, auth IAuthorization, userID, clinicID string) ([]Role, error) {
	domain := ClinicDomain(clinicID)
	subject := GroupSubject(userID)
	return auth.GetRolesForUserInDomain(ctx, subject, domain)
}

// AssignSystemRole assigns the platform superadmin role.
// This should only be called manually during initial platform setup.
func AssignSystemRole(ctx context.Context, auth IAuthorization, userID string, role Role) error {
	if role != RolePlatformSuperAdmin {
		return ErrInvalidArgs
	}
	subject := GroupSubject(userID)
	_, err := auth.AddRoleForUserInDomain(ctx, subject, role, DomainSys)
	return err
}

// RemoveSystemRole removes the platform superadmin role from a user.
func RemoveSystemRole(ctx context.Context, auth IAuthorization, userID string, role Role) error {
	subject := GroupSubject(userID)
	_, err := auth.RemoveRoleForUserInDomain(ctx, subject, role, DomainSys)
	return err
}
