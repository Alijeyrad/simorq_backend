package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/appointment"
)

type AppointmentHandler struct {
	svc appointment.Service
}

func NewAppointmentHandler(svc appointment.Service) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

func mapAppointmentError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, appointment.ErrNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, appointment.ErrSlotNotAvailable):
		return conflict(c, err.Error())
	case errors.Is(err, appointment.ErrAlreadyCompleted):
		return conflict(c, err.Error())
	case errors.Is(err, appointment.ErrAlreadyCancelled):
		return conflict(c, err.Error())
	default:
		return internalError(c)
	}
}

// GET /appointments
func (h *AppointmentHandler) List(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var q struct {
		TherapistID string `query:"therapist_id"`
		PatientID   string `query:"patient_id"`
		Status      string `query:"status"`
		From        string `query:"from"`
		To          string `query:"to"`
		Page        int    `query:"page"`
		PerPage     int    `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	req := appointment.ListRequest{
		Page:    q.Page,
		PerPage: q.PerPage,
	}
	if q.TherapistID != "" {
		id, err := uuid.Parse(q.TherapistID)
		if err != nil {
			return badRequest(c, "invalid therapist_id")
		}
		req.TherapistID = &id
	}
	if q.PatientID != "" {
		id, err := uuid.Parse(q.PatientID)
		if err != nil {
			return badRequest(c, "invalid patient_id")
		}
		req.PatientID = &id
	}
	if q.Status != "" {
		req.Status = &q.Status
	}
	if q.From != "" {
		if t, err := time.Parse(time.RFC3339, q.From); err == nil {
			req.From = &t
		}
	}
	if q.To != "" {
		if t, err := time.Parse(time.RFC3339, q.To); err == nil {
			req.To = &t
		}
	}

	appts, err := h.svc.List(c.Context(), clinicID, req)
	if err != nil {
		return mapAppointmentError(c, err)
	}

	return ok(c, appts)
}

// GET /appointments/:id
func (h *AppointmentHandler) GetByID(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	apptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid appointment id")
	}

	appt, err := h.svc.GetByID(c.Context(), clinicID, apptID)
	if err != nil {
		return mapAppointmentError(c, err)
	}

	return ok(c, appt)
}

// POST /appointments
func (h *AppointmentHandler) Book(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var body struct {
		TherapistID    string     `json:"therapist_id"`
		PatientID      string     `json:"patient_id"`
		TimeSlotID     *string    `json:"time_slot_id"`
		StartTime      time.Time  `json:"start_time"`
		EndTime        time.Time  `json:"end_time"`
		SessionPrice   int64      `json:"session_price"`
		ReservationFee int64      `json:"reservation_fee"`
		Notes          *string    `json:"notes"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.TherapistID == "" || body.PatientID == "" {
		return badRequest(c, "therapist_id and patient_id are required")
	}

	therapistID, err := uuid.Parse(body.TherapistID)
	if err != nil {
		return badRequest(c, "invalid therapist_id")
	}
	patientID, err := uuid.Parse(body.PatientID)
	if err != nil {
		return badRequest(c, "invalid patient_id")
	}

	req := appointment.BookRequest{
		TherapistID:    therapistID,
		PatientID:      patientID,
		StartTime:      body.StartTime,
		EndTime:        body.EndTime,
		SessionPrice:   body.SessionPrice,
		ReservationFee: body.ReservationFee,
		Notes:          body.Notes,
	}
	if body.TimeSlotID != nil {
		id, err := uuid.Parse(*body.TimeSlotID)
		if err != nil {
			return badRequest(c, "invalid time_slot_id")
		}
		req.TimeSlotID = &id
	}

	appt, err := h.svc.Book(c.Context(), clinicID, req)
	if err != nil {
		return mapAppointmentError(c, err)
	}

	return created(c, appt)
}

// PATCH /appointments/:id/cancel
func (h *AppointmentHandler) Cancel(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	apptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid appointment id")
	}

	var body struct {
		Reason          *string `json:"reason"`
		RequestedBy     string  `json:"requested_by"`
		CancellationFee int64   `json:"cancellation_fee"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.RequestedBy == "" {
		body.RequestedBy = "clinic"
	}

	if err := h.svc.Cancel(c.Context(), clinicID, apptID, appointment.CancelRequest{
		Reason:          body.Reason,
		RequestedBy:     body.RequestedBy,
		CancellationFee: body.CancellationFee,
	}); err != nil {
		return mapAppointmentError(c, err)
	}

	return noContent(c)
}

// PATCH /appointments/:id/complete
func (h *AppointmentHandler) Complete(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	apptID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid appointment id")
	}

	if err := h.svc.Complete(c.Context(), clinicID, memberID, apptID); err != nil {
		return mapAppointmentError(c, err)
	}

	return noContent(c)
}
