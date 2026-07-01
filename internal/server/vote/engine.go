package vote

import (
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Config controls vote session behavior.
type Config struct {
	Timeout time.Duration
}

// DefaultConfig returns production defaults.
func DefaultConfig() Config {
	return Config{Timeout: defaultTimeoutSeconds * time.Second}
}

// Result describes the outcome of casting or checking a vote.
type Result struct {
	Vote       *protocol.Vote
	Progress   protocol.VoteProgress
	Passed     bool
	PassedKind protocol.VoteKind
	Cancelled  bool
	CancelMsg  string
}

// Threshold returns votes required for >50% of online members (AC-038, AC-042).
func Threshold(onlineCount int) int {
	if onlineCount <= 0 {
		return 1
	}
	return onlineCount/2 + 1
}

// Progress builds a tally from an active vote.
func Progress(v *protocol.Vote) protocol.VoteProgress {
	if v == nil {
		return protocol.VoteProgress{}
	}
	return protocol.VoteProgress{
		Votes:      len(v.Voters),
		Threshold:  v.Threshold,
		OnlineSnap: v.OnlineSnap,
	}
}

// IsExpired reports whether a vote session timed out (AC-040).
func IsExpired(v *protocol.Vote, now time.Time, timeout time.Duration) bool {
	if v == nil || timeout <= 0 {
		return false
	}
	return now.Sub(v.StartedAt) >= timeout
}

func hasVoted(v *protocol.Vote, sessionID string) bool {
	for _, id := range v.Voters {
		if id == sessionID {
			return true
		}
	}
	return false
}

func addVoter(v *protocol.Vote, sessionID string) {
	if hasVoted(v, sessionID) {
		return
	}
	v.Voters = append(v.Voters, sessionID)
}

func startVote(kind protocol.VoteKind, targetID string, onlineCount int, sessionID string, now time.Time) *protocol.Vote {
	v := &protocol.Vote{
		Kind:       kind,
		TargetID:   targetID,
		StartedAt:  now,
		OnlineSnap: onlineCount,
		Threshold:  Threshold(onlineCount),
	}
	addVoter(v, sessionID)
	return v
}

func passed(v *protocol.Vote) bool {
	return len(v.Voters) >= v.Threshold
}
