package auth

import "errors"

var (
	ErrPhoneAlreadyExists  = errors.New("phone number already registered")
	ErrNationalIDExists    = errors.New("national ID already registered")
	ErrInvalidPhone        = errors.New("invalid phone number format")
	ErrInvalidNationalID   = errors.New("national ID must be exactly 10 digits")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters")
	ErrOTPExpired          = errors.New("OTP has expired or does not exist")
	ErrOTPInvalid          = errors.New("OTP code is incorrect")
	ErrOTPMaxAttempts      = errors.New("too many incorrect OTP attempts")
	ErrInvalidCredentials  = errors.New("phone/national ID or password is incorrect")
	ErrAccountSuspended    = errors.New("account is suspended")
	ErrPhoneNotVerified    = errors.New("phone number is not verified")
	ErrAccountLocked       = errors.New("account temporarily locked due to repeated login failures")
	ErrSessionNotFound     = errors.New("session not found or expired")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrNotIntern           = errors.New("intern setup requires an intern role in a clinic")
	ErrWrongPassword       = errors.New("current password is incorrect")
)
