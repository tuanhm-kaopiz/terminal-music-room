package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/panels"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func testStore(t *testing.T) *state.Store {
	t.Helper()
	store := state.NewStore()
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", panels.FixtureView().Room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestModelIsHost(t *testing.T) {
	v := panels.FixtureView()
	if !IsHost(v) {
		t.Fatal("expected host session")
	}
	v.SessionID = "sess-guest"
	if IsHost(v) {
		t.Fatal("guest should not be host")
	}
}

func TestModelDefaults(t *testing.T) {
	m := NewModel(context.Background(), Config{Store: testStore(t)})
	if m.mode != ModeDashboard {
		t.Fatalf("mode = %v, want dashboard", m.mode)
	}
	if m.focus != FocusChat {
		t.Fatalf("focus = %v, want chat", m.focus)
	}
}

func TestModelConfigRoomActionsFallback(t *testing.T) {
	store := state.NewStore()
	cfg := Config{
		Store: store,
		Send: func(ctx context.Context, msgType string, payload any) error {
			return nil
		},
	}
	act := cfg.roomActions()
	if act == nil {
		t.Fatal("expected actions from Send+Store fallback")
	}
}

func TestModelConfigRoomActionsExplicit(t *testing.T) {
	store := state.NewStore()
	act := actions.New(nil, store)
	cfg := Config{Store: store, Actions: act}
	if cfg.roomActions() != act {
		t.Fatal("expected explicit Actions")
	}
}

func TestModelQuitWithoutLeave(t *testing.T) {
	store := state.NewStore()
	left := false
	m := NewModel(context.Background(), Config{
		Store: store,
		Leave: func(ctx context.Context) error {
			left = true
			return nil
		},
	})
	next, cmd := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if left {
		t.Fatal("q must exit TUI without leaving room (AC-004)")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !next.(*Model).quit {
		t.Fatal("expected quit flag")
	}
}

func TestModelHelpToggle(t *testing.T) {
	store := state.NewStore()
	m := NewModel(context.Background(), Config{Store: store})

	m2, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m2.(*Model).mode != ModeHelp {
		t.Fatalf("mode = %v, want help", m2.(*Model).mode)
	}

	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m3.(*Model).mode != ModeDashboard {
		t.Fatalf("mode = %v, want dashboard after esc", m3.(*Model).mode)
	}
}

func TestModelRefreshClampsSelectedQueueIdx(t *testing.T) {
	store := state.NewStore()
	m := NewModel(context.Background(), Config{Store: store})
	m.selectedQueueIdx = 5
	m.view = state.View{
		InRoom: true,
		Room: protocol.RoomSnapshot{
			Queue: []protocol.QueueItem{{ID: "q1"}},
		},
	}
	m.refresh()
	if m.selectedQueueIdx != 0 {
		t.Fatalf("selectedQueueIdx = %d, want 0", m.selectedQueueIdx)
	}
}
