package authorize

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	casbin "github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// createTestEnforcer creates an in-memory Casbin enforcer for testing
func createTestEnforcer(t *testing.T) *casbin.DistributedEnforcer {
	t.Helper()

	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Write model config
	modelPath := filepath.Join(tmpDir, "model.conf")
	modelContent := `[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = (g(r.sub, p.sub, r.dom) || g2(r.sub, p.sub)) && (p.dom == "*" || p.dom == r.dom) && (p.obj == "*" || keyMatch2(r.obj, p.obj)) && (p.act == "*" || keyMatch(r.act, p.act))
`
	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		t.Fatalf("failed to write model file: %v", err)
	}

	// Write empty policy file
	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	// Create adapter with file
	a := fileadapter.NewAdapter(policyPath)

	e, err := casbin.NewDistributedEnforcer(modelPath, a)
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	e.EnableAutoSave(false)
	e.EnableEnforce(true)

	return e
}

func TestNewAuthorization(t *testing.T) {
	t.Run("returns error for nil enforcer", func(t *testing.T) {
		_, err := NewAuthorization(nil)
		if err == nil {
			t.Error("Expected error for nil enforcer")
		}
	})

	t.Run("succeeds with valid enforcer", func(t *testing.T) {
		e := createTestEnforcer(t)
		auth, err := NewAuthorization(e)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if auth == nil {
			t.Error("Expected non-nil authorization")
		}
	})
}

func TestEnforce(t *testing.T) {
	e := createTestEnforcer(t)
	auth, _ := NewAuthorization(e)
	ctx := context.Background()

	// Add test policies
	userID := "user-123"
	projectID := "550e8400-e29b-41d4-a716-446655440000"
	domain := ProjectDomain(projectID)

	// Add role to user
	_, err := auth.AddRoleForUserInDomain(ctx, GroupSubject(userID), RoleProjectOwner, domain)
	if err != nil {
		t.Fatalf("Failed to add role: %v", err)
	}

	// Add permission to role
	_, err = auth.AddPermission(ctx, RoleProjectOwner, domain, ResourceChat, ActionManage, EffectAllow)
	if err != nil {
		t.Fatalf("Failed to add permission: %v", err)
	}

	tests := []struct {
		name     string
		subject  GroupSubject
		domain   Domain
		resource Resource
		action   Action
		want     bool
		wantErr  bool
	}{
		{
			name:     "allowed when permission exists",
			subject:  GroupSubject(userID),
			domain:   domain,
			resource: ResourceChat,
			action:   ActionManage,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "denied when no permission",
			subject:  GroupSubject(userID),
			domain:   domain,
			resource: ResourceUser,
			action:   ActionRead,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "error for empty subject",
			subject:  "",
			domain:   domain,
			resource: ResourceChat,
			action:   ActionRead,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "error for invalid domain",
			subject:  GroupSubject(userID),
			domain:   Domain("invalid"),
			resource: ResourceChat,
			action:   ActionRead,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "error for unknown resource",
			subject:  GroupSubject(userID),
			domain:   domain,
			resource: Resource("unknown"),
			action:   ActionRead,
			want:     false,
			wantErr:  true,
		},
		{
			name:     "error for unknown action",
			subject:  GroupSubject(userID),
			domain:   domain,
			resource: ResourceChat,
			action:   Action("unknown"),
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := auth.Enforce(ctx, tt.subject, tt.domain, tt.resource, tt.action)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("Enforce() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMustEnforce(t *testing.T) {
	e := createTestEnforcer(t)
	auth, _ := NewAuthorization(e)
	ctx := context.Background()

	// Add test policies
	userID := "user-456"
	domain := DomainSys

	// Add role and permission
	auth.AddRoleForUserInDomain(ctx, GroupSubject(userID), RoleSysAdmin, domain)
	auth.AddPermission(ctx, RoleSysAdmin, domain, ResourceUser, ActionManage, EffectAllow)

	t.Run("returns nil when allowed", func(t *testing.T) {
		err := auth.MustEnforce(ctx, GroupSubject(userID), domain, ResourceUser, ActionManage)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("returns ErrForbidden when denied", func(t *testing.T) {
		err := auth.MustEnforce(ctx, GroupSubject(userID), domain, ResourceAudit, ActionDelete)
		if err != ErrForbidden {
			t.Errorf("Expected ErrForbidden, got %v", err)
		}
	})
}

func TestSuperAdminBypass(t *testing.T) {
	e := createTestEnforcer(t)
	auth, _ := NewAuthorization(e)
	ctx := context.Background()

	adminID := "super-admin-id"

	// Add superadmin role
	_, err := auth.AddRoleForUserInDomain(ctx, GroupSubject(adminID), RoleSysSuperAdmin, DomainSys)
	if err != nil {
		t.Fatalf("Failed to add superadmin role: %v", err)
	}

	// Superadmin should be allowed to do anything (bypass check)
	allowed, err := auth.Enforce(ctx, GroupSubject(adminID), DomainSys, ResourceUser, ActionDelete)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !allowed {
		t.Error("Expected superadmin to be allowed")
	}
}

func TestRoleManagement(t *testing.T) {
	e := createTestEnforcer(t)
	auth, _ := NewAuthorization(e)
	ctx := context.Background()

	userID := "user-789"
	projectID := "550e8400-e29b-41d4-a716-446655440000"
	domain := ProjectDomain(projectID)

	t.Run("add and get roles", func(t *testing.T) {
		// Add role
		added, err := auth.AddRoleForUserInDomain(ctx, GroupSubject(userID), RoleProjectMember, domain)
		if err != nil {
			t.Errorf("Failed to add role: %v", err)
		}
		if !added {
			t.Error("Expected role to be added")
		}

		// Get roles
		roles, err := auth.GetRolesForUserInDomain(ctx, GroupSubject(userID), domain)
		if err != nil {
			t.Errorf("Failed to get roles: %v", err)
		}
		if len(roles) != 1 {
			t.Errorf("Expected 1 role, got %d", len(roles))
		}
		if roles[0] != RoleProjectMember {
			t.Errorf("Expected role %q, got %q", RoleProjectMember, roles[0])
		}
	})

	t.Run("remove role", func(t *testing.T) {
		// Remove role
		removed, err := auth.RemoveRoleForUserInDomain(ctx, GroupSubject(userID), RoleProjectMember, domain)
		if err != nil {
			t.Errorf("Failed to remove role: %v", err)
		}
		if !removed {
			t.Error("Expected role to be removed")
		}

		// Verify removal
		roles, _ := auth.GetRolesForUserInDomain(ctx, GroupSubject(userID), domain)
		if len(roles) != 0 {
			t.Errorf("Expected 0 roles after removal, got %d", len(roles))
		}
	})

	t.Run("error for invalid role", func(t *testing.T) {
		_, err := auth.AddRoleForUserInDomain(ctx, GroupSubject(userID), Role("invalid-role"), domain)
		if err == nil {
			t.Error("Expected error for invalid role")
		}
	})
}

func TestPermissionManagement(t *testing.T) {
	e := createTestEnforcer(t)
	auth, _ := NewAuthorization(e)
	ctx := context.Background()

	t.Run("add and remove permission", func(t *testing.T) {
		// Add permission
		added, err := auth.AddPermission(ctx, RoleSysSupport, DomainSys, ResourceChat, ActionRead, EffectAllow)
		if err != nil {
			t.Errorf("Failed to add permission: %v", err)
		}
		if !added {
			t.Error("Expected permission to be added")
		}

		// Remove permission
		removed, err := auth.RemovePermission(ctx, RoleSysSupport, DomainSys, ResourceChat, ActionRead, EffectAllow)
		if err != nil {
			t.Errorf("Failed to remove permission: %v", err)
		}
		if !removed {
			t.Error("Expected permission to be removed")
		}
	})

	t.Run("error for invalid effect", func(t *testing.T) {
		_, err := auth.AddPermission(ctx, RoleSysAdmin, DomainSys, ResourceUser, ActionRead, PolicyEffect("invalid"))
		if err == nil {
			t.Error("Expected error for invalid effect")
		}
	})
}
