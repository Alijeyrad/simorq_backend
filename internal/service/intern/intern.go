package intern

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entipa "github.com/Alijeyrad/simorq_backend/internal/repo/internpatientaccess"
	entprofile "github.com/Alijeyrad/simorq_backend/internal/repo/internprofile"
	enttask "github.com/Alijeyrad/simorq_backend/internal/repo/interntask"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type UpsertProfileRequest struct {
	InternshipYear *int
	SupervisorIDs  []uuid.UUID
}

type CreateTaskRequest struct {
	Title   string
	Caption *string
}

type UpdateTaskRequest struct {
	Title   *string
	Caption *string
}

type AddFileRequest struct {
	FileKey  string
	FileName string
	FileSize *int64
	MimeType *string
}

type ReviewTaskRequest struct {
	Status  string // pending | reviewed | needs_revision
	Comment *string
	Grade   *string
}

type GrantAccessRequest struct {
	CanViewFiles    bool
	CanWriteReports bool
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	// Intern-facing
	GetMyProfile(ctx context.Context, clinicMemberID uuid.UUID) (*repo.InternProfile, error)
	UpsertMyProfile(ctx context.Context, clinicMemberID uuid.UUID, req UpsertProfileRequest) (*repo.InternProfile, error)
	ListMyTasks(ctx context.Context, internID uuid.UUID, page, perPage int) ([]*repo.InternTask, error)
	CreateTask(ctx context.Context, clinicID, internID uuid.UUID, req CreateTaskRequest) (*repo.InternTask, error)
	GetTask(ctx context.Context, taskID, internID uuid.UUID) (*repo.InternTask, error)
	UpdateTask(ctx context.Context, taskID, internID uuid.UUID, req UpdateTaskRequest) (*repo.InternTask, error)
	AddTaskFile(ctx context.Context, taskID uuid.UUID, req AddFileRequest) (*repo.InternTaskFile, error)
	ListMyPatients(ctx context.Context, internID uuid.UUID, page, perPage int) ([]*repo.InternPatientAccess, error)

	// Admin / supervisor-facing
	ListInterns(ctx context.Context, clinicID uuid.UUID, page, perPage int) ([]*repo.InternProfile, error)
	ListInternTasks(ctx context.Context, clinicID, internID uuid.UUID, page, perPage int) ([]*repo.InternTask, error)
	ReviewTask(ctx context.Context, taskID, reviewerID uuid.UUID, req ReviewTaskRequest) (*repo.InternTask, error)
	GrantAccess(ctx context.Context, internID, patientID, grantedBy uuid.UUID, req GrantAccessRequest) (*repo.InternPatientAccess, error)
	RevokeAccess(ctx context.Context, internID, patientID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type internService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &internService{db: db}
}

func (s *internService) GetMyProfile(ctx context.Context, clinicMemberID uuid.UUID) (*repo.InternProfile, error) {
	p, err := s.db.InternProfile.Query().
		Where(entprofile.ClinicMemberID(clinicMemberID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *internService) UpsertMyProfile(ctx context.Context, clinicMemberID uuid.UUID, req UpsertProfileRequest) (*repo.InternProfile, error) {
	existing, err := s.db.InternProfile.Query().
		Where(entprofile.ClinicMemberID(clinicMemberID)).
		Only(ctx)

	if err != nil && !repo.IsNotFound(err) {
		return nil, err
	}

	if repo.IsNotFound(err) {
		// Create
		q := s.db.InternProfile.Create().
			SetClinicMemberID(clinicMemberID).
			SetNillableInternshipYear(req.InternshipYear)
		if req.SupervisorIDs != nil {
			q = q.SetSupervisorIds(req.SupervisorIDs)
		}
		return q.Save(ctx)
	}

	// Update
	q := s.db.InternProfile.UpdateOne(existing).
		SetNillableInternshipYear(req.InternshipYear)
	if req.SupervisorIDs != nil {
		q = q.SetSupervisorIds(req.SupervisorIDs)
	}
	return q.Save(ctx)
}

func (s *internService) ListMyTasks(ctx context.Context, internID uuid.UUID, page, perPage int) ([]*repo.InternTask, error) {
	offset := (page - 1) * perPage
	return s.db.InternTask.Query().
		Where(enttask.InternID(internID)).
		Order(enttask.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
}

func (s *internService) CreateTask(ctx context.Context, clinicID, internID uuid.UUID, req CreateTaskRequest) (*repo.InternTask, error) {
	return s.db.InternTask.Create().
		SetClinicID(clinicID).
		SetInternID(internID).
		SetTitle(req.Title).
		SetNillableCaption(req.Caption).
		Save(ctx)
}

func (s *internService) GetTask(ctx context.Context, taskID, internID uuid.UUID) (*repo.InternTask, error) {
	t, err := s.db.InternTask.Get(ctx, taskID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if t.InternID != internID {
		return nil, ErrUnauthorized
	}
	return t, nil
}

func (s *internService) UpdateTask(ctx context.Context, taskID, internID uuid.UUID, req UpdateTaskRequest) (*repo.InternTask, error) {
	t, err := s.db.InternTask.Get(ctx, taskID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if t.InternID != internID {
		return nil, ErrUnauthorized
	}
	u := s.db.InternTask.UpdateOne(t)
	if req.Title != nil {
		u = u.SetTitle(*req.Title)
	}
	if req.Caption != nil {
		u = u.SetCaption(*req.Caption)
	}
	return u.Save(ctx)
}

func (s *internService) AddTaskFile(ctx context.Context, taskID uuid.UUID, req AddFileRequest) (*repo.InternTaskFile, error) {
	return s.db.InternTaskFile.Create().
		SetTaskID(taskID).
		SetFileKey(req.FileKey).
		SetFileName(req.FileName).
		SetNillableFileSize(req.FileSize).
		SetNillableMimeType(req.MimeType).
		Save(ctx)
}

func (s *internService) ListMyPatients(ctx context.Context, internID uuid.UUID, page, perPage int) ([]*repo.InternPatientAccess, error) {
	offset := (page - 1) * perPage
	return s.db.InternPatientAccess.Query().
		Where(entipa.InternID(internID)).
		Order(entipa.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
}

func (s *internService) ListInterns(ctx context.Context, clinicID uuid.UUID, page, perPage int) ([]*repo.InternProfile, error) {
	// InternProfile doesn't have clinic_id; list by supervisor presence or return all
	// For now, return all profiles ordered by created_at
	offset := (page - 1) * perPage
	return s.db.InternProfile.Query().
		Order(entprofile.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
}

func (s *internService) ListInternTasks(ctx context.Context, clinicID, internID uuid.UUID, page, perPage int) ([]*repo.InternTask, error) {
	offset := (page - 1) * perPage
	return s.db.InternTask.Query().
		Where(
			enttask.ClinicID(clinicID),
			enttask.InternID(internID),
		).
		Order(enttask.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
}

func (s *internService) ReviewTask(ctx context.Context, taskID, reviewerID uuid.UUID, req ReviewTaskRequest) (*repo.InternTask, error) {
	t, err := s.db.InternTask.Get(ctx, taskID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	now := time.Now()
	u := s.db.InternTask.UpdateOne(t).
		SetReviewStatus(enttask.ReviewStatus(req.Status)).
		SetReviewedBy(reviewerID).
		SetReviewedAt(now).
		SetNillableReviewComment(req.Comment).
		SetNillableGrade(req.Grade)

	return u.Save(ctx)
}

func (s *internService) GrantAccess(ctx context.Context, internID, patientID, grantedBy uuid.UUID, req GrantAccessRequest) (*repo.InternPatientAccess, error) {
	exists, err := s.db.InternPatientAccess.Query().
		Where(
			entipa.InternID(internID),
			entipa.PatientID(patientID),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccessAlreadyGranted
	}

	return s.db.InternPatientAccess.Create().
		SetInternID(internID).
		SetPatientID(patientID).
		SetGrantedBy(grantedBy).
		SetCanViewFiles(req.CanViewFiles).
		SetCanWriteReports(req.CanWriteReports).
		Save(ctx)
}

func (s *internService) RevokeAccess(ctx context.Context, internID, patientID uuid.UUID) error {
	n, err := s.db.InternPatientAccess.Delete().
		Where(
			entipa.InternID(internID),
			entipa.PatientID(patientID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
