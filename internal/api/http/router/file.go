package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerFileRoutes(
	api fiber.Router,
	h *handler.FileHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
) {
	files := api.Group("/files", authRequired, clinicHeader)
	files.Post("/upload", h.Upload)
	files.Get("/:key", h.GetByKey)
}
