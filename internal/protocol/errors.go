package protocol

// ErrorCode is a machine-readable protocol error.
type ErrorCode string

const (
	ErrRoomNotFound      ErrorCode = "ROOM_NOT_FOUND"
	ErrRoomFull          ErrorCode = "ROOM_FULL"
	ErrSlugTaken         ErrorCode = "SLUG_TAKEN"
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrInvalidSource     ErrorCode = "INVALID_SOURCE"
	ErrSourceUnavailable ErrorCode = "SOURCE_UNAVAILABLE"
	ErrRateLimited       ErrorCode = "RATE_LIMITED"
	ErrInvalidNickname   ErrorCode = "INVALID_NICKNAME"
	ErrInvalidSlug       ErrorCode = "INVALID_SLUG"
	ErrInvalidMessage    ErrorCode = "INVALID_MESSAGE"
	ErrNotInRoom         ErrorCode = "NOT_IN_ROOM"
	ErrNoTrackPlaying    ErrorCode = "NO_TRACK_PLAYING"
	ErrAuthFailed        ErrorCode = "AUTH_FAILED"
	ErrAuthRequired      ErrorCode = "AUTH_REQUIRED"
)

// KnownErrorCodes lists codes clients may handle explicitly.
var KnownErrorCodes = []ErrorCode{
	ErrRoomNotFound,
	ErrRoomFull,
	ErrSlugTaken,
	ErrForbidden,
	ErrInvalidSource,
	ErrSourceUnavailable,
	ErrRateLimited,
	ErrInvalidNickname,
	ErrInvalidSlug,
	ErrInvalidMessage,
	ErrNotInRoom,
	ErrNoTrackPlaying,
	ErrAuthFailed,
	ErrAuthRequired,
}

// IsKnownErrorCode reports whether code is a defined protocol error.
func IsKnownErrorCode(code ErrorCode) bool {
	for _, c := range KnownErrorCodes {
		if c == code {
			return true
		}
	}
	return false
}
