package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/notification"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type NotificationHandler struct {
	svc notification.Service
}

func NewNotificationHandler(svc notification.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func mapNotificationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, notification.ErrNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, notification.ErrUnauthorized):
		return forbidden(c)
	default:
		return internalError(c)
	}
}

// GET /notifications
func (h *NotificationHandler) List(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	var q struct {
		UnreadOnly bool `query:"unread_only"`
		Page       int  `query:"page"`
		PerPage    int  `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	notifs, err := h.svc.List(c.Context(), claims.UserID, q.UnreadOnly, q.Page, q.PerPage)
	if err != nil {
		return mapNotificationError(c, err)
	}

	return ok(c, notifs)
}

// PATCH /notifications/:id/read
func (h *NotificationHandler) MarkRead(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "invalid notification id")
	}

	if err := h.svc.MarkRead(c.Context(), notifID, claims.UserID); err != nil {
		return mapNotificationError(c, err)
	}

	return noContent(c)
}

// PATCH /notifications/read-all
func (h *NotificationHandler) MarkAllRead(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	if err := h.svc.MarkAllRead(c.Context(), claims.UserID); err != nil {
		return mapNotificationError(c, err)
	}

	return noContent(c)
}

// GET /users/me/notification-prefs
func (h *NotificationHandler) GetPrefs(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	prefs, err := h.svc.GetPrefs(c.Context(), claims.UserID)
	if err != nil {
		return mapNotificationError(c, err)
	}

	return ok(c, prefs)
}

// PUT /users/me/notification-prefs
func (h *NotificationHandler) UpdatePrefs(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	var body struct {
		AppointmentSMS  bool `json:"appointment_sms"`
		AppointmentPush bool `json:"appointment_push"`
		MessagePush     bool `json:"message_push"`
		TicketReplyPush bool `json:"ticket_reply_push"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	prefs, err := h.svc.UpsertPrefs(c.Context(), claims.UserID, notification.UpsertPrefsRequest{
		AppointmentSMS:  body.AppointmentSMS,
		AppointmentPush: body.AppointmentPush,
		MessagePush:     body.MessagePush,
		TicketReplyPush: body.TicketReplyPush,
	})
	if err != nil {
		return mapNotificationError(c, err)
	}

	return ok(c, prefs)
}

// POST /notifications/register-device
func (h *NotificationHandler) RegisterDevice(c fiber.Ctx) error {
	claims, claimsOK := pasetotoken.ClaimsFromFiber(c)
	if !claimsOK {
		return unauthorized(c)
	}

	var body struct {
		DeviceToken string `json:"device_token"`
		Platform    string `json:"platform"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.DeviceToken == "" || body.Platform == "" {
		return badRequest(c, "device_token and platform are required")
	}

	device, err := h.svc.RegisterDevice(c.Context(), notification.RegisterDeviceRequest{
		UserID:      claims.UserID,
		DeviceToken: body.DeviceToken,
		Platform:    body.Platform,
	})
	if err != nil {
		return mapNotificationError(c, err)
	}

	return created(c, device)
}
