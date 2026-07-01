package modals

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func inRoomStore(t *testing.T) *state.Store {
	t.Helper()
	store := state.NewStore()
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", protocol.RoomSnapshot{Slug: "team"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestAddSourceSubmitPlay(t *testing.T) {
	store := inRoomStore(t)
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewAddSource(80)
	m.Intent = IntentPlay
	m.Input.SetValue("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err := m.Submit(context.Background(), act); err != nil {
		t.Fatal(err)
	}
	if sent != protocol.MsgPlaybackPlay {
		t.Fatalf("sent %q, want play", sent)
	}
}

func TestAddSourceSubmitQueue(t *testing.T) {
	store := inRoomStore(t)
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewAddSource(80)
	m.Intent = IntentQueue
	m.Input.SetValue("neon nights synthwave")
	if err := m.Submit(context.Background(), act); err != nil {
		t.Fatal(err)
	}
	if sent != protocol.MsgQueueAdd {
		t.Fatalf("sent %q, want queue.add", sent)
	}
}

func TestAddSourceRejectEmpty(t *testing.T) {
	store := inRoomStore(t)
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		t.Fatal("should not send empty source")
		return nil
	}, store)

	m := NewAddSource(80)
	m.Input.SetValue("   ")
	if err := m.Submit(context.Background(), act); err == nil {
		t.Fatal("expected error for empty input (AC-027)")
	}
}

func TestAddSourceToggleIntent(t *testing.T) {
	if ToggleIntent(IntentPlay) != IntentQueue {
		t.Fatal("expected queue intent")
	}
	if ToggleIntent(IntentQueue) != IntentPlay {
		t.Fatal("expected play intent")
	}
}

func TestAddSourceView(t *testing.T) {
	m := NewAddSource(80)
	out := m.View(theme.Default(), 80)
	for _, want := range []string{"ADD SOURCE", "PLAY NOW", "YouTube URL"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in view: %s", want, out)
		}
	}
	m.Intent = IntentQueue
	out = m.View(theme.Default(), 80)
	if !strings.Contains(out, "ADD TO QUEUE") {
		t.Fatalf("expected queue label in view: %s", out)
	}
}

func TestAddSourceUpdate(t *testing.T) {
	m := NewAddSource(80)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.Input.Value() != "x" {
		t.Fatalf("input = %q", m.Input.Value())
	}
}
