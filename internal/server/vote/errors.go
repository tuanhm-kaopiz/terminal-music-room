package vote

import "errors"

var (
	ErrNoTrackPlaying   = errors.New("no track playing")
	ErrQueueTooShort    = errors.New("queue too short for priority vote")
	ErrInvalidTarget    = errors.New("queue item not found")
	ErrVoteKindConflict = errors.New("another vote is in progress")
)

const defaultTimeoutSeconds = 60
