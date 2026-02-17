package authorize

import (
	"testing"
)

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   Domain
		expected bool
	}{
		// Valid domains
		{"sys domain", DomainSys, true},
		{"wildcard domain", WildcardDomain, true},
		{"valid project domain", Domain("project:550e8400-e29b-41d4-a716-446655440000"), true},
		{"valid user domain", Domain("user:550e8400-e29b-41d4-a716-446655440000"), true},

		// Invalid domains
		{"empty domain", Domain(""), false},
		{"random string", Domain("random"), false},
		{"project without uuid", Domain("project:"), false},
		{"project with invalid uuid", Domain("project:invalid-uuid"), false},
		{"user without uuid", Domain("user:"), false},
		{"user with invalid uuid", Domain("user:not-a-uuid"), false},
		{"unknown prefix", Domain("unknown:550e8400-e29b-41d4-a716-446655440000"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("IsValidDomain(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestProjectDomain(t *testing.T) {
	projectID := "550e8400-e29b-41d4-a716-446655440000"
	expected := Domain("project:550e8400-e29b-41d4-a716-446655440000")

	result := ProjectDomain(projectID)
	if result != expected {
		t.Errorf("ProjectDomain(%q) = %q, want %q", projectID, result, expected)
	}
}

func TestUserDomain(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	expected := Domain("user:550e8400-e29b-41d4-a716-446655440000")

	result := UserDomain(userID)
	if result != expected {
		t.Errorf("UserDomain(%q) = %q, want %q", userID, result, expected)
	}
}

func TestKnownActions(t *testing.T) {
	// Verify all expected actions are in the known map
	expectedActions := []Action{
		ActionCreate, ActionRead, ActionUpdate, ActionDelete, ActionList,
		ActionManage, ActionExecute, ActionArchive, ActionClose,
		ActionGrant, ActionRevoke,
	}

	for _, action := range expectedActions {
		if _, ok := KnownActions[action]; !ok {
			t.Errorf("Expected action %q to be in KnownActions", action)
		}
	}
}

func TestKnownResources(t *testing.T) {
	// Verify all expected resources are in the known map
	expectedResources := []Resource{
		ResourceUser, ResourceProfile, ResourceAuthSession, ResourceRefreshToken,
		ResourceOTP, ResourceOAuthIdentity,
		ResourceProject, ResourceChat, ResourceInteraction,
		ResourceFeatureFlag, ResourceUserFeatureFlag,
		ResourceSystem, ResourceAudit, ResourceRBAC,
	}

	for _, resource := range expectedResources {
		if _, ok := KnownResources[resource]; !ok {
			t.Errorf("Expected resource %q to be in KnownResources", resource)
		}
	}
}

func TestKnownRoles(t *testing.T) {
	// Verify all expected roles are in the known map
	expectedRoles := []Role{
		RoleSysSuperAdmin, RoleSysAdmin, RoleSysSupport,
		RoleProjectOwner, RoleProjectAdmin, RoleProjectMember, RoleProjectViewer,
		RoleUserSelf,
	}

	for _, role := range expectedRoles {
		if _, ok := KnownRoles[role]; !ok {
			t.Errorf("Expected role %q to be in KnownRoles", role)
		}
	}
}

func TestRoleDisplayNamesFA(t *testing.T) {
	// Verify all roles have Persian display names
	expectedRoles := []Role{
		RoleSysSuperAdmin, RoleSysAdmin, RoleSysSupport,
		RoleProjectOwner, RoleProjectAdmin, RoleProjectMember, RoleProjectViewer,
		RoleUserSelf,
	}

	for _, role := range expectedRoles {
		if name, ok := RoleDisplayNamesFA[role]; !ok || name == "" {
			t.Errorf("Expected role %q to have a Persian display name", role)
		}
	}
}
