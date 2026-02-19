package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/internal/service/contact"
)

type ContactHandler struct {
	svc contact.Service
}

func NewContactHandler(svc contact.Service) *ContactHandler {
	return &ContactHandler{svc: svc}
}

type submitContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (h *ContactHandler) Submit(c fiber.Ctx) error {
	var req submitContactRequest
	if err := c.Bind().JSON(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Name == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		return badRequest(c, "name, email, subject, and message are required")
	}

	msg, err := h.svc.Submit(c.Context(), contact.CreateRequest{
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Message: req.Message,
	})
	if err != nil {
		return internalError(c)
	}
	return created(c, msg)
}
