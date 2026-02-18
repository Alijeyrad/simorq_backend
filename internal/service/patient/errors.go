package patient

import "errors"

var (
	ErrPatientNotFound       = errors.New("patient not found")
	ErrPatientAlreadyExists  = errors.New("user is already a patient in this clinic")
	ErrReportNotFound        = errors.New("patient report not found")
	ErrPrescriptionNotFound  = errors.New("prescription not found")
	ErrPatientTestNotFound   = errors.New("patient test not found")
	ErrInvalidStatus         = errors.New("invalid patient status")
	ErrAccessDenied          = errors.New("access denied to this patient record")
)
