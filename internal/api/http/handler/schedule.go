package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/scheduling"
)

type ScheduleHandler struct {
	svc scheduling.Service
}

func NewScheduleHandler(svc scheduling.Service) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

func mapScheduleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, scheduling.ErrSlotNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, scheduling.ErrSlotAlreadyBooked):
		return conflict(c, err.Error())
	case errors.Is(err, scheduling.ErrOverlappingSlot):
		return conflict(c, err.Error())
	case errors.Is(err, scheduling.ErrInvalidTimeRange):
		return badRequest(c, err.Error())
	case errors.Is(err, scheduling.ErrRuleNotFound):
		return notFound(c, err.Error())
	default:
		return internalError(c)
	}
}

// ---------------------------------------------------------------------------
// Schedule toggle
// ---------------------------------------------------------------------------

// PATCH /schedule/toggle
func (h *ScheduleHandler) Toggle(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.svc.ToggleSchedule(c.Context(), clinicID, memberID, body.Enabled); err != nil {
		return mapScheduleError(c, err)
	}

	return ok(c, fiber.Map{"enabled": body.Enabled})
}

// ---------------------------------------------------------------------------
// Slots
// ---------------------------------------------------------------------------

// GET /schedule/slots
func (h *ScheduleHandler) ListSlots(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	var q struct {
		From string `query:"from"`
		To   string `query:"to"`
	}
	_ = c.Bind().Query(&q)

	from := time.Now()
	to := from.AddDate(0, 1, 0)

	if q.From != "" {
		if t, err := time.Parse(time.RFC3339, q.From); err == nil {
			from = t
		}
	}
	if q.To != "" {
		if t, err := time.Parse(time.RFC3339, q.To); err == nil {
			to = t
		}
	}

	slots, err := h.svc.ListSlots(c.Context(), clinicID, memberID, from, to)
	if err != nil {
		return mapScheduleError(c, err)
	}

	return ok(c, slots)
}

// POST /schedule/slots
func (h *ScheduleHandler) CreateSlot(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		StartTime       time.Time `json:"start_time"`
		EndTime         time.Time `json:"end_time"`
		SessionPrice    *int64    `json:"session_price"`
		ReservationFee  *int64    `json:"reservation_fee"`
		IsRecurring     bool      `json:"is_recurring"`
		RecurringRuleID *string   `json:"recurring_rule_id"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.StartTime.IsZero() || body.EndTime.IsZero() {
		return badRequest(c, "start_time and end_time are required")
	}

	req := scheduling.CreateSlotRequest{
		StartTime:      body.StartTime,
		EndTime:        body.EndTime,
		SessionPrice:   body.SessionPrice,
		ReservationFee: body.ReservationFee,
		IsRecurring:    body.IsRecurring,
	}
	if body.RecurringRuleID != nil {
		id, err := uuid.Parse(*body.RecurringRuleID)
		if err != nil {
			return badRequest(c, "invalid recurring_rule_id")
		}
		req.RecurringRuleID = &id
	}

	slot, err := h.svc.CreateSlot(c.Context(), clinicID, memberID, req)
	if err != nil {
		return mapScheduleError(c, err)
	}

	return created(c, slot)
}

// DELETE /schedule/slots/:id
func (h *ScheduleHandler) DeleteSlot(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	slotID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid slot id")
	}

	if err := h.svc.DeleteSlot(c.Context(), clinicID, memberID, slotID); err != nil {
		return mapScheduleError(c, err)
	}

	return noContent(c)
}

// ---------------------------------------------------------------------------
// Recurring rules
// ---------------------------------------------------------------------------

// GET /schedule/recurring
func (h *ScheduleHandler) ListRecurring(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	rules, err := h.svc.ListRecurringRules(c.Context(), clinicID, memberID)
	if err != nil {
		return mapScheduleError(c, err)
	}

	return ok(c, rules)
}

// POST /schedule/recurring
func (h *ScheduleHandler) CreateRecurring(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		DayOfWeek      int8       `json:"day_of_week"`
		StartHour      int8       `json:"start_hour"`
		StartMinute    int8       `json:"start_minute"`
		EndHour        int8       `json:"end_hour"`
		EndMinute      int8       `json:"end_minute"`
		SessionPrice   *int64     `json:"session_price"`
		ReservationFee *int64     `json:"reservation_fee"`
		ValidFrom      time.Time  `json:"valid_from"`
		ValidUntil     *time.Time `json:"valid_until"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.ValidFrom.IsZero() {
		return badRequest(c, "valid_from is required")
	}

	rule, err := h.svc.CreateRecurringRule(c.Context(), clinicID, memberID, scheduling.CreateRecurringRuleRequest{
		DayOfWeek:      body.DayOfWeek,
		StartHour:      body.StartHour,
		StartMinute:    body.StartMinute,
		EndHour:        body.EndHour,
		EndMinute:      body.EndMinute,
		SessionPrice:   body.SessionPrice,
		ReservationFee: body.ReservationFee,
		ValidFrom:      body.ValidFrom,
		ValidUntil:     body.ValidUntil,
	})
	if err != nil {
		return mapScheduleError(c, err)
	}

	return created(c, rule)
}

// DELETE /schedule/recurring/:id
func (h *ScheduleHandler) DeleteRecurring(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}
	memberID, valid := memberIDFromLocals(c)
	if !valid {
		return unauthorized(c)
	}

	ruleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid rule id")
	}

	if err := h.svc.DeleteRecurringRule(c.Context(), clinicID, memberID, ruleID); err != nil {
		return mapScheduleError(c, err)
	}

	return noContent(c)
}

// ---------------------------------------------------------------------------
// Public listing
// ---------------------------------------------------------------------------

// GET /therapists/:mid/slots
func (h *ScheduleHandler) ListPublicSlots(c fiber.Ctx) error {
	therapistMemberID, err := uuid.Parse(c.Params("mid"))
	if err != nil {
		return badRequest(c, "invalid therapist member id")
	}

	var q struct {
		From string `query:"from"`
		To   string `query:"to"`
	}
	_ = c.Bind().Query(&q)

	from := time.Now()
	to := from.AddDate(0, 1, 0)

	if q.From != "" {
		if t, err := time.Parse(time.RFC3339, q.From); err == nil {
			from = t
		}
	}
	if q.To != "" {
		if t, err := time.Parse(time.RFC3339, q.To); err == nil {
			to = t
		}
	}

	slots, err := h.svc.ListPublicSlots(c.Context(), therapistMemberID, from, to)
	if err != nil {
		return internalError(c)
	}

	return ok(c, slots)
}
