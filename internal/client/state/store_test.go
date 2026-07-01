package state

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func mustEnvelope(t *testing.T, msgType string, payload any) protocol.Envelope {
	t.Helper()
	env, err := protocol.NewEnvelope(msgType, "c1", payload)
	if err != nil {
		t.Fatal(err)
	}
	return env
}

func TestApplyRoomSnapshot(t *testing.T) {
	s := NewStore()
	snap := protocol.RoomSnapshot{
		Slug:   "team",
		HostID: "h1",
		Members: []protocol.Member{
			{SessionID: "h1", Nickname: "host", IsHost: true},
		},
		Playback: protocol.PlaybackState{Status: protocol.PlaybackPlaying},
	}
	if err := s.Apply(mustEnvelope(t, protocol.MsgRoomSnapshot, snap)); err != nil {
		t.Fatal(err)
	}
	v := s.Snapshot()
	if !v.InRoom || v.Room.Slug != "team" || len(v.Room.Members) != 1 {
		t.Fatalf("view %+v", v)
	}
}

func TestApplyIncrementalUpdates(t *testing.T) {
	s := NewStore()
	_ = s.Apply(mustEnvelope(t, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{
		Slug:    "team",
		HostID:  "h1",
		Members: []protocol.Member{{SessionID: "h1", Nickname: "host", IsHost: true}},
	}))

	_ = s.Apply(mustEnvelope(t, protocol.MsgRoomMemberJoined, protocol.RoomMemberJoinedPayload{
		Member: protocol.Member{SessionID: "g1", Nickname: "guest"},
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgRoomMemberLeft, protocol.RoomMemberLeftPayload{SessionID: "g1"}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgRoomHostChanged, protocol.RoomHostChangedPayload{HostSessionID: "h1"}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgPlaybackState, protocol.PlaybackState{
		Status: protocol.PlaybackPaused, PositionMs: 1000,
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgPlaybackTick, protocol.PlaybackTickPayload{
		PositionMs: 1500, Status: protocol.PlaybackPlaying, ServerTime: time.Now(),
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgQueueUpdated, protocol.QueueUpdatedPayload{
		Items: []protocol.QueueItem{{ID: "q1", Title: "next"}},
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgChatMessage, protocol.ChatMessage{
		ID: "m1", Kind: protocol.ChatKindUser, Body: "hi",
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgVoteUpdated, protocol.VoteUpdatedPayload{
		Vote: &protocol.Vote{Kind: protocol.VoteKindSkip, Threshold: 2},
	}))
	_ = s.Apply(mustEnvelope(t, protocol.MsgReactionUpdated, protocol.ReactionUpdatedPayload{
		Counts: map[string]int{"🔥": 2},
	}))

	v := s.Snapshot()
	if v.Room.Playback.PositionMs != 1500 {
		t.Fatalf("tick position %d", v.Room.Playback.PositionMs)
	}
	if len(v.Room.Queue) != 1 || len(v.Room.Chat) != 1 {
		t.Fatalf("queue/chat %+v", v.Room)
	}
	if v.Room.Reactions["🔥"] != 2 {
		t.Fatalf("reactions %+v", v.Room.Reactions)
	}
}

func TestClearRoom(t *testing.T) {
	s := NewStore()
	_ = s.Apply(mustEnvelope(t, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "x"}))
	s.ClearRoom()
	v := s.Snapshot()
	if v.InRoom || v.Room.Slug != "" {
		t.Fatalf("view %+v", v)
	}
}
