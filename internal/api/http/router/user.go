package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerUserRoutes(api fiber.Router, h *handler.UserHandler, authRequired fiber.Handler) {
	users := api.Group("/users", authRequired)
	users.Get("/me", h.GetMe)
	users.Patch("/me", h.UpdateMe)
}
