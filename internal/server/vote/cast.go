package vote

import (
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

// CastSkip records a skip vote or starts a new session (AC-037–039).
func CastSkip(r *room.Room, sessionID string, now time.Time, cfg Config) (Result, error) {
	if r.Playback.Track() == nil {
		return Result{}, ErrNoTrackPlaying
	}
	if res, ok := expireIfNeeded(r, now, cfg); ok {
		return res, nil
	}
	if r.Vote != nil && r.Vote.Kind != protocol.VoteKindSkip {
		return Result{}, ErrVoteKindConflict
	}

	if r.Vote == nil {
		r.Vote = startVote(protocol.VoteKindSkip, "", r.MemberCount(), sessionID, now)
	} else {
		addVoter(r.Vote, sessionID)
	}
	progress := Progress(r.Vote)
	if !passed(r.Vote) {
		return Result{Vote: r.Vote, Progress: progress}, nil
	}

	r.Skip(now)
	r.Vote = nil
	return Result{Passed: true, PassedKind: protocol.VoteKindSkip, Progress: progress}, nil
}

// CastPriority records a priority vote for a queue item (AC-041–043).
func CastPriority(r *room.Room, sessionID, itemID string, now time.Time, cfg Config) (Result, error) {
	if len(r.Queue) < 2 {
		return Result{}, ErrQueueTooShort
	}
	if r.QueueIndex(itemID) < 0 {
		return Result{}, ErrInvalidTarget
	}
	if res, ok := expireIfNeeded(r, now, cfg); ok {
		return res, nil
	}
	if res, ok := cancelPriorityIfMissing(r); ok {
		return res, nil
	}

	if r.Vote != nil && r.Vote.Kind != protocol.VoteKindPriority {
		return Result{}, ErrVoteKindConflict
	}
	if r.Vote != nil && r.Vote.TargetID != itemID {
		return Result{}, ErrVoteKindConflict
	}

	if r.Vote == nil {
		r.Vote = startVote(protocol.VoteKindPriority, itemID, r.MemberCount(), sessionID, now)
	} else {
		addVoter(r.Vote, sessionID)
	}
	progress := Progress(r.Vote)
	if !passed(r.Vote) {
		return Result{Vote: r.Vote, Progress: progress}, nil
	}

	if err := r.PromoteQueueItem(itemID); err != nil {
		r.Vote = nil
		return Result{Cancelled: true, CancelMsg: "vote cancelled: queue item removed"}, nil
	}
	r.Vote = nil
	return Result{Passed: true, PassedKind: protocol.VoteKindPriority, Progress: progress}, nil
}

// ClearOnTrackChange ends skip votes when playback advances (AC-040).
func ClearOnTrackChange(r *room.Room) (Result, bool) {
	if r.Vote == nil || r.Vote.Kind != protocol.VoteKindSkip {
		return Result{}, false
	}
	r.Vote = nil
	return Result{Cancelled: true, CancelMsg: "vote ended: track changed"}, true
}

// CheckRoomVotes expires or cancels stale votes; returns true when state changed.
func CheckRoomVotes(r *room.Room, now time.Time, cfg Config) (Result, bool) {
	if res, ok := expireIfNeeded(r, now, cfg); ok {
		return res, true
	}
	if res, ok := cancelPriorityIfMissing(r); ok {
		return res, true
	}
	return Result{}, false
}

func expireIfNeeded(r *room.Room, now time.Time, cfg Config) (Result, bool) {
	if r.Vote == nil {
		return Result{}, false
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultConfig().Timeout
	}
	if !IsExpired(r.Vote, now, timeout) {
		return Result{}, false
	}
	r.Vote = nil
	return Result{Cancelled: true, CancelMsg: "vote timed out"}, true
}

func cancelPriorityIfMissing(r *room.Room) (Result, bool) {
	if r.Vote == nil || r.Vote.Kind != protocol.VoteKindPriority {
		return Result{}, false
	}
	if r.QueueIndex(r.Vote.TargetID) >= 0 {
		return Result{}, false
	}
	r.Vote = nil
	return Result{Cancelled: true, CancelMsg: "vote cancelled: queue item removed"}, true
}
