package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/tui/keys"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func focusTestModel(t *testing.T) Model {
	t.Helper()
	m := NewModel(context.Background(), Config{Store: chatTestStore(t)})
	m.width = 80
	m.height = 24
	return m
}

func TestFocusCycleTab(t *testing.T) {
	m := focusTestModel(t)
	if m.focus != FocusChat {
		t.Fatalf("focus = %v, want chat", m.focus)
	}
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyTab})
	got := next.(*Model)
	if got.focus != FocusMembers {
		t.Fatalf("focus = %v, want members after tab from chat", got.focus)
	}
}

func TestFocusCycleShiftTab(t *testing.T) {
	m := focusTestModel(t)
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	got := next.(*Model)
	if got.focus != FocusQueue {
		t.Fatalf("focus = %v, want queue after shift+tab from chat", got.focus)
	}
}

func TestFocusQueueScrollDown(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusQueue
	m.selectedQueueIdx = 0
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyDown})
	got := next.(*Model)
	if got.selectedQueueIdx != 1 {
		t.Fatalf("selectedQueueIdx = %d, want 1", got.selectedQueueIdx)
	}
}

func TestFocusQueueScrollUp(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusQueue
	m.selectedQueueIdx = 1
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyUp})
	got := next.(*Model)
	if got.selectedQueueIdx != 0 {
		t.Fatalf("selectedQueueIdx = %d, want 0", got.selectedQueueIdx)
	}
}

func TestFocusQueueEnsureVisible(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusQueue
	visible := m.queueVisibleRows()
	for i := 0; i < visible+3; i++ {
		m.view.Room.Queue = append(m.view.Room.Queue, protocol.QueueItem{
			ID:    fmt.Sprintf("extra-%d", i),
			Title: "overflow track",
		})
	}
	m.selectedQueueIdx = visible + 2
	m.queueScroll = 0
	m.ensureQueueVisible()
	if m.selectedQueueIdx < m.queueScroll || m.selectedQueueIdx >= m.queueScroll+visible {
		t.Fatalf("selection %d outside window [%d,%d)", m.selectedQueueIdx, m.queueScroll, m.queueScroll+visible)
	}
}

func TestFocusChatScroll(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusChat
	m.chatScroll = 0
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyUp})
	got := next.(*Model)
	if got.chatScroll != 1 {
		t.Fatalf("chatScroll = %d, want 1 after up", got.chatScroll)
	}
	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyDown})
	got = next.(*Model)
	if got.chatScroll != 0 {
		t.Fatalf("chatScroll = %d, want 0 after down", got.chatScroll)
	}
}

func TestFocusMembersScroll(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusMembers
	m.membersScroll = 0
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyUp})
	got := next.(*Model)
	if got.membersScroll != 1 {
		t.Fatalf("membersScroll = %d, want 1", got.membersScroll)
	}
}

func TestFocusRenderOptsWhenFocused(t *testing.T) {
	m := focusTestModel(t)
	m.focus = FocusQueue
	if !m.renderOpts(FocusQueue).Focused {
		t.Fatal("queue panel should be focused")
	}
	if m.renderOpts(FocusChat).Focused {
		t.Fatal("chat panel should not be focused")
	}
	out := m.View()
	if !strings.Contains(out, "QUEUE") {
		t.Fatal("expected queue in view")
	}
}

func TestFocusKeysDefined(t *testing.T) {
	if keys.KeyTab != "tab" || keys.KeyShiftTab != "shift+tab" {
		t.Fatal("unexpected tab key constants")
	}
	if keys.KeyUp != "up" || keys.KeyDown != "down" {
		t.Fatal("unexpected arrow key constants")
	}
}

func TestKeysCycleIndex(t *testing.T) {
	if got := keys.CycleIndex(1, 1, 3); got != 2 {
		t.Fatalf("CycleIndex(1,1,3) = %d, want 2", got)
	}
	if got := keys.CycleIndex(0, -1, 3); got != 2 {
		t.Fatalf("CycleIndex(0,-1,3) = %d, want 2", got)
	}
}
