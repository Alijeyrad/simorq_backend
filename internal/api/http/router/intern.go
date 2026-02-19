package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerInternRoutes(
	api fiber.Router,
	h *handler.InternHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	// --- Intern self-management (requires auth + clinic context) ---
	internSelf := api.Group("/intern",
		authRequired,
		clinicHeader,
	)
	internSelf.Get("/profile", h.GetMyProfile)
	internSelf.Put("/profile", h.UpsertMyProfile)
	internSelf.Get("/tasks", h.ListMyTasks)
	internSelf.Post("/tasks", h.CreateTask)
	internSelf.Get("/tasks/:taskID", h.GetTask)
	internSelf.Patch("/tasks/:taskID", h.UpdateTask)
	internSelf.Post("/tasks/:taskID/files", h.AddTaskFile)
	internSelf.Get("/patients", h.ListMyPatients)

	// --- Admin / supervisor management ---
	admin := api.Group("/clinics/interns",
		authRequired,
		clinicHeader,
		requirePerm(authorize.ResourceInternTask, authorize.ActionManage),
	)
	admin.Get("/", h.ListInterns)
	admin.Get("/:internID/tasks", h.ListInternTasks)
	admin.Patch("/:internID/tasks/:taskID/review", h.ReviewTask)
	admin.Post("/:internID/patients/:patientID/access", h.GrantAccess)
	admin.Delete("/:internID/patients/:patientID/access", h.RevokeAccess)
}
