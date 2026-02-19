package ticket

import "errors"

var (
	ErrNotFound      = errors.New("ticket not found")
	ErrUnauthorized  = errors.New("not authorized to access this ticket")
	ErrAlreadyClosed = errors.New("ticket is already closed")
)
