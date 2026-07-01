package youtube

import "errors"

var (
	// ErrInvalidSource is returned for malformed or non-YouTube inputs (AC-019).
	ErrInvalidSource = errors.New("invalid source")

	// ErrSourceUnavailable is returned when yt-dlp cannot resolve or search (AC-020).
	ErrSourceUnavailable = errors.New("source unavailable")
)
