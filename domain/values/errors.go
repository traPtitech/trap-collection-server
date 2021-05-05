package values

import "errors"

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrTooLong = errors.New("too long")
)
