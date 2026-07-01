package modals

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/keys"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func seekStore(t *testing.T, withTrack bool) *state.Store {
	t.Helper()
	store := state.NewStore()
	room := protocol.RoomSnapshot{Slug: "team"}
	if withTrack {
		room.Playback = protocol.PlaybackState{
			Status: protocol.PlaybackPlaying,
			Track:  &protocol.Track{Title: "Neon Nights", DurationMs: 180000},
		}
	}
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func playingView() state.View {
	return state.View{
		InRoom: true,
		Room: protocol.RoomSnapshot{
			Playback: protocol.PlaybackState{
				Status: protocol.PlaybackPlaying,
				Track:  &protocol.Track{Title: "Neon Nights"},
			},
		},
	}
}

func TestSeekSubmit(t *testing.T) {
	store := seekStore(t, true)
	var sent int64
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		if msgType != protocol.MsgPlaybackSeek {
			t.Fatalf("msgType = %q", msgType)
		}
		p, ok := payload.(protocol.PlaybackSeekPayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		sent = p.PositionMs
		return nil
	}, store)

	m := NewSeek(80)
	m.Input.SetValue("60000")
	if err := m.Submit(context.Background(), act, playingView()); err != nil {
		t.Fatal(err)
	}
	if sent != 60000 {
		t.Fatalf("position = %d, want 60000", sent)
	}
}

func TestSeekRejectEmpty(t *testing.T) {
	store := seekStore(t, true)
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		t.Fatal("should not send empty position")
		return nil
	}, store)

	m := NewSeek(80)
	m.Input.SetValue("   ")
	if err := m.Submit(context.Background(), act, playingView()); err == nil {
		t.Fatal("expected error for empty position")
	}
}

func TestSeekInvalidPosition(t *testing.T) {
	store := seekStore(t, true)
	act := actions.New(nil, store)
	m := NewSeek(80)
	m.Input.SetValue("not-a-number")
	if err := m.Submit(context.Background(), act, playingView()); err == nil {
		t.Fatal("expected invalid position error")
	}
}

func TestSeekNoTrackGuard(t *testing.T) {
	store := seekStore(t, false)
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		t.Fatal("should not seek without track")
		return nil
	}, store)

	m := NewSeek(80)
	m.Input.SetValue("1000")
	err := m.Submit(context.Background(), act, state.View{InRoom: true})
	if !errors.Is(err, keys.ErrNoTrack) {
		t.Fatalf("err = %v, want ErrNoTrack", err)
	}
}

func TestSeekView(t *testing.T) {
	out := NewSeek(80).View(theme.Default(), 80)
	for _, want := range []string{"SEEK", "milliseconds", "Esc cancel"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in view", want)
		}
	}
}

func TestSeekUpdate(t *testing.T) {
	m := NewSeek(80)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	if m.Input.Value() != "9" {
		t.Fatalf("input = %q", m.Input.Value())
	}
}
