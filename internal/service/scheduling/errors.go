package scheduling

import "errors"

var (
	ErrSlotNotFound      = errors.New("time slot not found")
	ErrSlotAlreadyBooked = errors.New("time slot is already booked")
	ErrOverlappingSlot   = errors.New("time slot overlaps with an existing slot")
	ErrInvalidTimeRange  = errors.New("end_time must be after start_time")
	ErrRuleNotFound      = errors.New("recurring rule not found")
)
