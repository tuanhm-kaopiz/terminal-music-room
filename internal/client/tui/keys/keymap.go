package keys

import (
	"errors"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Global shortcut keys (ADR-006).
const (
	KeyQuit        = "q"
	KeyHelp        = "?"
	KeyPauseToggle = " "
	KeySkip        = "s"
	KeyAddSource   = "a"
	KeySeek        = "S"
	KeyVoteSkip    = "v"
	KeyVotePriority = "V"
	KeyQueueRemove  = "d"
	KeyQueueUp      = "ctrl+up"
	KeyQueueDown    = "ctrl+down"
	KeyLeave        = "l"
	KeyKick         = "k"
	KeyKickDel      = "delete"
	KeyTab          = "tab"
	KeyShiftTab     = "shift+tab"
	KeyUp           = "up"
	KeyDown         = "down"
)

// ErrHostOnly is returned when a host-only action is attempted by a member (AC-022).
var ErrHostOnly = errors.New("host only")

// ReactionForKey maps quick-react digit keys to emoji (ADR-006).
func ReactionForKey(key string) (string, bool) {
	switch key {
	case "1":
		return "🔥", true
	case "2":
		return "❤️", true
	case "3":
		return "😂", true
	case "4":
		return "👍", true
	default:
		return "", false
	}
}

// ErrNoTrack is returned when playback actions require a loaded track (AC-016).
var ErrNoTrack = errors.New("nothing playing")

// HasTrack reports whether the room has a loaded playback track.
func HasTrack(v state.View) bool {
	return v.Room.Playback.Track != nil
}

// PlaybackToggleAction is the client-side pause/resume decision for Space.
type PlaybackToggleAction int

const (
	PlaybackPause PlaybackToggleAction = iota
	PlaybackResume
)

// PlaybackToggle decides pause vs resume for the current room playback state.
func PlaybackToggle(v state.View) (PlaybackToggleAction, error) {
	if !HasTrack(v) {
		return 0, ErrNoTrack
	}
	if v.Room.Playback.Status == protocol.PlaybackPaused {
		return PlaybackResume, nil
	}
	return PlaybackPause, nil
}

// RequireTrack guards skip/seek/pause when nothing is loaded (AC-016).
func RequireTrack(v state.View) error {
	if !HasTrack(v) {
		return ErrNoTrack
	}
	return nil
}

// RequireHost guards host-only queue admin actions (AC-022).
func RequireHost(isHost bool) error {
	if !isHost {
		return ErrHostOnly
	}
	return nil
}

// QueueReorderTargets returns item and after IDs for ctrl+↑/↓ reorder.
// direction -1 moves earlier; +1 moves later in the queue.
func QueueReorderTargets(items []protocol.QueueItem, idx, direction int) (itemID, afterID string, ok bool) {
	if idx < 0 || idx >= len(items) || direction == 0 {
		return "", "", false
	}
	itemID = items[idx].ID
	switch direction {
	case -1:
		if idx == 0 {
			return "", "", false
		}
		if idx == 1 {
			return itemID, "", true
		}
		return itemID, items[idx-2].ID, true
	case 1:
		if idx >= len(items)-1 {
			return "", "", false
		}
		return itemID, items[idx+1].ID, true
	default:
		return "", "", false
	}
}

// CycleIndex advances an index by delta within [0, n).
func CycleIndex(current, delta, n int) int {
	if n <= 0 {
		return 0
	}
	return (current + delta + n) % n
}
