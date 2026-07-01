package protocol_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestRoundTripClientMessages(t *testing.T) {
	cases := []struct {
		name    string
		msgType string
		id      string
		payload any
	}{
		{"session.hello", protocol.MsgSessionHello, "c1", protocol.SessionHelloPayload{Nickname: "kaopiz"}},
		{"room.create", protocol.MsgRoomCreate, "c2", protocol.RoomCreatePayload{Slug: "backend-team"}},
		{"room.join", protocol.MsgRoomJoin, "c3", protocol.RoomJoinPayload{Slug: "backend-team"}},
		{"room.leave", protocol.MsgRoomLeave, "c4", protocol.RoomLeavePayload{}},
		{"playback.play url", protocol.MsgPlaybackPlay, "c5", protocol.PlaybackPlayPayload{URL: "https://youtube.com/watch?v=abc"}},
		{"playback.play query", protocol.MsgPlaybackPlay, "c6", protocol.PlaybackPlayPayload{Query: "lofi"}},
		{"playback.seek", protocol.MsgPlaybackSeek, "c7", protocol.PlaybackSeekPayload{PositionMs: 151000}},
		{"queue.add", protocol.MsgQueueAdd, "c8", protocol.QueueAddPayload{Query: "rain"}},
		{"queue.remove", protocol.MsgQueueRemove, "c9", protocol.QueueRemovePayload{ItemID: "q1"}},
		{"queue.reorder", protocol.MsgQueueReorder, "c10", protocol.QueueReorderPayload{ItemID: "q1", AfterID: "q0"}},
		{"chat.send", protocol.MsgChatSend, "c11", protocol.ChatSendPayload{Body: "hello 🔥"}},
		{"vote.skip", protocol.MsgVoteSkip, "c12", protocol.VoteSkipPayload{}},
		{"vote.priority", protocol.MsgVotePriority, "c13", protocol.VotePriorityPayload{ItemID: "q2"}},
		{"reaction.send", protocol.MsgReactionSend, "c14", protocol.ReactionSendPayload{Emoji: "🔥"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := protocol.EncodeMessage(tc.msgType, tc.id, tc.payload)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			env, err := protocol.Decode(data)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if env.Type != tc.msgType {
				t.Fatalf("type: got %q want %q", env.Type, tc.msgType)
			}
			if env.ID != tc.id {
				t.Fatalf("id: got %q want %q", env.ID, tc.id)
			}
		})
	}
}

func TestDecodePayloadRoomJoin(t *testing.T) {
	data, err := protocol.EncodeMessage(protocol.MsgRoomJoin, "req-1", protocol.RoomJoinPayload{Slug: "team"})
	if err != nil {
		t.Fatal(err)
	}
	env, payload, err := protocol.DecodePayload[protocol.RoomJoinPayload](data)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != protocol.MsgRoomJoin {
		t.Fatalf("type %q", env.Type)
	}
	if payload.Slug != "team" {
		t.Fatalf("slug %q", payload.Slug)
	}
}

func TestServerMessages(t *testing.T) {
	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)

	t.Run("session.ack", func(t *testing.T) {
		data, err := protocol.EncodeMessage(protocol.MsgSessionAck, "", protocol.SessionAckPayload{
			SessionID:   "sess-1",
			DisplayName: "kaopiz#a1b2",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, p, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
		if err != nil {
			t.Fatal(err)
		}
		if p.SessionID != "sess-1" || p.DisplayName != "kaopiz#a1b2" {
			t.Fatalf("payload %+v", p)
		}
	})

	t.Run("room.snapshot", func(t *testing.T) {
		snap := protocol.RoomSnapshot{
			Slug:   "backend-team",
			HostID: "sess-1",
			Members: []protocol.Member{{
				SessionID: "sess-1", Nickname: "kaopiz", DisplayName: "kaopiz", JoinedAt: now, IsHost: true,
			}},
			Playback: protocol.PlaybackState{Status: protocol.PlaybackPaused},
			Reactions: map[string]int{"🔥": 2},
		}
		data, err := protocol.EncodeMessage(protocol.MsgRoomSnapshot, "", snap)
		if err != nil {
			t.Fatal(err)
		}
		_, p, err := protocol.DecodePayload[protocol.RoomSnapshot](data)
		if err != nil {
			t.Fatal(err)
		}
		if p.Slug != "backend-team" || len(p.Members) != 1 {
			t.Fatalf("snapshot %+v", p)
		}
	})

	t.Run("playback.tick", func(t *testing.T) {
		data, err := protocol.EncodeMessage(protocol.MsgPlaybackTick, "", protocol.PlaybackTickPayload{
			PositionMs: 31000,
			Status:     protocol.PlaybackPlaying,
			ServerTime: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, p, err := protocol.DecodePayload[protocol.PlaybackTickPayload](data)
		if err != nil {
			t.Fatal(err)
		}
		if p.PositionMs != 31000 || p.Status != protocol.PlaybackPlaying {
			t.Fatalf("tick %+v", p)
		}
	})

	t.Run("error", func(t *testing.T) {
		retry := 30
		env, err := protocol.NewErrorEnvelope("e1", protocol.ErrRateLimited, "slow down", &retry)
		if err != nil {
			t.Fatal(err)
		}
		data, err := protocol.Encode(env)
		if err != nil {
			t.Fatal(err)
		}
		_, p, err := protocol.DecodePayload[protocol.ErrorPayload](data)
		if err != nil {
			t.Fatal(err)
		}
		if p.Code != protocol.ErrRateLimited || p.RetryAfter == nil || *p.RetryAfter != 30 {
			t.Fatalf("error %+v", p)
		}
	})
}

func TestDecodeMissingType(t *testing.T) {
	_, err := protocol.Decode([]byte(`{"id":"x","payload":{}}`))
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestIsKnownErrorCode(t *testing.T) {
	if !protocol.IsKnownErrorCode(protocol.ErrRoomNotFound) {
		t.Fatal("ROOM_NOT_FOUND should be known")
	}
	if protocol.IsKnownErrorCode(protocol.ErrorCode("NOPE")) {
		t.Fatal("unknown code should return false")
	}
}

func TestEnvelopeWireFormat(t *testing.T) {
	data, err := protocol.EncodeMessage(protocol.MsgRoomJoin, "corr-uuid", protocol.RoomJoinPayload{Slug: "jam"})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"type", "id", "payload"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("missing key %q in %s", key, string(data))
		}
	}
}
