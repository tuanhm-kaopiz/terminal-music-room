package modals

import (
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

func TestLeaveConfirmView(t *testing.T) {
	out := NewConfirmLeave("backend-team").View(theme.Default(), 80)
	for _, want := range []string{"LEAVE ROOM", "backend-team", "Esc cancel"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in view", want)
		}
	}
}

func TestLeaveConfirmEmptyRoom(t *testing.T) {
	out := NewConfirmLeave("").View(theme.Default(), 80)
	if !strings.Contains(out, "this room") {
		t.Fatalf("expected fallback room label: %q", out)
	}
}
