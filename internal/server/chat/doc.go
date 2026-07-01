package chat

import "errors"

var (
	// ErrEmptyMessage is returned when chat body is blank after trim (AC-036).
	ErrEmptyMessage = errors.New("message must not be empty")

	// ErrBodyTooLong is returned when chat body exceeds the limit.
	ErrBodyTooLong = errors.New("message too long")
)

const (
	defaultCapacity = 100
	maxBodyRunes    = 500
)
