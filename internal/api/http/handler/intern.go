package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/intern"
)

type InternHandler struct {
	svc intern.Service
}

func NewInternHandler(svc intern.Service) *InternHandler {
	return &InternHandler{svc: svc}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func pageParams(c fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func mapInternError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, intern.ErrNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, intern.ErrUnauthorized):
		return forbidden(c)
	case errors.Is(err, intern.ErrAccessAlreadyGranted):
		return conflict(c, err.Error())
	default:
		return internalError(c)
	}
}

// ---------------------------------------------------------------------------
// Intern-facing handlers
// ---------------------------------------------------------------------------

func (h *InternHandler) GetMyProfile(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	profile, err := h.svc.GetMyProfile(c.Context(), memberID)
	if err != nil {
		return mapInternError(c, err)
	}
	return ok(c, profile)
}

func (h *InternHandler) UpsertMyProfile(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		InternshipYear *int        `json:"internship_year"`
		SupervisorIDs  []uuid.UUID `json:"supervisor_ids"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	profile, err := h.svc.UpsertMyProfile(c.Context(), memberID, intern.UpsertProfileRequest{
		InternshipYear: body.InternshipYear,
		SupervisorIDs:  body.SupervisorIDs,
	})
	if err != nil {
		return mapInternError(c, err)
	}
	return ok(c, profile)
}

func (h *InternHandler) ListMyTasks(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	page, perPage := pageParams(c)
	tasks, err := h.svc.ListMyTasks(c.Context(), memberID, page, perPage)
	if err != nil {
		return internalError(c)
	}
	return ok(c, tasks)
}

func (h *InternHandler) CreateTask(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid2 := memberIDFromLocals(c)
	if !valid2 {
		return unauthorized(c)
	}

	var body struct {
		Title   string  `json:"title"`
		Caption *string `json:"caption"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Title == "" {
		return badRequest(c, "title is required")
	}

	task, err := h.svc.CreateTask(c.Context(), clinicID, memberID, intern.CreateTaskRequest{
		Title:   body.Title,
		Caption: body.Caption,
	})
	if err != nil {
		return internalError(c)
	}
	return created(c, task)
}

func (h *InternHandler) GetTask(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return badRequest(c, "invalid task id")
	}

	task, err := h.svc.GetTask(c.Context(), taskID, memberID)
	if err != nil {
		return mapInternError(c, err)
	}
	return ok(c, task)
}

func (h *InternHandler) UpdateTask(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return badRequest(c, "invalid task id")
	}

	var body struct {
		Title   *string `json:"title"`
		Caption *string `json:"caption"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	task, err := h.svc.UpdateTask(c.Context(), taskID, memberID, intern.UpdateTaskRequest{
		Title:   body.Title,
		Caption: body.Caption,
	})
	if err != nil {
		return mapInternError(c, err)
	}
	return ok(c, task)
}

func (h *InternHandler) AddTaskFile(c fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return badRequest(c, "invalid task id")
	}

	var body struct {
		FileKey  string  `json:"file_key"`
		FileName string  `json:"file_name"`
		FileSize *int64  `json:"file_size"`
		MimeType *string `json:"mime_type"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.FileKey == "" || body.FileName == "" {
		return badRequest(c, "file_key and file_name are required")
	}

	f, err := h.svc.AddTaskFile(c.Context(), taskID, intern.AddFileRequest{
		FileKey:  body.FileKey,
		FileName: body.FileName,
		FileSize: body.FileSize,
		MimeType: body.MimeType,
	})
	if err != nil {
		return internalError(c)
	}
	return created(c, f)
}

func (h *InternHandler) ListMyPatients(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	page, perPage := pageParams(c)
	accesses, err := h.svc.ListMyPatients(c.Context(), memberID, page, perPage)
	if err != nil {
		return internalError(c)
	}
	return ok(c, accesses)
}

// ---------------------------------------------------------------------------
// Admin / supervisor-facing handlers
// ---------------------------------------------------------------------------

func (h *InternHandler) ListInterns(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	page, perPage := pageParams(c)
	profiles, err := h.svc.ListInterns(c.Context(), clinicID, page, perPage)
	if err != nil {
		return internalError(c)
	}
	return ok(c, profiles)
}

func (h *InternHandler) ListInternTasks(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	internID, err := uuid.Parse(c.Params("internID"))
	if err != nil {
		return badRequest(c, "invalid intern id")
	}
	page, perPage := pageParams(c)
	tasks, err := h.svc.ListInternTasks(c.Context(), clinicID, internID, page, perPage)
	if err != nil {
		return internalError(c)
	}
	return ok(c, tasks)
}

func (h *InternHandler) ReviewTask(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return badRequest(c, "invalid task id")
	}

	var body struct {
		Status  string  `json:"status"`
		Comment *string `json:"comment"`
		Grade   *string `json:"grade"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Status == "" {
		return badRequest(c, "status is required")
	}

	task, err := h.svc.ReviewTask(c.Context(), taskID, memberID, intern.ReviewTaskRequest{
		Status:  body.Status,
		Comment: body.Comment,
		Grade:   body.Grade,
	})
	if err != nil {
		return mapInternError(c, err)
	}
	return ok(c, task)
}

func (h *InternHandler) GrantAccess(c fiber.Ctx) error {
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}
	internID, err := uuid.Parse(c.Params("internID"))
	if err != nil {
		return badRequest(c, "invalid intern id")
	}
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	var body struct {
		CanViewFiles    bool `json:"can_view_files"`
		CanWriteReports bool `json:"can_write_reports"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	access, err := h.svc.GrantAccess(c.Context(), internID, patientID, memberID, intern.GrantAccessRequest{
		CanViewFiles:    body.CanViewFiles,
		CanWriteReports: body.CanWriteReports,
	})
	if err != nil {
		return mapInternError(c, err)
	}
	return created(c, access)
}

func (h *InternHandler) RevokeAccess(c fiber.Ctx) error {
	internID, err := uuid.Parse(c.Params("internID"))
	if err != nil {
		return badRequest(c, "invalid intern id")
	}
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return badRequest(c, "invalid patient id")
	}

	if err := h.svc.RevokeAccess(c.Context(), internID, patientID); err != nil {
		return mapInternError(c, err)
	}
	return noContent(c)
}
