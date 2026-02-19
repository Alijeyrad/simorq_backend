package router

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
)

func (r *Router) registerNotificationRoutes(
	api fiber.Router,
	nh *handler.NotificationHandler,
	authRequired fiber.Handler,
) {
	notifs := api.Group("/notifications", authRequired)

	notifs.Get("/", nh.List)
	notifs.Post("/register-device", nh.RegisterDevice)
	notifs.Patch("/read-all", nh.MarkAllRead)
	notifs.Patch("/:id/read", nh.MarkRead)

	// Notification preferences nested under /users/me
	me := api.Group("/users/me", authRequired)
	me.Get("/notification-prefs", nh.GetPrefs)
	me.Put("/notification-prefs", nh.UpdatePrefs)
}
