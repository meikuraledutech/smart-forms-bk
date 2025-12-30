package analytics

import "errors"

var (
	ErrFormNotFound        = errors.New("form not found")
	ErrNoResponses         = errors.New("no responses found for this form")
	ErrCalculationFailed   = errors.New("analytics calculation failed")
	ErrCalculationPending  = errors.New("analytics calculation is pending")
	ErrInvalidInput        = errors.New("invalid input")
)
