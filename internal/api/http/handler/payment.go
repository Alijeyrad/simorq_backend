package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/payment"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type PaymentHandler struct {
	svc payment.Service
}

func NewPaymentHandler(svc payment.Service) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func userIDFromClaims(c fiber.Ctx) (uuid.UUID, bool) {
	claims, found := pasetotoken.ClaimsFromFiber(c)
	if !found {
		return uuid.UUID{}, false
	}
	return claims.UserID, true
}

func mapPaymentError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, payment.ErrPaymentNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, payment.ErrPaymentFailed):
		return badRequest(c, err.Error())
	case errors.Is(err, payment.ErrZarinPalFailure):
		return internalError(c)
	case errors.Is(err, payment.ErrAmountMismatch):
		return badRequest(c, err.Error())
	case errors.Is(err, payment.ErrWalletNotFound):
		return notFound(c, err.Error())
	case errors.Is(err, payment.ErrInsufficientFunds):
		return badRequest(c, err.Error())
	default:
		return internalError(c)
	}
}

// ---------------------------------------------------------------------------
// ZarinPal
// ---------------------------------------------------------------------------

// POST /payments/pay
func (h *PaymentHandler) Initiate(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	userID, found := userIDFromClaims(c)
	if !found {
		return unauthorized(c)
	}

	var body struct {
		AppointmentID *string `json:"appointment_id"`
		Amount        int64   `json:"amount"`
		Description   string  `json:"description"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Amount <= 0 {
		return badRequest(c, "amount must be positive")
	}
	if body.Description == "" {
		return badRequest(c, "description is required")
	}

	var apptID *uuid.UUID
	if body.AppointmentID != nil {
		id, err := uuid.Parse(*body.AppointmentID)
		if err != nil {
			return badRequest(c, "invalid appointment_id")
		}
		apptID = &id
	}

	payURL, err := h.svc.InitiatePayment(c.Context(), clinicID, userID, apptID, body.Amount, body.Description)
	if err != nil {
		return mapPaymentError(c, err)
	}

	return ok(c, fiber.Map{"pay_url": payURL})
}

// GET /payments/verify
// Public callback from ZarinPal: ?Authority=...&Status=OK|NOK
func (h *PaymentHandler) Verify(c fiber.Ctx) error {
	authority := c.Query("Authority")
	status := c.Query("Status")

	if authority == "" {
		return badRequest(c, "missing Authority parameter")
	}

	pr, err := h.svc.VerifyPayment(c.Context(), authority, status)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentFailed) {
			return c.Redirect().To("/payments/result?status=failed")
		}
		return mapPaymentError(c, err)
	}

	refID := ""
	if pr.ZarinpalRefID != nil {
		refID = *pr.ZarinpalRefID
	}

	return c.Redirect().To("/payments/result?status=success&ref=" + refID)
}

// ---------------------------------------------------------------------------
// Wallet
// ---------------------------------------------------------------------------

// GET /payments/wallet
func (h *PaymentHandler) GetWallet(c fiber.Ctx) error {
	userID, found := userIDFromClaims(c)
	if !found {
		return unauthorized(c)
	}

	w, err := h.svc.GetOrCreateWallet(c.Context(), "user", userID)
	if err != nil {
		return mapPaymentError(c, err)
	}

	return ok(c, w)
}

// GET /payments/transactions
func (h *PaymentHandler) GetTransactions(c fiber.Ctx) error {
	userID, found := userIDFromClaims(c)
	if !found {
		return unauthorized(c)
	}

	w, err := h.svc.GetOrCreateWallet(c.Context(), "user", userID)
	if err != nil {
		return mapPaymentError(c, err)
	}

	var q struct {
		Page    int `query:"page"`
		PerPage int `query:"per_page"`
	}
	_ = c.Bind().Query(&q)

	txns, err := h.svc.GetTransactions(c.Context(), w.ID, q.Page, q.PerPage)
	if err != nil {
		return mapPaymentError(c, err)
	}

	return ok(c, txns)
}

// POST /payments/wallet/iban
func (h *PaymentHandler) SetIBAN(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var body struct {
		IBAN          string `json:"iban"`
		AccountHolder string `json:"account_holder"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.IBAN == "" {
		return badRequest(c, "iban is required")
	}
	if body.AccountHolder == "" {
		return badRequest(c, "account_holder is required")
	}

	w, err := h.svc.GetOrCreateWallet(c.Context(), "clinic", clinicID)
	if err != nil {
		return mapPaymentError(c, err)
	}

	if err := h.svc.SetIBAN(c.Context(), w.ID, body.IBAN, body.AccountHolder); err != nil {
		return mapPaymentError(c, err)
	}

	return ok(c, fiber.Map{"message": "IBAN updated successfully"})
}

// POST /payments/withdraw
func (h *PaymentHandler) Withdraw(c fiber.Ctx) error {
	clinicID, valid := clinicIDFromLocals(c)
	if !valid {
		return badRequest(c, "missing clinic context")
	}

	var body struct {
		Amount int64 `json:"amount"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.Amount <= 0 {
		return badRequest(c, "amount must be positive")
	}

	w, err := h.svc.GetOrCreateWallet(c.Context(), "clinic", clinicID)
	if err != nil {
		return mapPaymentError(c, err)
	}

	wr, err := h.svc.RequestWithdrawal(c.Context(), w.ID, clinicID, body.Amount)
	if err != nil {
		return mapPaymentError(c, err)
	}

	return created(c, wr)
}
