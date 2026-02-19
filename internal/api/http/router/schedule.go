package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerScheduleRoutes(
	api fiber.Router,
	sh *handler.ScheduleHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	// Public: list available slots for a therapist (no auth required)
	api.Get("/therapists/:mid/slots", sh.ListPublicSlots)

	// Authenticated + clinic-scoped routes
	schedule := api.Group("/schedule", authRequired, clinicHeader)

	schedule.Patch("/toggle", requirePerm(authorize.ResourceTimeSlot, authorize.ActionUpdate), sh.Toggle)

	schedule.Get("/slots", requirePerm(authorize.ResourceTimeSlot, authorize.ActionRead), sh.ListSlots)
	schedule.Post("/slots", requirePerm(authorize.ResourceTimeSlot, authorize.ActionCreate), sh.CreateSlot)
	schedule.Delete("/slots/:id", requirePerm(authorize.ResourceTimeSlot, authorize.ActionDelete), sh.DeleteSlot)

	schedule.Get("/recurring", requirePerm(authorize.ResourceRecurringRule, authorize.ActionRead), sh.ListRecurring)
	schedule.Post("/recurring", requirePerm(authorize.ResourceRecurringRule, authorize.ActionCreate), sh.CreateRecurring)
	schedule.Delete("/recurring/:id", requirePerm(authorize.ResourceRecurringRule, authorize.ActionDelete), sh.DeleteRecurring)
}
