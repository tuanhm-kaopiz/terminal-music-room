package modals

import (
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

func TestHelpViewMember(t *testing.T) {
	out := Help(theme.Default(), 80, false)
	checks := []string{
		"KEYBOARD SHORTCUTS",
		"pause / resume",
		"seek (milliseconds)",
		"exit TUI",
		"vote skip",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in help", want)
		}
	}
	if strings.Contains(out, "Host only") {
		t.Fatal("member help should not list host-only section")
	}
}

func TestHelpViewHost(t *testing.T) {
	out := Help(theme.Default(), 80, true)
	if !strings.Contains(out, "Host only") {
		t.Fatal("host help should list host-only keys")
	}
	if !strings.Contains(out, "remove selected") {
		t.Fatal("expected queue remove hint")
	}
}
