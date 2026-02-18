package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/api/http/middleware"
	"github.com/Alijeyrad/simorq_backend/internal/service/patient"
)

type PatientHandler struct {
	svc patient.Service
}

func NewPatientHandler(svc patient.Service) *PatientHandler {
	return &PatientHandler{svc: svc}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func clinicIDFromLocals(c fiber.Ctx) (uuid.UUID, bool) {
	s, hasKey := c.Locals(middleware.LocalsClinicID).(string)
	if !hasKey || s == "" {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(s)
	return id, err == nil
}

func memberIDFromLocals(c fiber.Ctx) (uuid.UUID, bool) {
	s, hasKey := c.Locals(middleware.LocalsMemberID).(string)
	if !hasKey || s == "" {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(s)
	return id, err == nil
}

func mapPatientError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, patient.ErrPatientNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, patient.ErrPatientAlreadyExists):
		return conflict(c, err.Error())
	case errors.Is(err, patient.ErrReportNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, patient.ErrPrescriptionNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, patient.ErrPatientTestNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, patient.ErrInvalidStatus):
		return badRequest(c, err.Error())
	case errors.Is(err, patient.ErrAccessDenied):
		return forbidden(c)
	default:
		return internalError(c)
	}
}

// ---------------------------------------------------------------------------
// Patient CRUD
// ---------------------------------------------------------------------------

// GET /patients
func (h *PatientHandler) List(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var q struct {
		Page          int     `query:"page"`
		PerPage       int     `query:"per_page"`
		TherapistID   string  `query:"therapist_id"`
		Status        string  `query:"status"`
		PaymentStatus string  `query:"payment_status"`
		HasDiscount   *bool   `query:"has_discount"`
		Sort          string  `query:"sort"`
		Order         string  `query:"order"`
	}
	_ = c.Bind().Query(&q)

	req := patient.ListPatientsRequest{
		Page:        q.Page,
		PerPage:     q.PerPage,
		Sort:        q.Sort,
		Order:       q.Order,
		HasDiscount: q.HasDiscount,
	}
	if q.TherapistID != "" {
		id, err := uuid.Parse(q.TherapistID)
		if err != nil {
			return badRequest(c, "invalid therapist_id")
		}
		req.TherapistID = &id
	}
	if q.Status != "" {
		req.Status = &q.Status
	}
	if q.PaymentStatus != "" {
		req.PaymentStatus = &q.PaymentStatus
	}

	result, err := h.svc.List(c.Context(), clinicID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, fiber.Map{
		"patients":    result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// POST /patients
func (h *PatientHandler) Create(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var body struct {
		UserID             string     `json:"user_id"`
		PrimaryTherapistID *string    `json:"primary_therapist_id"`
		FileNumber         *string    `json:"file_number"`
		Notes              *string    `json:"notes"`
		ReferralSource     *string    `json:"referral_source"`
		ChiefComplaint     *string    `json:"chief_complaint"`
		IsChild            bool       `json:"is_child"`
		ChildBirthDate     *time.Time `json:"child_birth_date"`
		ChildSchool        *string    `json:"child_school"`
		ChildGrade         *string    `json:"child_grade"`
		ParentName         *string    `json:"parent_name"`
		ParentPhone        *string    `json:"parent_phone"`
		ParentRelation     *string    `json:"parent_relation"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.UserID == "" {
		return badRequest(c, "user_id is required")
	}

	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		return badRequest(c, "invalid user_id")
	}

	req := patient.CreatePatientRequest{
		UserID:         userID,
		FileNumber:     body.FileNumber,
		Notes:          body.Notes,
		ReferralSource: body.ReferralSource,
		ChiefComplaint: body.ChiefComplaint,
		IsChild:        body.IsChild,
		ChildBirthDate: body.ChildBirthDate,
		ChildSchool:    body.ChildSchool,
		ChildGrade:     body.ChildGrade,
		ParentName:     body.ParentName,
		ParentPhone:    body.ParentPhone,
		ParentRelation: body.ParentRelation,
	}
	if body.PrimaryTherapistID != nil {
		id, err := uuid.Parse(*body.PrimaryTherapistID)
		if err != nil {
			return badRequest(c, "invalid primary_therapist_id")
		}
		req.PrimaryTherapistID = &id
	}

	p, err := h.svc.Create(c.Context(), clinicID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return created(c, p)
}

// GET /patients/:id
func (h *PatientHandler) Get(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	p, err := h.svc.GetByID(c.Context(), clinicID, patientID)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, p)
}

// PATCH /patients/:id
func (h *PatientHandler) Update(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	var body struct {
		PrimaryTherapistID *string    `json:"primary_therapist_id"`
		FileNumber         *string    `json:"file_number"`
		Status             *string    `json:"status"`
		HasDiscount        *bool      `json:"has_discount"`
		DiscountPercent    *int       `json:"discount_percent"`
		PaymentStatus      *string    `json:"payment_status"`
		Notes              *string    `json:"notes"`
		ReferralSource     *string    `json:"referral_source"`
		ChiefComplaint     *string    `json:"chief_complaint"`
		IsChild            *bool      `json:"is_child"`
		ChildBirthDate     *time.Time `json:"child_birth_date"`
		ChildSchool        *string    `json:"child_school"`
		ChildGrade         *string    `json:"child_grade"`
		ParentName         *string    `json:"parent_name"`
		ParentPhone        *string    `json:"parent_phone"`
		ParentRelation     *string    `json:"parent_relation"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	req := patient.UpdatePatientRequest{
		FileNumber:      body.FileNumber,
		Status:          body.Status,
		HasDiscount:     body.HasDiscount,
		DiscountPercent: body.DiscountPercent,
		PaymentStatus:   body.PaymentStatus,
		Notes:           body.Notes,
		ReferralSource:  body.ReferralSource,
		ChiefComplaint:  body.ChiefComplaint,
		IsChild:         body.IsChild,
		ChildBirthDate:  body.ChildBirthDate,
		ChildSchool:     body.ChildSchool,
		ChildGrade:      body.ChildGrade,
		ParentName:      body.ParentName,
		ParentPhone:     body.ParentPhone,
		ParentRelation:  body.ParentRelation,
	}
	if body.PrimaryTherapistID != nil {
		id, err := uuid.Parse(*body.PrimaryTherapistID)
		if err != nil {
			return badRequest(c, "invalid primary_therapist_id")
		}
		req.PrimaryTherapistID = &id
	}

	p, err := h.svc.Update(c.Context(), clinicID, patientID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, p)
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

// GET /patients/:id/reports
func (h *PatientHandler) ListReports(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	reports, err := h.svc.ListReports(c.Context(), clinicID, patientID)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, reports)
}

// POST /patients/:id/reports
func (h *PatientHandler) CreateReport(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	memberID, ok := memberIDFromLocals(c)
	if !ok {
		return unauthorized(c)
	}

	var body struct {
		AppointmentID *string    `json:"appointment_id"`
		Title         *string    `json:"title"`
		Content       *string    `json:"content"`
		ReportDate    *time.Time `json:"report_date"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	req := patient.CreateReportRequest{
		Title:      body.Title,
		Content:    body.Content,
		ReportDate: body.ReportDate,
	}
	if body.AppointmentID != nil {
		id, err := uuid.Parse(*body.AppointmentID)
		if err != nil {
			return badRequest(c, "invalid appointment_id")
		}
		req.AppointmentID = &id
	}

	r, err := h.svc.CreateReport(c.Context(), clinicID, patientID, memberID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return created(c, r)
}

// PATCH /patients/:id/reports/:rid
func (h *PatientHandler) UpdateReport(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	reportID, err := uuid.Parse(c.Params("rid"))
	if err != nil {
		return badRequest(c, "invalid report id")
	}

	var body struct {
		Title      *string    `json:"title"`
		Content    *string    `json:"content"`
		ReportDate *time.Time `json:"report_date"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	r, err := h.svc.UpdateReport(c.Context(), clinicID, patientID, reportID, patient.UpdateReportRequest{
		Title:      body.Title,
		Content:    body.Content,
		ReportDate: body.ReportDate,
	})
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, r)
}

// DELETE /patients/:id/reports/:rid
func (h *PatientHandler) DeleteReport(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	reportID, err := uuid.Parse(c.Params("rid"))
	if err != nil {
		return badRequest(c, "invalid report id")
	}

	if err := h.svc.DeleteReport(c.Context(), clinicID, patientID, reportID); err != nil {
		return mapPatientError(c, err)
	}

	return noContent(c)
}

// ---------------------------------------------------------------------------
// Prescriptions
// ---------------------------------------------------------------------------

// GET /patients/:id/prescriptions
func (h *PatientHandler) ListPrescriptions(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	rxs, err := h.svc.ListPrescriptions(c.Context(), clinicID, patientID)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, rxs)
}

// POST /patients/:id/prescriptions
func (h *PatientHandler) CreatePrescription(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	memberID, ok := memberIDFromLocals(c)
	if !ok {
		return unauthorized(c)
	}

	var body struct {
		Title          *string    `json:"title"`
		Notes          *string    `json:"notes"`
		FileKey        *string    `json:"file_key"`
		FileName       *string    `json:"file_name"`
		PrescribedDate *time.Time `json:"prescribed_date"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	rx, err := h.svc.CreatePrescription(c.Context(), clinicID, patientID, memberID, patient.CreatePrescriptionRequest{
		Title:          body.Title,
		Notes:          body.Notes,
		FileKey:        body.FileKey,
		FileName:       body.FileName,
		PrescribedDate: body.PrescribedDate,
	})
	if err != nil {
		return mapPatientError(c, err)
	}

	return created(c, rx)
}

// PATCH /patients/:id/prescriptions/:pid
func (h *PatientHandler) UpdatePrescription(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	prescriptionID, err := uuid.Parse(c.Params("pid"))
	if err != nil {
		return badRequest(c, "invalid prescription id")
	}

	var body struct {
		Title          *string    `json:"title"`
		Notes          *string    `json:"notes"`
		FileKey        *string    `json:"file_key"`
		FileName       *string    `json:"file_name"`
		PrescribedDate *time.Time `json:"prescribed_date"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	rx, err := h.svc.UpdatePrescription(c.Context(), clinicID, patientID, prescriptionID, patient.UpdatePrescriptionRequest{
		Title:          body.Title,
		Notes:          body.Notes,
		FileKey:        body.FileKey,
		FileName:       body.FileName,
		PrescribedDate: body.PrescribedDate,
	})
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, rx)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// GET /patients/:id/tests
func (h *PatientHandler) ListTests(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	tests, err := h.svc.ListTests(c.Context(), clinicID, patientID)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, tests)
}

// POST /patients/:id/tests
func (h *PatientHandler) CreateTest(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	var body struct {
		TestID         *string        `json:"test_id"`
		AdministeredBy *string        `json:"administered_by"`
		TestName       *string        `json:"test_name"`
		RawScores      map[string]any `json:"raw_scores"`
		ComputedScores map[string]any `json:"computed_scores"`
		Interpretation *string        `json:"interpretation"`
		TestDate       *time.Time     `json:"test_date"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	req := patient.CreateTestRequest{
		TestName:       body.TestName,
		RawScores:      body.RawScores,
		ComputedScores: body.ComputedScores,
		Interpretation: body.Interpretation,
		TestDate:       body.TestDate,
	}
	if body.TestID != nil {
		id, err := uuid.Parse(*body.TestID)
		if err != nil {
			return badRequest(c, "invalid test_id")
		}
		req.TestID = &id
	}
	if body.AdministeredBy != nil {
		id, err := uuid.Parse(*body.AdministeredBy)
		if err != nil {
			return badRequest(c, "invalid administered_by")
		}
		req.AdministeredBy = &id
	}

	t, err := h.svc.CreateTest(c.Context(), clinicID, patientID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return created(c, t)
}

// PATCH /patients/:id/tests/:tid
func (h *PatientHandler) UpdateTest(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	patientID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	testID, err := uuid.Parse(c.Params("tid"))
	if err != nil {
		return badRequest(c, "invalid test id")
	}

	var body struct {
		AdministeredBy *string        `json:"administered_by"`
		RawScores      map[string]any `json:"raw_scores"`
		ComputedScores map[string]any `json:"computed_scores"`
		Interpretation *string        `json:"interpretation"`
		Status         *string        `json:"status"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	req := patient.UpdateTestRequest{
		RawScores:      body.RawScores,
		ComputedScores: body.ComputedScores,
		Interpretation: body.Interpretation,
		Status:         body.Status,
	}
	if body.AdministeredBy != nil {
		id, err := uuid.Parse(*body.AdministeredBy)
		if err != nil {
			return badRequest(c, "invalid administered_by")
		}
		req.AdministeredBy = &id
	}

	t, err := h.svc.UpdateTest(c.Context(), clinicID, patientID, testID, req)
	if err != nil {
		return mapPatientError(c, err)
	}

	return ok(c, t)
}
