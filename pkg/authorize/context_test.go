package authorize

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockClaimsProvider implements ClaimsProvider for testing
type mockClaimsProvider struct {
	userID uuid.UUID
}

func (m *mockClaimsProvider) GetUserID() uuid.UUID {
	return m.userID
}

func TestSubjectFromContext(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		wantSubject GroupSubject
		wantErr     bool
	}{
		{
			name: "valid claims provider",
			setupCtx: func() context.Context {
				cp := &mockClaimsProvider{userID: validUUID}
				return WithClaimsProvider(context.Background(), cp)
			},
			wantSubject: GroupSubject(validUUID.String()),
			wantErr:     false,
		},
		{
			name: "no claims provider in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantSubject: "",
			wantErr:     true,
		},
		{
			name: "nil uuid in claims provider",
			setupCtx: func() context.Context {
				cp := &mockClaimsProvider{userID: uuid.Nil}
				return WithClaimsProvider(context.Background(), cp)
			},
			wantSubject: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			subject, err := SubjectFromContext(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if subject != tt.wantSubject {
					t.Errorf("SubjectFromContext() = %q, want %q", subject, tt.wantSubject)
				}
			}
		})
	}
}

func TestMustSubjectFromContext(t *testing.T) {
	// Test panic case
	t.Run("panics when no claims", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but didn't get one")
			}
		}()
		MustSubjectFromContext(context.Background())
	})

	// Test success case
	t.Run("returns subject when claims exist", func(t *testing.T) {
		validUUID := uuid.New()
		cp := &mockClaimsProvider{userID: validUUID}
		ctx := WithClaimsProvider(context.Background(), cp)

		subject := MustSubjectFromContext(ctx)
		expected := GroupSubject(validUUID.String())
		if subject != expected {
			t.Errorf("MustSubjectFromContext() = %q, want %q", subject, expected)
		}
	})
}

func TestDomainFromResource(t *testing.T) {
	projectID := "project-123"
	userID := "user-456"

	tests := []struct {
		name       string
		projectID  *string
		userID     *string
		wantDomain Domain
	}{
		{
			name:       "project domain when projectID provided",
			projectID:  &projectID,
			userID:     nil,
			wantDomain: Domain("project:project-123"),
		},
		{
			name:       "user domain when userID provided",
			projectID:  nil,
			userID:     &userID,
			wantDomain: Domain("user:user-456"),
		},
		{
			name:       "project takes precedence over user",
			projectID:  &projectID,
			userID:     &userID,
			wantDomain: Domain("project:project-123"),
		},
		{
			name:       "sys domain when neither provided",
			projectID:  nil,
			userID:     nil,
			wantDomain: DomainSys,
		},
		{
			name:       "sys domain when empty strings provided",
			projectID:  strPtr(""),
			userID:     strPtr(""),
			wantDomain: DomainSys,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DomainFromResource(tt.projectID, tt.userID)
			if result != tt.wantDomain {
				t.Errorf("DomainFromResource() = %q, want %q", result, tt.wantDomain)
			}
		})
	}
}

func TestUserSelfDomain(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	expected := Domain("user:550e8400-e29b-41d4-a716-446655440000")

	result := UserSelfDomain(userID)
	if result != expected {
		t.Errorf("UserSelfDomain(%q) = %q, want %q", userID, result, expected)
	}
}

func strPtr(s string) *string {
	return &s
}
