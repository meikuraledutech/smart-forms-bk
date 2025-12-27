package links

import "errors"

var (
	ErrFormNotFound     = errors.New("form not found")
	ErrFormNotPublished = errors.New("form is not published")
	ErrSlugTaken        = errors.New("slug already taken")
	ErrInvalidSlug      = errors.New("invalid slug format")
	ErrInvalidInput     = errors.New("invalid input")
)
