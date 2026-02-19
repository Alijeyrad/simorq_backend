package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerContactRoutes(api fiber.Router, h *handler.ContactHandler) {
	api.Post("/contact", h.Submit)
}
