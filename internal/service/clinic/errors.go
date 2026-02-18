package clinic

import "errors"

var (
	ErrClinicNotFound    = errors.New("clinic not found")
	ErrSlugAlreadyExists = errors.New("clinic slug already taken")
	ErrMemberNotFound    = errors.New("clinic member not found")
	ErrAlreadyMember     = errors.New("user is already a member of this clinic")
	ErrInvalidRole       = errors.New("invalid clinic member role")
	ErrCannotRemoveOwner = errors.New("cannot remove the clinic owner")
	ErrNotMember         = errors.New("user is not a member of this clinic")
)
