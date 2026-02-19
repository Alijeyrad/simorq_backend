package notification

import "errors"

var (
	ErrNotFound     = errors.New("notification not found")
	ErrUnauthorized = errors.New("not authorized to access this notification")
	ErrDeviceExists = errors.New("device token already registered")
)
