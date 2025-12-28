package responses

import "errors"

var (
	ErrFormNotFound        = errors.New("form not found")
	ErrFormNotPublished    = errors.New("form is not published")
	ErrFormNotAccepting    = errors.New("form is not accepting responses")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidFlowConnection = errors.New("invalid flow connection id")
)
