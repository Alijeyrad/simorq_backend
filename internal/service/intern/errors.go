package intern

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrAccessAlreadyGranted = errors.New("access already granted")
)
