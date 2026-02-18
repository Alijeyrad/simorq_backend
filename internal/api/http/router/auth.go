package router

import (
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) registerAuthRoutes(api fiber.Router, h *handler.AuthHandler, authRequired fiber.Handler) {
	group := api.Group("/auth")
	group.Post("/register", h.Register)
	group.Post("/verify-otp", h.VerifyOTP)
	group.Post("/login", h.Login)
	group.Post("/refresh", h.Refresh)
	group.Post("/logout", authRequired, h.Logout)
	group.Post("/intern-setup", authRequired, h.InternSetup)
}
