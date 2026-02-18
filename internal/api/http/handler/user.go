package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type UserHandler struct {
	svc user.Service
}

func NewUserHandler(svc user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// GET /api/v1/users/me
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	claims, valid := pasetotoken.ClaimsFromFiber(c)
	if !valid {
		return unauthorized(c)
	}

	u, err := h.svc.GetMe(c.Context(), claims.UserID)
	if err != nil {
		return internalError(c)
	}

	return ok(c, u)
}

// PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(c fiber.Ctx) error {
	claims, valid := pasetotoken.ClaimsFromFiber(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		FirstName     *string `json:"first_name"`
		LastName      *string `json:"last_name"`
		Gender        *string `json:"gender"`
		MaritalStatus *string `json:"marital_status"`
		BirthYear     *int    `json:"birth_year"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	result, err := h.svc.UpdateMe(c.Context(), claims.UserID, user.UpdateMeRequest{
		FirstName:     body.FirstName,
		LastName:      body.LastName,
		Gender:        body.Gender,
		MaritalStatus: body.MaritalStatus,
		BirthYear:     body.BirthYear,
	})
	if err != nil {
		return internalError(c)
	}

	return ok(c, result)
}
