package room

import "errors"

var (
	ErrRoomFull       = errors.New("room is full")
	ErrAlreadyMember  = errors.New("already in room")
	ErrNoTrackPlaying = errors.New("no track playing")
	ErrSlugTaken      = errors.New("slug taken")
	ErrRoomNotFound   = errors.New("room not found")
	ErrNotInRoom      = errors.New("not in room")
	ErrAlreadyInRoom  = errors.New("already in a room")
	ErrQueueItemNotFound = errors.New("queue item not found")
	ErrForbidden      = errors.New("forbidden")
	ErrAuthFailed     = errors.New("authentication failed")
	ErrAuthRequired   = errors.New("password required")
)
