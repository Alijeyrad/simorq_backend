package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerPatientRoutes(
	api fiber.Router,
	ph *handler.PatientHandler,
	fh *handler.FileHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	patients := api.Group("/patients", authRequired, clinicHeader)

	// Patient CRUD
	patients.Get("/", requirePerm(authorize.ResourcePatient, authorize.ActionRead), ph.List)
	patients.Post("/", requirePerm(authorize.ResourcePatient, authorize.ActionCreate), ph.Create)

	p := patients.Group("/:id")
	p.Get("/", requirePerm(authorize.ResourcePatient, authorize.ActionRead), ph.Get)
	p.Patch("/", requirePerm(authorize.ResourcePatient, authorize.ActionUpdate), ph.Update)

	// Reports
	p.Get("/reports", requirePerm(authorize.ResourcePatientReport, authorize.ActionRead), ph.ListReports)
	p.Post("/reports", requirePerm(authorize.ResourcePatientReport, authorize.ActionCreate), ph.CreateReport)
	p.Patch("/reports/:rid", requirePerm(authorize.ResourcePatientReport, authorize.ActionUpdate), ph.UpdateReport)
	p.Delete("/reports/:rid", requirePerm(authorize.ResourcePatientReport, authorize.ActionDelete), ph.DeleteReport)

	// Files
	p.Get("/files", requirePerm(authorize.ResourcePatientFile, authorize.ActionRead), fh.ListPatientFiles)
	p.Post("/files", requirePerm(authorize.ResourcePatientFile, authorize.ActionCreate), fh.UploadPatientFile)
	p.Get("/files/:fid/download", requirePerm(authorize.ResourcePatientFile, authorize.ActionRead), fh.DownloadPatientFile)
	p.Delete("/files/:fid", requirePerm(authorize.ResourcePatientFile, authorize.ActionDelete), fh.DeletePatientFile)

	// Prescriptions
	p.Get("/prescriptions", requirePerm(authorize.ResourcePatientPrescription, authorize.ActionRead), ph.ListPrescriptions)
	p.Post("/prescriptions", requirePerm(authorize.ResourcePatientPrescription, authorize.ActionCreate), ph.CreatePrescription)
	p.Patch("/prescriptions/:pid", requirePerm(authorize.ResourcePatientPrescription, authorize.ActionUpdate), ph.UpdatePrescription)

	// Tests
	p.Get("/tests", requirePerm(authorize.ResourcePatientTest, authorize.ActionRead), ph.ListTests)
	p.Post("/tests", requirePerm(authorize.ResourcePatientTest, authorize.ActionCreate), ph.CreateTest)
	p.Patch("/tests/:tid", requirePerm(authorize.ResourcePatientTest, authorize.ActionUpdate), ph.UpdateTest)
}
