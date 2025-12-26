package questions

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("question not found")
	ErrInvalidType  = errors.New("invalid question type")
)
