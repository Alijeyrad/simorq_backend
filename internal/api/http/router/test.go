package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerTestRoutes(
	api fiber.Router,
	h *handler.TestHandler,
	authRequired fiber.Handler,
) {
	tests := api.Group("/tests", authRequired)
	tests.Get("/", h.List)
	tests.Get("/:id", h.GetByID)
}
