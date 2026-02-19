package conversation

import "errors"

var (
	ErrNotFound        = errors.New("conversation not found")
	ErrUnauthorized    = errors.New("not a participant in this conversation")
	ErrAlreadyExists   = errors.New("conversation already exists between these participants")
	ErrMessageNotFound = errors.New("message not found")
)
