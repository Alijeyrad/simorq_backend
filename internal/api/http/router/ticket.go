package router

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
)

func (r *Router) registerTicketRoutes(
	api fiber.Router,
	th *handler.TicketHandler,
	authRequired fiber.Handler,
) {
	tickets := api.Group("/tickets", authRequired)

	tickets.Get("/", th.List)
	tickets.Post("/", th.Create)

	t := tickets.Group("/:id")
	t.Get("/", th.Get)
	t.Patch("/status", th.UpdateStatus)
	t.Get("/messages", th.ListMessages)
	t.Post("/messages", th.Reply)
}
