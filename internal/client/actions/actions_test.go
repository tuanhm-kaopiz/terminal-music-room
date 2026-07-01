package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantURL string
		wantQ   string
		wantErr bool
	}{
		{name: "youtube url", input: "https://youtube.com/watch?v=abc", wantURL: "https://youtube.com/watch?v=abc"},
		{name: "youtu.be", input: "https://youtu.be/abc", wantURL: "https://youtu.be/abc"},
		{name: "search query", input: "lofi hip hop", wantQ: "lofi hip hop"},
		{name: "empty", input: "  ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, query, err := ParseSource(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if url != tt.wantURL || query != tt.wantQ {
				t.Fatalf("got url=%q query=%q", url, query)
			}
		})
	}
}

func TestParseSourceArgs(t *testing.T) {
	url, query, err := ParseSourceArgs([]string{"jazz", "playlist"})
	if err != nil {
		t.Fatal(err)
	}
	if url != "" || query != "jazz playlist" {
		t.Fatalf("got url=%q query=%q", url, query)
	}

	_, _, err = ParseSourceArgs(nil)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestPlaybackPlayPayload(t *testing.T) {
	p, err := PlaybackPlayPayload("https://youtu.be/x", "")
	if err != nil || p.URL == "" {
		t.Fatalf("unexpected: %+v %v", p, err)
	}
	_, err = PlaybackPlayPayload("u", "q")
	if err == nil {
		t.Fatal("expected error for both url and query")
	}
	_, err = PlaybackPlayPayload("", "")
	if err == nil {
		t.Fatal("expected error for neither")
	}
}

func TestQueueAddPayload(t *testing.T) {
	p, err := QueueAddPayload("", "ambient")
	if err != nil || p.Query != "ambient" {
		t.Fatalf("unexpected: %+v %v", p, err)
	}
}

func TestRoomRequireInRoom(t *testing.T) {
	r := New(nil, state.NewStore())
	if err := r.Pause(context.Background()); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("got %v", err)
	}
}

func TestRoomPlaybackMessages(t *testing.T) {
	var gotType string
	var gotPayload any
	store := state.NewStore()
	mustApply(t, store, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "test"})

	r := New(func(_ context.Context, msgType string, payload any) error {
		gotType = msgType
		gotPayload = payload
		return nil
	}, store)

	if err := r.Play(context.Background(), "chill beats"); err != nil {
		t.Fatal(err)
	}
	if gotType != protocol.MsgPlaybackPlay {
		t.Fatalf("type %q", gotType)
	}
	p, ok := gotPayload.(protocol.PlaybackPlayPayload)
	if !ok || p.Query != "chill beats" {
		t.Fatalf("payload %+v", gotPayload)
	}

	if err := r.Pause(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotType != protocol.MsgPlaybackPause {
		t.Fatalf("type %q", gotType)
	}

	if err := r.Seek(context.Background(), 1500); err != nil {
		t.Fatal(err)
	}
	sp, ok := gotPayload.(protocol.PlaybackSeekPayload)
	if !ok || sp.PositionMs != 1500 {
		t.Fatalf("payload %+v", gotPayload)
	}

	if err := r.Seek(context.Background(), -1); err == nil {
		t.Fatal("expected seek validation error")
	}
}

func TestRoomQueueMessages(t *testing.T) {
	var gotType string
	store := state.NewStore()
	mustApply(t, store, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "q"})
	r := New(func(_ context.Context, msgType string, _ any) error {
		gotType = msgType
		return nil
	}, store)

	if err := r.QueueAdd(context.Background(), "https://youtube.com/watch?v=1"); err != nil {
		t.Fatal(err)
	}
	if gotType != protocol.MsgQueueAdd {
		t.Fatalf("type %q", gotType)
	}

	if err := r.QueueRemove(context.Background(), "item-1"); err != nil {
		t.Fatal(err)
	}
	if gotType != protocol.MsgQueueRemove {
		t.Fatalf("type %q", gotType)
	}

	if err := r.QueueReorder(context.Background(), "a", "b"); err != nil {
		t.Fatal(err)
	}
	if gotType != protocol.MsgQueueReorder {
		t.Fatalf("type %q", gotType)
	}
}

func TestRoomSocialMessages(t *testing.T) {
	store := state.NewStore()
	mustApply(t, store, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "s"})

	var types []string
	r := New(func(_ context.Context, msgType string, _ any) error {
		types = append(types, msgType)
		return nil
	}, store)

	if err := r.Chat(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if err := r.Chat(context.Background(), "   "); err == nil {
		t.Fatal("expected empty chat error")
	}
	if err := r.VoteSkip(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.VotePriority(context.Background(), "q1"); err != nil {
		t.Fatal(err)
	}
	if err := r.React(context.Background(), "🔥"); err != nil {
		t.Fatal(err)
	}
	if err := r.Leave(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		protocol.MsgChatSend,
		protocol.MsgVoteSkip,
		protocol.MsgVotePriority,
		protocol.MsgReactionSend,
		protocol.MsgRoomLeave,
	}
	if len(types) != len(want) {
		t.Fatalf("types %v", types)
	}
	for i, w := range want {
		if types[i] != w {
			t.Fatalf("idx %d: got %q want %q", i, types[i], w)
		}
	}
}

func mustApply(t *testing.T, store *state.Store, msgType string, payload any) {
	t.Helper()
	env, err := protocol.NewEnvelope(msgType, "", payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
}
