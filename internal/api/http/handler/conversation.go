package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/conversation"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type ConversationHandler struct {
	svc conversation.Service
}

func NewConversationHandler(svc conversation.Service) *ConversationHandler {
	return &ConversationHandler{svc: svc}
}

func mapConversationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, conversation.ErrNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, conversation.ErrUnauthorized):
		return forbidden(c)
	case errors.Is(err, conversation.ErrAlreadyExists):
		return conflict(c, err.Error())
	case errors.Is(err, conversation.ErrMessageNotFound):
		return notFound(c, err.Error())
	default:
		return internalError(c)
	}
}

// GET /conversations
func (h *ConversationHandler) List(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	userID, valid := userIDFromClaims(c)
	if !valid {
		return unauthorized(c)
	}

	var q struct {
		Page    int `query:"page"`
		PerPage int `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	convs, err := h.svc.List(c.Context(), clinicID, userID, q.Page, q.PerPage)
	if err != nil {
		return mapConversationError(c, err)
	}

	return ok(c, convs)
}

// POST /conversations
func (h *ConversationHandler) Create(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var body struct {
		ParticipantB string  `json:"participant_b"`
		PatientID    *string `json:"patient_id"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	participantB, err := uuid.Parse(body.ParticipantB)
	if err != nil {
		return badRequest(c, "invalid participant_b")
	}

	req := conversation.CreateRequest{
		ParticipantA: claims.UserID,
		ParticipantB: participantB,
	}
	if body.PatientID != nil {
		pid, err := uuid.Parse(*body.PatientID)
		if err != nil {
			return badRequest(c, "invalid patient_id")
		}
		req.PatientID = &pid
	}

	conv, err := h.svc.Create(c.Context(), clinicID, req)
	if err != nil {
		return mapConversationError(c, err)
	}

	return created(c, conv)
}

// GET /conversations/:id
func (h *ConversationHandler) Get(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	userID, valid := userIDFromClaims(c)
	if !valid {
		return unauthorized(c)
	}

	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid conversation id")
	}

	conv, err := h.svc.GetByID(c.Context(), clinicID, convID, userID)
	if err != nil {
		return mapConversationError(c, err)
	}

	return ok(c, conv)
}

// GET /conversations/:id/messages
func (h *ConversationHandler) ListMessages(c fiber.Ctx) error {
	userID, valid := userIDFromClaims(c)
	if !valid {
		return unauthorized(c)
	}

	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid conversation id")
	}

	var q struct {
		Page    int `query:"page"`
		PerPage int `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	msgs, err := h.svc.ListMessages(c.Context(), convID, userID, conversation.ListMessagesRequest{
		Page:    q.Page,
		PerPage: q.PerPage,
	})
	if err != nil {
		return mapConversationError(c, err)
	}

	return ok(c, msgs)
}

// POST /conversations/:id/messages
func (h *ConversationHandler) SendMessage(c fiber.Ctx) error {
	userID, valid := userIDFromClaims(c)
	if !valid {
		return unauthorized(c)
	}

	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid conversation id")
	}

	var body struct {
		Content  *string `json:"content"`
		FileKey  *string `json:"file_key"`
		FileName *string `json:"file_name"`
		FileMime *string `json:"file_mime"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	if body.Content == nil && body.FileKey == nil {
		return badRequest(c, "content or file_key is required")
	}

	msg, err := h.svc.SendMessage(c.Context(), convID, conversation.SendMessageRequest{
		SenderID: userID,
		Content:  body.Content,
		FileKey:  body.FileKey,
		FileName: body.FileName,
		FileMime: body.FileMime,
	})
	if err != nil {
		return mapConversationError(c, err)
	}

	return created(c, msg)
}

// DELETE /conversations/:id/messages/:msg_id
func (h *ConversationHandler) DeleteMessage(c fiber.Ctx) error {
	userID, valid := userIDFromClaims(c)
	if !valid {
		return unauthorized(c)
	}

	convID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid conversation id")
	}

	msgID, err := uuid.Parse(c.Params("msg_id"))
	if err != nil {
		return badRequest(c, "invalid message id")
	}

	if err := h.svc.DeleteMessage(c.Context(), convID, msgID, userID); err != nil {
		return mapConversationError(c, err)
	}

	return noContent(c)
}
