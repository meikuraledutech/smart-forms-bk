package flows

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("flow not found")
	ErrFormNotFound = errors.New("form not found")
)
