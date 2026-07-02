package room

import (
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
	"github.com/terminal-music-room/music-room/internal/server/playback"
)

const maxMembers = 20

// Room is the in-memory aggregate for a music room session.
type Room struct {
	Slug          string
	HostSessionID string
	CreatedAt     time.Time
	passwordHash  []byte
	Members       []protocol.Member
	Playback      *playback.Clock
	Queue         []protocol.QueueItem
	Chat          *chat.Buffer
	Vote          *protocol.Vote
	Reactions     map[string]int
}

// NewRoom creates a room with the given slug and host member.
func NewRoom(slug string, host protocol.Member, now time.Time, chatOpts chat.Options) *Room {
	host.IsHost = true
	host.JoinedAt = now
	if host.DisplayName == "" {
		host.DisplayName = host.Nickname
	}
	return &Room{
		Slug:          slug,
		HostSessionID: host.SessionID,
		CreatedAt:     now,
		Members:       []protocol.Member{host},
		Playback:      playback.NewClock(),
		Chat:          chat.NewBuffer(chatOpts, slug),
		Reactions:     make(map[string]int),
	}
}

// MemberCount returns online member count.
func (r *Room) MemberCount() int {
	return len(r.Members)
}

// IsFull reports whether the room reached capacity (AC-009).
func (r *Room) IsFull() bool {
	return len(r.Members) >= maxMembers
}

// FindMember returns a member by session ID.
func (r *Room) FindMember(sessionID string) (protocol.Member, bool) {
	for _, m := range r.Members {
		if m.SessionID == sessionID {
			return m, true
		}
	}
	return protocol.Member{}, false
}

// AddMember appends a member and recomputes display names.
func (r *Room) AddMember(m protocol.Member, now time.Time) error {
	if r.IsFull() {
		return ErrRoomFull
	}
	if _, ok := r.FindMember(m.SessionID); ok {
		return ErrAlreadyMember
	}
	m.JoinedAt = now
	m.IsHost = false
	r.Members = append(r.Members, m)
	r.Members = RecomputeDisplayNames(r.Members)
	return nil
}

// RemoveMember removes a member and handles host transfer (AC-013).
// Returns emptied=true when the last member left (AC-014).
func (r *Room) RemoveMember(sessionID string) (emptied bool, hostChanged bool) {
	idx := -1
	for i, m := range r.Members {
		if m.SessionID == sessionID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false, false
	}
	wasHost := r.Members[idx].SessionID == r.HostSessionID
	r.Members = append(r.Members[:idx], r.Members[idx+1:]...)
	if len(r.Members) == 0 {
		return true, false
	}
	r.Members = RecomputeDisplayNames(r.Members)
	if wasHost {
		r.HostSessionID = r.Members[0].SessionID
		for i := range r.Members {
			r.Members[i].IsHost = r.Members[i].SessionID == r.HostSessionID
		}
		return false, true
	}
	return false, false
}

// PasswordProtected reports whether join requires a password.
func (r *Room) PasswordProtected() bool {
	return len(r.passwordHash) > 0
}

// SetPassword validates and stores a bcrypt hash for the room.
func (r *Room) SetPassword(plain string) error {
	trimmed, err := ValidatePassword(plain)
	if err != nil {
		return err
	}
	hash, err := HashPassword(trimmed)
	if err != nil {
		return err
	}
	r.passwordHash = hash
	return nil
}

// CheckPassword reports whether plain matches the room password (open rooms always match).
func (r *Room) CheckPassword(plain string) bool {
	if !r.PasswordProtected() {
		return true
	}
	trimmed, err := ValidatePassword(plain)
	if err != nil {
		return false
	}
	return CheckPassword(r.passwordHash, trimmed)
}

// Snapshot builds the wire snapshot for join/reconnect (AC-010).
func (r *Room) Snapshot(now time.Time) protocol.RoomSnapshot {
	playbackState := r.Playback.State()
	if r.Playback.Status() == protocol.PlaybackPlaying {
		playbackState.PositionMs = r.Playback.EffectivePositionMs(now)
	}
	reactions := make(map[string]int, len(r.Reactions))
	for k, v := range r.Reactions {
		reactions[k] = v
	}
	members := make([]protocol.Member, len(r.Members))
	copy(members, r.Members)
	queue := make([]protocol.QueueItem, len(r.Queue))
	copy(queue, r.Queue)
	chat := r.Chat.Messages()
	var vote *protocol.Vote
	if r.Vote != nil {
		v := *r.Vote
		vote = &v
	}
	return protocol.RoomSnapshot{
		Slug:              r.Slug,
		HostID:            r.HostSessionID,
		PasswordProtected: r.PasswordProtected(),
		Members:           members,
		Playback:  playbackState,
		Queue:     queue,
		Chat:      chat,
		Vote:      vote,
		Reactions: reactions,
	}
}
