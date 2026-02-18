package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerClinicRoutes(
	api fiber.Router,
	h *handler.ClinicHandler,
	authRequired fiber.Handler,
	clinicCtx fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	clinics := api.Group("/clinics")

	clinics.Get("/", h.List)
	clinics.Get("/:slug", h.GetBySlug)
	clinics.Post("/", authRequired, h.Create)

	mgmt := clinics.Group("/:id", authRequired, clinicCtx)
	mgmt.Patch("/", requirePerm(authorize.ResourceClinic, authorize.ActionUpdate), h.Update)
	mgmt.Get("/settings", h.GetSettings)
	mgmt.Patch("/settings", requirePerm(authorize.ResourceClinicSettings, authorize.ActionUpdate), h.UpdateSettings)
	mgmt.Get("/members", h.ListMembers)
	mgmt.Post("/members", requirePerm(authorize.ResourceClinicMember, authorize.ActionCreate), h.AddMember)
	mgmt.Patch("/members/:mid", requirePerm(authorize.ResourceClinicMember, authorize.ActionUpdate), h.UpdateMember)
	mgmt.Delete("/members/:mid", requirePerm(authorize.ResourceClinicMember, authorize.ActionDelete), h.RemoveMember)
	mgmt.Get("/therapists", h.ListTherapists)
}
