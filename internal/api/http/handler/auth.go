package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/service/auth"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
)

type AuthHandler struct {
	svc auth.Service
}

func NewAuthHandler(svc auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var body struct {
		Phone      string `json:"phone"`
		Password   string `json:"password"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		NationalID string `json:"national_id"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.svc.Register(c.Context(), auth.RegisterRequest{
		Phone:      body.Phone,
		Password:   body.Password,
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		NationalID: body.NationalID,
	}); err != nil {
		return mapAuthError(c, err)
	}

	return created(c, fiber.Map{"message": "verification code sent to your phone"})
}

// POST /api/v1/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var body struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	tokens, err := h.svc.VerifyOTP(c.Context(), auth.VerifyOTPRequest{
		Phone: body.Phone,
		Code:  body.Code,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return ok(c, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var body struct {
		Phone      string `json:"phone"`
		NationalID string `json:"national_id"`
		Password   string `json:"password"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	tokens, err := h.svc.Login(c.Context(), auth.LoginRequest{
		Phone:      body.Phone,
		NationalID: body.NationalID,
		Password:   body.Password,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return ok(c, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.RefreshToken == "" {
		return badRequest(c, "refresh_token is required")
	}

	tokens, err := h.svc.RefreshTokens(c.Context(), body.RefreshToken)
	if err != nil {
		return mapAuthError(c, err)
	}

	return ok(c, fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

// POST /api/v1/auth/logout  (requires AuthRequired middleware)
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	claims, ok := pasetotoken.ClaimsFromFiber(c)
	if !ok || claims.SessionID == nil {
		return unauthorized(c)
	}

	if err := h.svc.Logout(c.Context(), *claims.SessionID); err != nil {
		return internalError(c)
	}

	return noContent(c)
}

// POST /api/v1/auth/intern-setup  (requires AuthRequired middleware)
func (h *AuthHandler) InternSetup(c fiber.Ctx) error {
	claims, valid := pasetotoken.ClaimsFromFiber(c)
	if !valid {
		return unauthorized(c)
	}

	var body struct {
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		InternshipYear int    `json:"internship_year"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}

	result, err := h.svc.InternSetup(c.Context(), claims.UserID, auth.InternSetupRequest{
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		InternshipYear: body.InternshipYear,
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return ok(c, result)
}

// POST /api/v1/auth/change-password  (requires AuthRequired middleware)
func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	claims, ok := pasetotoken.ClaimsFromFiber(c)
	if !ok || claims.SessionID == nil {
		return unauthorized(c)
	}

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return badRequest(c, "invalid request body")
	}
	if body.CurrentPassword == "" || body.NewPassword == "" {
		return badRequest(c, "current_password and new_password are required")
	}

	if err := h.svc.ChangePassword(c.Context(), claims.UserID, *claims.SessionID, auth.ChangePasswordRequest{
		CurrentPassword: body.CurrentPassword,
		NewPassword:     body.NewPassword,
	}); err != nil {
		return mapAuthError(c, err)
	}

	return noContent(c)
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

func mapAuthError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, auth.ErrPhoneAlreadyExists):
		return conflict(c, err.Error())
	case errors.Is(err, auth.ErrNationalIDExists):
		return conflict(c, err.Error())
	case errors.Is(err, auth.ErrInvalidPhone):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrInvalidNationalID):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrPasswordTooShort):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrOTPExpired):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrOTPInvalid):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrOTPMaxAttempts):
		return tooManyRequests(c, err.Error())
	case errors.Is(err, auth.ErrWrongPassword):
		return badRequest(c, err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, auth.ErrAccountSuspended):
		return forbidden(c)
	case errors.Is(err, auth.ErrPhoneNotVerified):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, auth.ErrAccountLocked):
		return tooManyRequests(c, err.Error())
	case errors.Is(err, auth.ErrSessionNotFound):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, auth.ErrInvalidToken):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	default:
		// Parse UUID as a sanity-check sentinel — if it's a user-not-found from UUID parse, treat it as 401
		if _, parseErr := uuid.Parse(err.Error()); parseErr == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return internalError(c)
	}
}
