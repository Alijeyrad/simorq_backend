package patient

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entpatient "github.com/Alijeyrad/simorq_backend/internal/repo/patient"
	entprescription "github.com/Alijeyrad/simorq_backend/internal/repo/patientprescription"
	entreport "github.com/Alijeyrad/simorq_backend/internal/repo/patientreport"
	enttest "github.com/Alijeyrad/simorq_backend/internal/repo/patienttest"
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

type ListPatientsRequest struct {
	Page          int
	PerPage       int
	TherapistID   *uuid.UUID
	Status        *string
	PaymentStatus *string
	HasDiscount   *bool
	Sort          string // created_at | appointment_date
	Order         string // asc | desc
}

type CreatePatientRequest struct {
	UserID            uuid.UUID
	PrimaryTherapistID *uuid.UUID
	FileNumber        *string
	Notes             *string
	ReferralSource    *string
	ChiefComplaint    *string
	IsChild           bool
	ChildBirthDate    *time.Time
	ChildSchool       *string
	ChildGrade        *string
	ParentName        *string
	ParentPhone       *string
	ParentRelation    *string
}

type UpdatePatientRequest struct {
	PrimaryTherapistID *uuid.UUID
	FileNumber         *string
	Status             *string
	HasDiscount        *bool
	DiscountPercent    *int
	PaymentStatus      *string
	Notes              *string
	ReferralSource     *string
	ChiefComplaint     *string
	IsChild            *bool
	ChildBirthDate     *time.Time
	ChildSchool        *string
	ChildGrade         *string
	ParentName         *string
	ParentPhone        *string
	ParentRelation     *string
}

type CreateReportRequest struct {
	AppointmentID *uuid.UUID
	Title         *string
	Content       *string
	ReportDate    *time.Time
}

type UpdateReportRequest struct {
	Title      *string
	Content    *string
	ReportDate *time.Time
}

type CreatePrescriptionRequest struct {
	Title          *string
	Notes          *string
	FileKey        *string
	FileName       *string
	PrescribedDate *time.Time
}

type UpdatePrescriptionRequest struct {
	Title          *string
	Notes          *string
	FileKey        *string
	FileName       *string
	PrescribedDate *time.Time
}

type CreateTestRequest struct {
	TestID          *uuid.UUID
	AdministeredBy  *uuid.UUID
	TestName        *string
	RawScores       map[string]any
	ComputedScores  map[string]any
	Interpretation  *string
	TestDate        *time.Time
}

type UpdateTestRequest struct {
	AdministeredBy *uuid.UUID
	RawScores      map[string]any
	ComputedScores map[string]any
	Interpretation *string
	Status         *string
}

// ---------------------------------------------------------------------------
// Service interface
// ---------------------------------------------------------------------------

type Service interface {
	// Patient CRUD
	Create(ctx context.Context, clinicID uuid.UUID, req CreatePatientRequest) (*repo.Patient, error)
	GetByID(ctx context.Context, clinicID, patientID uuid.UUID) (*repo.Patient, error)
	List(ctx context.Context, clinicID uuid.UUID, req ListPatientsRequest) (*PaginatedResult[*repo.Patient], error)
	Update(ctx context.Context, clinicID, patientID uuid.UUID, req UpdatePatientRequest) (*repo.Patient, error)

	// Reports
	CreateReport(ctx context.Context, clinicID, patientID, therapistMemberID uuid.UUID, req CreateReportRequest) (*repo.PatientReport, error)
	ListReports(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientReport, error)
	GetReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID) (*repo.PatientReport, error)
	UpdateReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID, req UpdateReportRequest) (*repo.PatientReport, error)
	DeleteReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID) error

	// Prescriptions
	CreatePrescription(ctx context.Context, clinicID, patientID, therapistMemberID uuid.UUID, req CreatePrescriptionRequest) (*repo.PatientPrescription, error)
	ListPrescriptions(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientPrescription, error)
	UpdatePrescription(ctx context.Context, clinicID, patientID, prescriptionID uuid.UUID, req UpdatePrescriptionRequest) (*repo.PatientPrescription, error)

	// Tests
	CreateTest(ctx context.Context, clinicID, patientID uuid.UUID, req CreateTestRequest) (*repo.PatientTest, error)
	ListTests(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientTest, error)
	UpdateTest(ctx context.Context, clinicID, patientID, testID uuid.UUID, req UpdateTestRequest) (*repo.PatientTest, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type patientService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &patientService{db: db}
}

// ---------------------------------------------------------------------------
// Patient CRUD
// ---------------------------------------------------------------------------

func (s *patientService) Create(ctx context.Context, clinicID uuid.UUID, req CreatePatientRequest) (*repo.Patient, error) {
	// Check uniqueness
	exists, err := s.db.Patient.Query().
		Where(entpatient.ClinicID(clinicID), entpatient.UserID(req.UserID), entpatient.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check patient: %w", err)
	}
	if exists {
		return nil, ErrPatientAlreadyExists
	}

	c := s.db.Patient.Create().
		SetClinicID(clinicID).
		SetUserID(req.UserID).
		SetIsChild(req.IsChild)

	if req.PrimaryTherapistID != nil {
		c = c.SetPrimaryTherapistID(*req.PrimaryTherapistID)
	}
	if req.FileNumber != nil {
		c = c.SetNillableFileNumber(req.FileNumber)
	}
	if req.Notes != nil {
		c = c.SetNillableNotes(req.Notes)
	}
	if req.ReferralSource != nil {
		c = c.SetNillableReferralSource(req.ReferralSource)
	}
	if req.ChiefComplaint != nil {
		c = c.SetNillableChiefComplaint(req.ChiefComplaint)
	}
	if req.IsChild {
		if req.ChildBirthDate != nil {
			c = c.SetNillableChildBirthDate(req.ChildBirthDate)
		}
		if req.ChildSchool != nil {
			c = c.SetNillableChildSchool(req.ChildSchool)
		}
		if req.ChildGrade != nil {
			c = c.SetNillableChildGrade(req.ChildGrade)
		}
		if req.ParentName != nil {
			c = c.SetNillableParentName(req.ParentName)
		}
		if req.ParentPhone != nil {
			c = c.SetNillableParentPhone(req.ParentPhone)
		}
		if req.ParentRelation != nil {
			c = c.SetNillableParentRelation(req.ParentRelation)
		}
	}

	return c.Save(ctx)
}

func (s *patientService) GetByID(ctx context.Context, clinicID, patientID uuid.UUID) (*repo.Patient, error) {
	p, err := s.db.Patient.Query().
		Where(entpatient.ID(patientID), entpatient.ClinicID(clinicID), entpatient.DeletedAtIsNil()).
		WithUser().
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrPatientNotFound
		}
		return nil, fmt.Errorf("get patient: %w", err)
	}
	return p, nil
}

func (s *patientService) List(ctx context.Context, clinicID uuid.UUID, req ListPatientsRequest) (*PaginatedResult[*repo.Patient], error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	offset := (req.Page - 1) * req.PerPage

	q := s.db.Patient.Query().
		Where(entpatient.ClinicID(clinicID), entpatient.DeletedAtIsNil())

	if req.TherapistID != nil {
		q = q.Where(entpatient.PrimaryTherapistID(*req.TherapistID))
	}
	if req.Status != nil {
		q = q.Where(entpatient.StatusEQ(entpatient.Status(*req.Status)))
	}
	if req.PaymentStatus != nil {
		q = q.Where(entpatient.PaymentStatusEQ(entpatient.PaymentStatus(*req.PaymentStatus)))
	}
	if req.HasDiscount != nil {
		q = q.Where(entpatient.HasDiscount(*req.HasDiscount))
	}

	// Sorting
	if req.Order == "asc" {
		q = q.Order(entpatient.ByCreatedAt(sql.OrderAsc()))
	} else {
		q = q.Order(entpatient.ByCreatedAt(sql.OrderDesc()))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count patients: %w", err)
	}

	patients, err := q.WithUser().Offset(offset).Limit(req.PerPage).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}

	totalPages := (total + req.PerPage - 1) / req.PerPage
	return &PaginatedResult[*repo.Patient]{
		Data:       patients,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *patientService) Update(ctx context.Context, clinicID, patientID uuid.UUID, req UpdatePatientRequest) (*repo.Patient, error) {
	p, err := s.GetByID(ctx, clinicID, patientID)
	if err != nil {
		return nil, err
	}

	u := s.db.Patient.UpdateOne(p)

	if req.PrimaryTherapistID != nil {
		u = u.SetPrimaryTherapistID(*req.PrimaryTherapistID)
	}
	if req.FileNumber != nil {
		u = u.SetNillableFileNumber(req.FileNumber)
	}
	if req.Status != nil {
		u = u.SetStatus(entpatient.Status(*req.Status))
	}
	if req.HasDiscount != nil {
		u = u.SetHasDiscount(*req.HasDiscount)
	}
	if req.DiscountPercent != nil {
		u = u.SetDiscountPercent(*req.DiscountPercent)
	}
	if req.PaymentStatus != nil {
		u = u.SetPaymentStatus(entpatient.PaymentStatus(*req.PaymentStatus))
	}
	if req.Notes != nil {
		u = u.SetNillableNotes(req.Notes)
	}
	if req.ReferralSource != nil {
		u = u.SetNillableReferralSource(req.ReferralSource)
	}
	if req.ChiefComplaint != nil {
		u = u.SetNillableChiefComplaint(req.ChiefComplaint)
	}
	if req.IsChild != nil {
		u = u.SetIsChild(*req.IsChild)
	}
	if req.ChildBirthDate != nil {
		u = u.SetNillableChildBirthDate(req.ChildBirthDate)
	}
	if req.ChildSchool != nil {
		u = u.SetNillableChildSchool(req.ChildSchool)
	}
	if req.ChildGrade != nil {
		u = u.SetNillableChildGrade(req.ChildGrade)
	}
	if req.ParentName != nil {
		u = u.SetNillableParentName(req.ParentName)
	}
	if req.ParentPhone != nil {
		u = u.SetNillableParentPhone(req.ParentPhone)
	}
	if req.ParentRelation != nil {
		u = u.SetNillableParentRelation(req.ParentRelation)
	}

	return u.Save(ctx)
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

func (s *patientService) CreateReport(ctx context.Context, clinicID, patientID, therapistMemberID uuid.UUID, req CreateReportRequest) (*repo.PatientReport, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}

	c := s.db.PatientReport.Create().
		SetPatientID(patientID).
		SetClinicID(clinicID).
		SetTherapistID(therapistMemberID)

	if req.AppointmentID != nil {
		c = c.SetNillableAppointmentID(req.AppointmentID)
	}
	if req.Title != nil {
		c = c.SetNillableTitle(req.Title)
	}
	if req.Content != nil {
		c = c.SetNillableContent(req.Content)
	}
	if req.ReportDate != nil {
		c = c.SetReportDate(*req.ReportDate)
	}

	return c.Save(ctx)
}

func (s *patientService) ListReports(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientReport, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}
	return s.db.PatientReport.Query().
		Where(entreport.PatientID(patientID), entreport.ClinicID(clinicID)).
		Order(entreport.ByReportDate(sql.OrderDesc())).
		All(ctx)
}

func (s *patientService) GetReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID) (*repo.PatientReport, error) {
	r, err := s.db.PatientReport.Query().
		Where(entreport.ID(reportID), entreport.PatientID(patientID), entreport.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrReportNotFound
		}
		return nil, fmt.Errorf("get report: %w", err)
	}
	return r, nil
}

func (s *patientService) UpdateReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID, req UpdateReportRequest) (*repo.PatientReport, error) {
	r, err := s.GetReport(ctx, clinicID, patientID, reportID)
	if err != nil {
		return nil, err
	}

	u := s.db.PatientReport.UpdateOne(r)
	if req.Title != nil {
		u = u.SetNillableTitle(req.Title)
	}
	if req.Content != nil {
		u = u.SetNillableContent(req.Content)
	}
	if req.ReportDate != nil {
		u = u.SetReportDate(*req.ReportDate)
	}
	return u.Save(ctx)
}

func (s *patientService) DeleteReport(ctx context.Context, clinicID, patientID, reportID uuid.UUID) error {
	r, err := s.GetReport(ctx, clinicID, patientID, reportID)
	if err != nil {
		return err
	}
	return s.db.PatientReport.DeleteOne(r).Exec(ctx)
}

// ---------------------------------------------------------------------------
// Prescriptions
// ---------------------------------------------------------------------------

func (s *patientService) CreatePrescription(ctx context.Context, clinicID, patientID, therapistMemberID uuid.UUID, req CreatePrescriptionRequest) (*repo.PatientPrescription, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}

	c := s.db.PatientPrescription.Create().
		SetPatientID(patientID).
		SetClinicID(clinicID).
		SetTherapistID(therapistMemberID)

	if req.Title != nil {
		c = c.SetNillableTitle(req.Title)
	}
	if req.Notes != nil {
		c = c.SetNillableNotes(req.Notes)
	}
	if req.FileKey != nil {
		c = c.SetNillableFileKey(req.FileKey)
	}
	if req.FileName != nil {
		c = c.SetNillableFileName(req.FileName)
	}
	if req.PrescribedDate != nil {
		c = c.SetPrescribedDate(*req.PrescribedDate)
	}

	return c.Save(ctx)
}

func (s *patientService) ListPrescriptions(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientPrescription, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}
	return s.db.PatientPrescription.Query().
		Where(entprescription.PatientID(patientID), entprescription.ClinicID(clinicID)).
		Order(entprescription.ByPrescribedDate(sql.OrderDesc())).
		All(ctx)
}

func (s *patientService) UpdatePrescription(ctx context.Context, clinicID, patientID, prescriptionID uuid.UUID, req UpdatePrescriptionRequest) (*repo.PatientPrescription, error) {
	rx, err := s.db.PatientPrescription.Query().
		Where(entprescription.ID(prescriptionID), entprescription.PatientID(patientID), entprescription.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrPrescriptionNotFound
		}
		return nil, fmt.Errorf("get prescription: %w", err)
	}

	u := s.db.PatientPrescription.UpdateOne(rx)
	if req.Title != nil {
		u = u.SetNillableTitle(req.Title)
	}
	if req.Notes != nil {
		u = u.SetNillableNotes(req.Notes)
	}
	if req.FileKey != nil {
		u = u.SetNillableFileKey(req.FileKey)
	}
	if req.FileName != nil {
		u = u.SetNillableFileName(req.FileName)
	}
	if req.PrescribedDate != nil {
		u = u.SetPrescribedDate(*req.PrescribedDate)
	}
	return u.Save(ctx)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func (s *patientService) CreateTest(ctx context.Context, clinicID, patientID uuid.UUID, req CreateTestRequest) (*repo.PatientTest, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}

	c := s.db.PatientTest.Create().
		SetPatientID(patientID).
		SetClinicID(clinicID)

	if req.TestID != nil {
		c = c.SetNillableTestID(req.TestID)
	}
	if req.AdministeredBy != nil {
		c = c.SetNillableAdministeredBy(req.AdministeredBy)
	}
	if req.TestName != nil {
		c = c.SetNillableTestName(req.TestName)
	}
	if req.RawScores != nil {
		c = c.SetRawScores(req.RawScores)
	}
	if req.ComputedScores != nil {
		c = c.SetComputedScores(req.ComputedScores)
	}
	if req.Interpretation != nil {
		c = c.SetNillableInterpretation(req.Interpretation)
	}
	if req.TestDate != nil {
		c = c.SetTestDate(*req.TestDate)
	}

	return c.Save(ctx)
}

func (s *patientService) ListTests(ctx context.Context, clinicID, patientID uuid.UUID) ([]*repo.PatientTest, error) {
	if _, err := s.GetByID(ctx, clinicID, patientID); err != nil {
		return nil, err
	}
	return s.db.PatientTest.Query().
		Where(enttest.PatientID(patientID), enttest.ClinicID(clinicID)).
		Order(enttest.ByTestDate(sql.OrderDesc())).
		All(ctx)
}

func (s *patientService) UpdateTest(ctx context.Context, clinicID, patientID, testID uuid.UUID, req UpdateTestRequest) (*repo.PatientTest, error) {
	t, err := s.db.PatientTest.Query().
		Where(enttest.ID(testID), enttest.PatientID(patientID), enttest.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrPatientTestNotFound
		}
		return nil, fmt.Errorf("get test: %w", err)
	}

	u := s.db.PatientTest.UpdateOne(t)
	if req.AdministeredBy != nil {
		u = u.SetNillableAdministeredBy(req.AdministeredBy)
	}
	if req.RawScores != nil {
		u = u.SetRawScores(req.RawScores)
	}
	if req.ComputedScores != nil {
		u = u.SetComputedScores(req.ComputedScores)
	}
	if req.Interpretation != nil {
		u = u.SetNillableInterpretation(req.Interpretation)
	}
	if req.Status != nil {
		u = u.SetStatus(enttest.Status(*req.Status))
	}
	return u.Save(ctx)
}
