package router

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
)

func (r *Router) registerConversationRoutes(
	api fiber.Router,
	ch *handler.ConversationHandler,
	authRequired fiber.Handler,
	clinicHeader fiber.Handler,
	requirePerm func(authorize.Resource, authorize.Action) fiber.Handler,
) {
	convs := api.Group("/conversations", authRequired, clinicHeader)

	convs.Get("/", requirePerm(authorize.ResourceConversation, authorize.ActionRead), ch.List)
	convs.Post("/", requirePerm(authorize.ResourceConversation, authorize.ActionCreate), ch.Create)

	c := convs.Group("/:id")
	c.Get("/", requirePerm(authorize.ResourceConversation, authorize.ActionRead), ch.Get)
	c.Get("/messages", requirePerm(authorize.ResourceMessage, authorize.ActionRead), ch.ListMessages)
	c.Post("/messages", requirePerm(authorize.ResourceMessage, authorize.ActionCreate), ch.SendMessage)
	c.Delete("/messages/:msg_id", requirePerm(authorize.ResourceMessage, authorize.ActionDelete), ch.DeleteMessage)
}
