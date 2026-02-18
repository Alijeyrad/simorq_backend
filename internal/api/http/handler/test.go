package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/psychtest"
)

type TestHandler struct {
	svc psychtest.Service
}

func NewTestHandler(svc psychtest.Service) *TestHandler {
	return &TestHandler{svc: svc}
}

// GET /tests
func (h *TestHandler) List(c fiber.Ctx) error {
	tests, err := h.svc.List(c.Context())
	if err != nil {
		return internalError(c)
	}
	return ok(c, tests)
}

// GET /tests/:id
func (h *TestHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid test id")
	}

	t, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, psychtest.ErrNotFound) {
			return notFound(c, err.Error())
		}
		return internalError(c)
	}

	return ok(c, t)
}
