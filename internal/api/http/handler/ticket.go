package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/ticket"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type TicketHandler struct {
	svc ticket.Service
}

func NewTicketHandler(svc ticket.Service) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func mapTicketError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ticket.ErrNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, ticket.ErrUnauthorized):
		return forbidden(c)
	case errors.Is(err, ticket.ErrAlreadyClosed):
		return conflict(c, err.Error())
	default:
		return internalError(c)
	}
}

// GET /tickets
func (h *TicketHandler) List(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	var q struct {
		Status  string `query:"status"`
		Page    int    `query:"page"`
		PerPage int    `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	req := ticket.ListRequest{
		Page:    q.Page,
		PerPage: q.PerPage,
	}
	if q.Status != "" {
		req.Status = &q.Status
	}

	tickets, err := h.svc.List(c.Context(), claims.UserID, req)
	if err != nil {
		return mapTicketError(c, err)
	}

	return ok(c, tickets)
}

// POST /tickets
func (h *TicketHandler) Create(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	var body struct {
		Subject  string  `json:"subject"`
		Priority string  `json:"priority"`
		Content  string  `json:"content"`
		ClinicID *string `json:"clinic_id"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Subject == "" || body.Content == "" {
		return badRequest(c, "subject and content are required")
	}

	req := ticket.CreateRequest{
		UserID:   claims.UserID,
		Subject:  body.Subject,
		Priority: body.Priority,
		Content:  body.Content,
	}
	if body.ClinicID != nil {
		cid, err := uuid.Parse(*body.ClinicID)
		if err != nil {
			return badRequest(c, "invalid clinic_id")
		}
		req.ClinicID = &cid
	}

	t, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return mapTicketError(c, err)
	}

	return created(c, t)
}

// GET /tickets/:id
func (h *TicketHandler) Get(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid ticket id")
	}

	t, err := h.svc.GetByID(c.Context(), ticketID, claims.UserID)
	if err != nil {
		return mapTicketError(c, err)
	}

	return ok(c, t)
}

// PATCH /tickets/:id/status
func (h *TicketHandler) UpdateStatus(c fiber.Ctx) error {
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid ticket id")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Status == "" {
		return badRequest(c, "status is required")
	}

	if err := h.svc.UpdateStatus(c.Context(), ticketID, ticket.UpdateStatusRequest{
		Status: body.Status,
	}); err != nil {
		return mapTicketError(c, err)
	}

	return noContent(c)
}

// GET /tickets/:id/messages
func (h *TicketHandler) ListMessages(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid ticket id")
	}

	var q struct {
		Page    int `query:"page"`
		PerPage int `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	msgs, err := h.svc.ListMessages(c.Context(), ticketID, claims.UserID, q.Page, q.PerPage)
	if err != nil {
		return mapTicketError(c, err)
	}

	return ok(c, msgs)
}

// POST /tickets/:id/messages
func (h *TicketHandler) Reply(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid ticket id")
	}

	var body struct {
		Content  string  `json:"content"`
		FileKey  *string `json:"file_key"`
		FileName *string `json:"file_name"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Content == "" {
		return badRequest(c, "content is required")
	}

	msg, err := h.svc.Reply(c.Context(), ticketID, ticket.ReplyRequest{
		SenderID: claims.UserID,
		Content:  body.Content,
		FileKey:  body.FileKey,
		FileName: body.FileName,
	})
	if err != nil {
		return mapTicketError(c, err)
	}

	return created(c, msg)
}
