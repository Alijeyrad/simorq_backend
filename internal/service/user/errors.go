package user

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPassword     = errors.New("current password is incorrect")
	ErrPasswordRequired    = errors.New("password already exists, use change password instead")
	ErrPhoneNotVerified    = errors.New("phone number is not verified")
	ErrEmailNotVerified    = errors.New("email address is not verified")
	ErrInvalidProfileField = errors.New("invalid profile field")
	ErrInvalidDisplayName  = errors.New("display name must be between 1 and 100 characters")
	ErrInvalidBio          = errors.New("bio must be 500 characters or less")
	ErrInvalidURL          = errors.New("avatar URL must be a valid URL")
	ErrInvalidLocale       = errors.New("invalid locale code")
	ErrInvalidTimezone     = errors.New("invalid timezone")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrInvalidPhone        = errors.New("invalid phone number for the specified region")
	ErrPhoneAlreadyExists  = errors.New("phone number is already in use")
	ErrEmailAlreadyExists  = errors.New("email address is already in use")
)
