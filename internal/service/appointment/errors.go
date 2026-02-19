package appointment

import "errors"

var (
	ErrNotFound         = errors.New("appointment not found")
	ErrSlotNotAvailable = errors.New("time slot is not available for booking")
	ErrAlreadyCompleted = errors.New("appointment is already completed")
	ErrAlreadyCancelled = errors.New("appointment is already cancelled")
)
