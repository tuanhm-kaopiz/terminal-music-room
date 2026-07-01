package protocol

import "time"

// Client → server message types.
const (
	MsgSessionHello   = "session.hello"
	MsgRoomCreate     = "room.create"
	MsgRoomJoin       = "room.join"
	MsgRoomLeave      = "room.leave"
	MsgPlaybackPlay   = "playback.play"
	MsgPlaybackPause  = "playback.pause"
	MsgPlaybackResume = "playback.resume"
	MsgPlaybackSkip   = "playback.skip"
	MsgPlaybackSeek   = "playback.seek"
	MsgQueueAdd       = "queue.add"
	MsgQueueRemove    = "queue.remove"
	MsgQueueReorder   = "queue.reorder"
	MsgChatSend       = "chat.send"
	MsgVoteSkip       = "vote.skip"
	MsgVotePriority   = "vote.priority"
	MsgReactionSend   = "reaction.send"
)

// Server → client message types.
const (
	MsgSessionAck       = "session.ack"
	MsgRoomSnapshot     = "room.snapshot"
	MsgRoomMemberJoined = "room.member_joined"
	MsgRoomMemberLeft   = "room.member_left"
	MsgRoomHostChanged  = "room.host_changed"
	MsgPlaybackState    = "playback.state"
	MsgPlaybackTick     = "playback.tick"
	MsgQueueUpdated     = "queue.updated"
	MsgChatMessage      = "chat.message"
	MsgVoteUpdated      = "vote.updated"
	MsgReactionUpdated  = "reaction.updated"
	MsgError            = "error"
)

// --- Client payloads ---

type SessionHelloPayload struct {
	Nickname string `json:"nickname"`
}

type RoomCreatePayload struct {
	Slug string `json:"slug"`
}

type RoomJoinPayload struct {
	Slug string `json:"slug"`
}

type RoomLeavePayload struct{}

type PlaybackPlayPayload struct {
	Query string `json:"query,omitempty"`
	URL   string `json:"url,omitempty"`
}

type PlaybackPausePayload struct{}

type PlaybackResumePayload struct{}

type PlaybackSkipPayload struct{}

type PlaybackSeekPayload struct {
	PositionMs int64 `json:"position_ms"`
}

type QueueAddPayload struct {
	Query string `json:"query,omitempty"`
	URL   string `json:"url,omitempty"`
}

type QueueRemovePayload struct {
	ItemID string `json:"item_id"`
}

type QueueReorderPayload struct {
	ItemID  string `json:"item_id"`
	AfterID string `json:"after_id"`
}

type ChatSendPayload struct {
	Body string `json:"body"`
}

type VoteSkipPayload struct{}

type VotePriorityPayload struct {
	ItemID string `json:"item_id"`
}

type ReactionSendPayload struct {
	Emoji string `json:"emoji"`
}

// --- Server payloads ---

type SessionAckPayload struct {
	SessionID   string `json:"session_id"`
	DisplayName string `json:"display_name"`
}

type RoomMemberJoinedPayload struct {
	Member Member `json:"member"`
}

type RoomMemberLeftPayload struct {
	SessionID string `json:"session_id"`
}

type RoomHostChangedPayload struct {
	HostSessionID string `json:"host_session_id"`
}

type PlaybackTickPayload struct {
	PositionMs int64          `json:"position_ms"`
	Status     PlaybackStatus `json:"status"`
	ServerTime time.Time      `json:"server_time"`
}

type QueueUpdatedPayload struct {
	Items []QueueItem `json:"items"`
}

type VoteUpdatedPayload struct {
	Vote     *Vote         `json:"vote,omitempty"`
	Progress *VoteProgress `json:"progress,omitempty"`
}

type ReactionUpdatedPayload struct {
	Counts map[string]int `json:"counts"`
}

type ErrorPayload struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	RetryAfter *int      `json:"retry_after,omitempty"`
}
