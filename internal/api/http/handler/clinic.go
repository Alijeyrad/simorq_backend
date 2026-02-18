package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/clinic"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type ClinicHandler struct {
	svc clinic.Service
}

func NewClinicHandler(svc clinic.Service) *ClinicHandler {
	return &ClinicHandler{svc: svc}
}

// GET /api/v1/clinics
func (h *ClinicHandler) List(c fiber.Ctx) error {
	var q struct {
		Page    int `query:"page"`
		PerPage int `query:"per_page"`
	}
	if err := c.Bind().Query(&q); err != nil || q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 20
	}

	result, err := h.svc.ListClinics(c.Context(), clinic.ListClinicsRequest{
		Page:    q.Page,
		PerPage: q.PerPage,
	})
	if err != nil {
		return internalError(c)
	}

	return ok(c, fiber.Map{
		"clinics":     result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// GET /api/v1/clinics/:slug
func (h *ClinicHandler) GetBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	cl, err := h.svc.GetClinicBySlug(c.Context(), slug)
	if err != nil {
		return mapClinicError(c, err)
	}
	return ok(c, cl)
}

// POST /api/v1/clinics
func (h *ClinicHandler) Create(c fiber.Ctx) error {
	claims, ok := pasetotoken.ClaimsFromFiber(c)
	if !ok {
		return unauthorized(c)
	}

	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		City        string `json:"city"`
		Province    string `json:"province"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Name == "" {
		return badRequest(c, "name is required")
	}

	cl, err := h.svc.CreateClinic(c.Context(), claims.UserID, clinic.CreateClinicRequest{
		Name:        body.Name,
		Slug:        body.Slug,
		Description: body.Description,
		Phone:       body.Phone,
		Address:     body.Address,
		City:        body.City,
		Province:    body.Province,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return created(c, cl)
}

// PATCH /api/v1/clinics/:id
func (h *ClinicHandler) Update(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Phone       *string `json:"phone"`
		Address     *string `json:"address"`
		City        *string `json:"city"`
		Province    *string `json:"province"`
		LogoKey     *string `json:"logo_key"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	cl, err := h.svc.UpdateClinic(c.Context(), clinicID, clinic.UpdateClinicRequest{
		Name:        body.Name,
		Description: body.Description,
		Phone:       body.Phone,
		Address:     body.Address,
		City:        body.City,
		Province:    body.Province,
		LogoKey:     body.LogoKey,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, cl)
}

// GET /api/v1/clinics/:id/settings
func (h *ClinicHandler) GetSettings(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	st, err := h.svc.GetSettings(c.Context(), clinicID)
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, st)
}

// PATCH /api/v1/clinics/:id/settings
func (h *ClinicHandler) UpdateSettings(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	var body struct {
		ReservationFeeAmount      *int64         `json:"reservation_fee_amount"`
		ReservationFeePercent     *int           `json:"reservation_fee_percent"`
		CancellationWindowHours   *int           `json:"cancellation_window_hours"`
		CancellationFeeAmount     *int64         `json:"cancellation_fee_amount"`
		CancellationFeePercent    *int           `json:"cancellation_fee_percent"`
		AllowClientSelfBook       *bool          `json:"allow_client_self_book"`
		DefaultSessionDurationMin *int           `json:"default_session_duration_min"`
		DefaultSessionPrice       *int64         `json:"default_session_price"`
		WorkingHours              map[string]any `json:"working_hours"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	st, err := h.svc.UpdateSettings(c.Context(), clinicID, clinic.UpdateSettingsRequest{
		ReservationFeeAmount:      body.ReservationFeeAmount,
		ReservationFeePercent:     body.ReservationFeePercent,
		CancellationWindowHours:   body.CancellationWindowHours,
		CancellationFeeAmount:     body.CancellationFeeAmount,
		CancellationFeePercent:    body.CancellationFeePercent,
		AllowClientSelfBook:       body.AllowClientSelfBook,
		DefaultSessionDurationMin: body.DefaultSessionDurationMin,
		DefaultSessionPrice:       body.DefaultSessionPrice,
		WorkingHours:              body.WorkingHours,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, st)
}

// GET /api/v1/clinics/:id/members
func (h *ClinicHandler) ListMembers(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	members, err := h.svc.ListMembers(c.Context(), clinicID)
	if err != nil {
		return internalError(c)
	}

	return ok(c, members)
}

// POST /api/v1/clinics/:id/members
func (h *ClinicHandler) AddMember(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	var body struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		return badRequest(c, "invalid user_id")
	}

	m, err := h.svc.AddMember(c.Context(), clinicID, clinic.AddMemberRequest{
		UserID: userID,
		Role:   body.Role,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return created(c, m)
}

// PATCH /api/v1/clinics/:id/members/:mid
func (h *ClinicHandler) UpdateMember(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	memberID, err := uuid.Parse(c.Params("mid"))
	if err != nil {
		return badRequest(c, "invalid member id")
	}

	var body struct {
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	m, err := h.svc.UpdateMember(c.Context(), clinicID, memberID, clinic.UpdateMemberRequest{
		Role:     body.Role,
		IsActive: body.IsActive,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, m)
}

// DELETE /api/v1/clinics/:id/members/:mid
func (h *ClinicHandler) RemoveMember(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	memberID, err := uuid.Parse(c.Params("mid"))
	if err != nil {
		return badRequest(c, "invalid member id")
	}

	if err := h.svc.RemoveMember(c.Context(), clinicID, memberID); err != nil {
		return mapClinicError(c, err)
	}

	return noContent(c)
}

// GET /api/v1/clinics/:id/therapists
func (h *ClinicHandler) ListTherapists(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	therapists, err := h.svc.ListTherapists(c.Context(), clinicID)
	if err != nil {
		return internalError(c)
	}

	return ok(c, therapists)
}

// GET /api/v1/clinics/:id/permissions
func (h *ClinicHandler) GetPermissions(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	perms, err := h.svc.GetPermissions(c.Context(), clinicID)
	if err != nil {
		return internalError(c)
	}

	return ok(c, perms)
}

// PATCH /api/v1/clinics/:id/permissions
func (h *ClinicHandler) SetPermission(c fiber.Ctx) error {
	clinicID, err := parseClinicID(c)
	if err != nil {
		return badRequest(c, "invalid clinic id")
	}

	var body struct {
		UserID       string  `json:"user_id"`
		ResourceType string  `json:"resource_type"`
		ResourceID   *string `json:"resource_id"`
		Action       string  `json:"action"`
		Granted      bool    `json:"granted"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.UserID == "" || body.ResourceType == "" || body.Action == "" {
		return badRequest(c, "user_id, resource_type, and action are required")
	}

	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		return badRequest(c, "invalid user_id")
	}

	req := clinic.SetPermissionRequest{
		UserID:       userID,
		ResourceType: body.ResourceType,
		Action:       body.Action,
		Granted:      body.Granted,
	}
	if body.ResourceID != nil {
		rid, err := uuid.Parse(*body.ResourceID)
		if err != nil {
			return badRequest(c, "invalid resource_id")
		}
		req.ResourceID = &rid
	}

	if err := h.svc.SetPermission(c.Context(), clinicID, req); err != nil {
		return mapClinicError(c, err)
	}

	return noContent(c)
}

// GET /api/v1/clinics/:id/members/:mid/profile
func (h *ClinicHandler) GetTherapistProfile(c fiber.Ctx) error {
	memberID, err := uuid.Parse(c.Params("mid"))
	if err != nil {
		return badRequest(c, "invalid member id")
	}

	profile, err := h.svc.GetTherapistProfile(c.Context(), memberID)
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, profile)
}

// PATCH /api/v1/clinics/:id/members/:mid/profile
func (h *ClinicHandler) UpdateTherapistProfile(c fiber.Ctx) error {
	memberID, err := uuid.Parse(c.Params("mid"))
	if err != nil {
		return badRequest(c, "invalid member id")
	}

	var body struct {
		Education          *string  `json:"education"`
		PsychologyLicense  *string  `json:"psychology_license"`
		Approach           *string  `json:"approach"`
		Specialties        []string `json:"specialties"`
		Bio                *string  `json:"bio"`
		SessionPrice       *int64   `json:"session_price"`
		SessionDurationMin *int     `json:"session_duration_min"`
		IsAccepting        *bool    `json:"is_accepting"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	profile, err := h.svc.UpdateTherapistProfile(c.Context(), memberID, clinic.UpdateTherapistProfileRequest{
		Education:          body.Education,
		PsychologyLicense:  body.PsychologyLicense,
		Approach:           body.Approach,
		Specialties:        body.Specialties,
		Bio:                body.Bio,
		SessionPrice:       body.SessionPrice,
		SessionDurationMin: body.SessionDurationMin,
		IsAccepting:        body.IsAccepting,
	})
	if err != nil {
		return mapClinicError(c, err)
	}

	return ok(c, profile)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseClinicID(c fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

func mapClinicError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, clinic.ErrClinicNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, clinic.ErrSlugAlreadyExists):
		return conflict(c, err.Error())
	case errors.Is(err, clinic.ErrMemberNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, clinic.ErrAlreadyMember):
		return conflict(c, err.Error())
	case errors.Is(err, clinic.ErrInvalidRole):
		return badRequest(c, err.Error())
	case errors.Is(err, clinic.ErrCannotRemoveOwner):
		return badRequest(c, err.Error())
	case errors.Is(err, clinic.ErrTherapistProfileNotFound):
		return notFound(c, err.Error())
	default:
		return internalError(c)
	}
}
