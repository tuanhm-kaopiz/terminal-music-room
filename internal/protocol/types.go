package protocol

import "time"

// PlaybackStatus is the authoritative playback state reported by the server.
type PlaybackStatus string

const (
	PlaybackPlaying   PlaybackStatus = "playing"
	PlaybackPaused    PlaybackStatus = "paused"
	PlaybackBuffering PlaybackStatus = "buffering"
	PlaybackEnded     PlaybackStatus = "ended"
)

// ChatKind distinguishes user chat from system events.
type ChatKind string

const (
	ChatKindUser   ChatKind = "user"
	ChatKindSystem ChatKind = "system"
)

// VoteKind is skip-current or priority-next-queue-item.
type VoteKind string

const (
	VoteKindSkip     VoteKind = "skip"
	VoteKindPriority VoteKind = "priority"
)

// Track is a YouTube audio source (v1).
type Track struct {
	VideoID    string `json:"video_id"`
	Title      string `json:"title"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	SourceURL  string `json:"source_url,omitempty"`
}

// Member is a participant in a room.
type Member struct {
	SessionID   string    `json:"session_id"`
	Nickname    string    `json:"nickname"`
	DisplayName string    `json:"display_name"`
	JoinedAt    time.Time `json:"joined_at"`
	IsHost      bool      `json:"is_host"`
}

// PlaybackState is server-authoritative playback snapshot.
type PlaybackState struct {
	Status     PlaybackStatus `json:"status"`
	Track      *Track         `json:"track,omitempty"`
	PositionMs int64          `json:"position_ms"`
	AnchorTime time.Time      `json:"anchor_time,omitempty"`
	DurationMs int64          `json:"duration_ms,omitempty"`
}

// QueueItem is an upcoming track in the room queue.
type QueueItem struct {
	ID         string    `json:"id"`
	VideoID    string    `json:"video_id"`
	Title      string    `json:"title"`
	DurationMs int64     `json:"duration_ms,omitempty"`
	AddedBy    string    `json:"added_by"`
	AddedAt    time.Time `json:"added_at"`
}

// ChatMessage is a room chat or system line.
type ChatMessage struct {
	ID     string    `json:"id"`
	Kind   ChatKind  `json:"kind"`
	Author string    `json:"author,omitempty"`
	Body   string    `json:"body"`
	At     time.Time `json:"at"`
}

// Vote tracks an in-progress skip or priority vote.
type Vote struct {
	Kind       VoteKind  `json:"kind"`
	TargetID   string    `json:"target_id,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	Voters     []string  `json:"voters"`
	Threshold  int       `json:"threshold"`
	OnlineSnap int       `json:"online_snap"`
}

// VoteProgress is broadcast vote tally state.
type VoteProgress struct {
	Votes      int `json:"votes"`
	Threshold  int `json:"threshold"`
	OnlineSnap int `json:"online_snap"`
}

// RoomSnapshot is the full room state sent after join or reconnect.
type RoomSnapshot struct {
	Slug              string         `json:"slug"`
	HostID            string         `json:"host_session_id"`
	PasswordProtected bool           `json:"password_protected"`
	Members           []Member       `json:"members"`
	Playback  PlaybackState  `json:"playback"`
	Queue     []QueueItem    `json:"queue"`
	Chat      []ChatMessage  `json:"chat"`
	Vote      *Vote          `json:"vote,omitempty"`
	Reactions map[string]int `json:"reactions,omitempty"`
}
