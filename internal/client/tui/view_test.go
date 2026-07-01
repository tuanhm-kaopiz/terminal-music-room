package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/terminal-music-room/music-room/internal/client/tui/panels"
)

func testHUDModel(width, height int) Model {
	in := textinput.New()
	in.Prompt = "> "
	in.Width = width - 4
	return Model{
		width:  width,
		height: height,
		view:   panels.FixtureView(),
		input:  in,
	}
}

func TestView80x24(t *testing.T) {
	out := testHUDModel(80, 24).View()
	checks := []string{
		"backend-team",
		"NOW PLAYING",
		"Neon Nights",
		"CREW",
		"SIGNALS",
		"QUEUE",
		"COMMS",
		"Track Two",
		"hello",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in view output", want)
		}
	}
}

func TestView120x40(t *testing.T) {
	out := testHUDModel(120, 40).View()
	if !strings.Contains(out, "NOW PLAYING") || !strings.Contains(out, "QUEUE") {
		t.Fatalf("expected full HUD at 120x40: len=%d", len(out))
	}
	if strings.Contains(out, "degraded HUD") {
		t.Fatal("should not be degraded at 120x40")
	}
}

func TestViewDegraded60x20(t *testing.T) {
	out := testHUDModel(60, 20).View()
	if !strings.Contains(out, "degraded HUD") {
		t.Fatalf("expected degraded warning: %q", out[:min(200, len(out))])
	}
	if !strings.Contains(out, "Neon Nights") {
		t.Fatal("now playing should remain in degraded mode")
	}
	if strings.Contains(out, "COMMS") {
		t.Fatal("chat panel should be hidden in degraded layout")
	}
}

func TestViewLoading(t *testing.T) {
	m := Model{}
	if !strings.Contains(m.View(), "Loading") {
		t.Fatal("expected loading state")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
