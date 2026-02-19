package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerAppointmentRoutes(
	api fiber.Router,
	ah *handler.AppointmentHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	appts := api.Group("/appointments", authRequired, clinicHeader)

	appts.Get("/", requirePerm(authorize.ResourceAppointment, authorize.ActionRead), ah.List)
	appts.Post("/", requirePerm(authorize.ResourceAppointment, authorize.ActionCreate), ah.Book)

	a := appts.Group("/:id")
	a.Get("/", requirePerm(authorize.ResourceAppointment, authorize.ActionRead), ah.GetByID)
	a.Patch("/cancel", requirePerm(authorize.ResourceAppointment, authorize.ActionUpdate), ah.Cancel)
	a.Patch("/complete", requirePerm(authorize.ResourceAppointment, authorize.ActionUpdate), ah.Complete)
}
