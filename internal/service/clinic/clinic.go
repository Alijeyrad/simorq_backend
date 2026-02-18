package clinic

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entclinic "github.com/Alijeyrad/simorq_backend/internal/repo/clinic"
	entmember "github.com/Alijeyrad/simorq_backend/internal/repo/clinicmember"
	entsettings "github.com/Alijeyrad/simorq_backend/internal/repo/clinicsettings"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type PaginatedResult[T any] struct {
	Data       []T
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

type ListClinicsRequest struct {
	Page    int
	PerPage int
	Active  *bool
}

type CreateClinicRequest struct {
	Name        string
	Slug        string
	Description string
	Phone       string
	Address     string
	City        string
	Province    string
}

type UpdateClinicRequest struct {
	Name        *string
	Description *string
	Phone       *string
	Address     *string
	City        *string
	Province    *string
	LogoKey     *string
}

type UpdateSettingsRequest struct {
	ReservationFeeAmount    *int64
	ReservationFeePercent   *int
	CancellationWindowHours *int
	CancellationFeeAmount   *int64
	CancellationFeePercent  *int
	AllowClientSelfBook     *bool
	DefaultSessionDurationMin *int
	DefaultSessionPrice     *int64
	WorkingHours            map[string]any
}

type AddMemberRequest struct {
	UserID uuid.UUID
	Role   string // owner | admin | therapist | intern
}

type UpdateMemberRequest struct {
	Role     *string
	IsActive *bool
}

// ---------------------------------------------------------------------------
// Service interface
// ---------------------------------------------------------------------------

type Service interface {
	CreateClinic(ctx context.Context, userID uuid.UUID, req CreateClinicRequest) (*repo.Clinic, error)
	GetClinic(ctx context.Context, clinicID uuid.UUID) (*repo.Clinic, error)
	GetClinicBySlug(ctx context.Context, slug string) (*repo.Clinic, error)
	ListClinics(ctx context.Context, req ListClinicsRequest) (*PaginatedResult[*repo.Clinic], error)
	UpdateClinic(ctx context.Context, clinicID uuid.UUID, req UpdateClinicRequest) (*repo.Clinic, error)

	GetSettings(ctx context.Context, clinicID uuid.UUID) (*repo.ClinicSettings, error)
	UpdateSettings(ctx context.Context, clinicID uuid.UUID, req UpdateSettingsRequest) (*repo.ClinicSettings, error)

	ListMembers(ctx context.Context, clinicID uuid.UUID) ([]*repo.ClinicMember, error)
	ListTherapists(ctx context.Context, clinicID uuid.UUID) ([]*repo.ClinicMember, error)
	AddMember(ctx context.Context, clinicID uuid.UUID, req AddMemberRequest) (*repo.ClinicMember, error)
	UpdateMember(ctx context.Context, clinicID, memberID uuid.UUID, req UpdateMemberRequest) (*repo.ClinicMember, error)
	RemoveMember(ctx context.Context, clinicID, memberID uuid.UUID) error

	IsMember(ctx context.Context, clinicID, userID uuid.UUID) (bool, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type clinicService struct {
	db   *repo.Client
	auth authorize.IAuthorization
}

func New(db *repo.Client, auth authorize.IAuthorization) Service {
	return &clinicService{db: db, auth: auth}
}

// ---------------------------------------------------------------------------
// Clinic CRUD
// ---------------------------------------------------------------------------

func (s *clinicService) CreateClinic(ctx context.Context, userID uuid.UUID, req CreateClinicRequest) (*repo.Clinic, error) {
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
	if req.Slug == "" {
		req.Slug = slugify(req.Name)
	}

	// Check slug uniqueness
	exists, err := s.db.Clinic.Query().Where(entclinic.Slug(req.Slug)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check slug: %w", err)
	}
	if exists {
		return nil, ErrSlugAlreadyExists
	}

	// Create clinic + member + settings in a transaction
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	c, err := tx.Clinic.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetNillableDescription(nilIfEmpty(req.Description)).
		SetNillablePhone(nilIfEmpty(req.Phone)).
		SetNillableAddress(nilIfEmpty(req.Address)).
		SetNillableCity(nilIfEmpty(req.City)).
		SetNillableProvince(nilIfEmpty(req.Province)).
		SetIsActive(true).
		SetIsVerified(false).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create clinic: %w", err)
	}

	// Add creator as owner
	_, err = tx.ClinicMember.Create().
		SetClinicID(c.ID).
		SetUserID(userID).
		SetRole("owner").
		SetIsActive(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create owner member: %w", err)
	}

	// Create default settings
	_, err = tx.ClinicSettings.Create().
		SetClinicID(c.ID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create settings: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Assign RBAC role
	if err := authorize.AssignClinicOwnerRole(ctx, s.auth, userID.String(), c.ID.String()); err != nil {
		// Log but don't fail the request — RBAC can be repaired
		fmt.Printf("warn: assign clinic owner role: %v\n", err)
	}

	return c, nil
}

func (s *clinicService) GetClinic(ctx context.Context, clinicID uuid.UUID) (*repo.Clinic, error) {
	c, err := s.db.Clinic.Query().
		Where(entclinic.ID(clinicID), entclinic.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrClinicNotFound
		}
		return nil, fmt.Errorf("get clinic: %w", err)
	}
	return c, nil
}

func (s *clinicService) GetClinicBySlug(ctx context.Context, slug string) (*repo.Clinic, error) {
	c, err := s.db.Clinic.Query().
		Where(entclinic.Slug(slug), entclinic.DeletedAtIsNil(), entclinic.IsActive(true)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrClinicNotFound
		}
		return nil, fmt.Errorf("get clinic by slug: %w", err)
	}
	return c, nil
}

func (s *clinicService) ListClinics(ctx context.Context, req ListClinicsRequest) (*PaginatedResult[*repo.Clinic], error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	offset := (req.Page - 1) * req.PerPage

	q := s.db.Clinic.Query().Where(entclinic.DeletedAtIsNil())
	if req.Active != nil {
		q = q.Where(entclinic.IsActive(*req.Active))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count clinics: %w", err)
	}

	clinics, err := q.Offset(offset).Limit(req.PerPage).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list clinics: %w", err)
	}

	totalPages := (total + req.PerPage - 1) / req.PerPage
	return &PaginatedResult[*repo.Clinic]{
		Data:       clinics,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, clinicID uuid.UUID, req UpdateClinicRequest) (*repo.Clinic, error) {
	c, err := s.GetClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	upd := s.db.Clinic.UpdateOne(c)
	if req.Name != nil {
		upd = upd.SetName(*req.Name)
	}
	if req.Description != nil {
		upd = upd.SetNillableDescription(req.Description)
	}
	if req.Phone != nil {
		upd = upd.SetNillablePhone(req.Phone)
	}
	if req.Address != nil {
		upd = upd.SetNillableAddress(req.Address)
	}
	if req.City != nil {
		upd = upd.SetNillableCity(req.City)
	}
	if req.Province != nil {
		upd = upd.SetNillableProvince(req.Province)
	}
	if req.LogoKey != nil {
		upd = upd.SetNillableLogoKey(req.LogoKey)
	}

	return upd.Save(ctx)
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (s *clinicService) GetSettings(ctx context.Context, clinicID uuid.UUID) (*repo.ClinicSettings, error) {
	st, err := s.db.ClinicSettings.Query().
		Where(entsettings.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrClinicNotFound
		}
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return st, nil
}

func (s *clinicService) UpdateSettings(ctx context.Context, clinicID uuid.UUID, req UpdateSettingsRequest) (*repo.ClinicSettings, error) {
	st, err := s.GetSettings(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	upd := s.db.ClinicSettings.UpdateOne(st)
	if req.ReservationFeeAmount != nil {
		upd = upd.SetReservationFeeAmount(*req.ReservationFeeAmount)
	}
	if req.ReservationFeePercent != nil {
		upd = upd.SetReservationFeePercent(*req.ReservationFeePercent)
	}
	if req.CancellationWindowHours != nil {
		upd = upd.SetCancellationWindowHours(*req.CancellationWindowHours)
	}
	if req.CancellationFeeAmount != nil {
		upd = upd.SetCancellationFeeAmount(*req.CancellationFeeAmount)
	}
	if req.CancellationFeePercent != nil {
		upd = upd.SetCancellationFeePercent(*req.CancellationFeePercent)
	}
	if req.AllowClientSelfBook != nil {
		upd = upd.SetAllowClientSelfBook(*req.AllowClientSelfBook)
	}
	if req.DefaultSessionDurationMin != nil {
		upd = upd.SetDefaultSessionDurationMin(*req.DefaultSessionDurationMin)
	}
	if req.DefaultSessionPrice != nil {
		upd = upd.SetDefaultSessionPrice(*req.DefaultSessionPrice)
	}
	if req.WorkingHours != nil {
		upd = upd.SetWorkingHours(req.WorkingHours)
	}

	return upd.Save(ctx)
}

// ---------------------------------------------------------------------------
// Members
// ---------------------------------------------------------------------------

func (s *clinicService) ListMembers(ctx context.Context, clinicID uuid.UUID) ([]*repo.ClinicMember, error) {
	return s.db.ClinicMember.Query().
		Where(entmember.ClinicID(clinicID), entmember.IsActive(true)).
		WithUser().
		All(ctx)
}

func (s *clinicService) ListTherapists(ctx context.Context, clinicID uuid.UUID) ([]*repo.ClinicMember, error) {
	return s.db.ClinicMember.Query().
		Where(
			entmember.ClinicID(clinicID),
			entmember.IsActive(true),
			entmember.RoleEQ(entmember.RoleTherapist),
		).
		WithUser().
		All(ctx)
}

func (s *clinicService) AddMember(ctx context.Context, clinicID uuid.UUID, req AddMemberRequest) (*repo.ClinicMember, error) {
	if !isValidRole(req.Role) {
		return nil, ErrInvalidRole
	}

	// Check not already a member
	exists, err := s.db.ClinicMember.Query().
		Where(entmember.ClinicID(clinicID), entmember.UserID(req.UserID)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check member: %w", err)
	}
	if exists {
		return nil, ErrAlreadyMember
	}

	m, err := s.db.ClinicMember.Create().
		SetClinicID(clinicID).
		SetUserID(req.UserID).
		SetRole(entmember.Role(req.Role)).
		SetIsActive(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}

	// Assign RBAC role
	casbinRole := authorize.ClinicMemberRoleToRBACRole[req.Role]
	if casbinRole != "" {
		authorize.AssignClinicRole(ctx, s.auth, req.UserID.String(), clinicID.String(), casbinRole)
	}

	return m, nil
}

func (s *clinicService) UpdateMember(ctx context.Context, clinicID, memberID uuid.UUID, req UpdateMemberRequest) (*repo.ClinicMember, error) {
	m, err := s.db.ClinicMember.Query().
		Where(entmember.ID(memberID), entmember.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("get member: %w", err)
	}

	if req.Role != nil && !isValidRole(*req.Role) {
		return nil, ErrInvalidRole
	}

	upd := s.db.ClinicMember.UpdateOne(m)
	if req.Role != nil {
		// Remove old RBAC role, assign new one
		oldCasbinRole := authorize.ClinicMemberRoleToRBACRole[string(m.Role)]
		if oldCasbinRole != "" {
			authorize.RemoveClinicRole(ctx, s.auth, m.UserID.String(), clinicID.String(), oldCasbinRole)
		}
		newCasbinRole := authorize.ClinicMemberRoleToRBACRole[*req.Role]
		if newCasbinRole != "" {
			authorize.AssignClinicRole(ctx, s.auth, m.UserID.String(), clinicID.String(), newCasbinRole)
		}
		upd = upd.SetRole(entmember.Role(*req.Role))
	}
	if req.IsActive != nil {
		upd = upd.SetIsActive(*req.IsActive)
	}

	return upd.Save(ctx)
}

func (s *clinicService) RemoveMember(ctx context.Context, clinicID, memberID uuid.UUID) error {
	m, err := s.db.ClinicMember.Query().
		Where(entmember.ID(memberID), entmember.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("get member: %w", err)
	}
	if m.Role == "owner" {
		return ErrCannotRemoveOwner
	}

	// Remove RBAC role
	casbinRole := authorize.ClinicMemberRoleToRBACRole[string(m.Role)]
	if casbinRole != "" {
		authorize.RemoveClinicRole(ctx, s.auth, m.UserID.String(), clinicID.String(), casbinRole)
	}

	return s.db.ClinicMember.DeleteOne(m).Exec(ctx)
}

func (s *clinicService) IsMember(ctx context.Context, clinicID, userID uuid.UUID) (bool, error) {
	return s.db.ClinicMember.Query().
		Where(entmember.ClinicID(clinicID), entmember.UserID(userID), entmember.IsActive(true)).
		Exist(ctx)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isValidRole(role string) bool {
	switch role {
	case "owner", "admin", "therapist", "intern":
		return true
	}
	return false
}

func nilIfEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
