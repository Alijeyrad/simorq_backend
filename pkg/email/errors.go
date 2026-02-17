package email

import "fmt"

type ErrDisabled struct{}

func (e ErrDisabled) Error() string { return "email is disabled" }

type ErrInvalidMessage struct{ Reason string }

func (e ErrInvalidMessage) Error() string { return "invalid email message: " + e.Reason }

type ErrSend struct {
	Provider string
	Err      error
}

func (e ErrSend) Error() string { return fmt.Sprintf("email send failed (%s): %v", e.Provider, e.Err) }
func (e ErrSend) Unwrap() error { return e.Err }
