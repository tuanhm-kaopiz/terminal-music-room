package modals

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

func TestPasswordMaskedInput(t *testing.T) {
	m := NewPassword(80, PasswordJoin, "locked-room")
	if m.Input.EchoMode != textinput.EchoPassword {
		t.Fatalf("echo mode = %v, want password", m.Input.EchoMode)
	}
}

func TestPasswordViewJoin(t *testing.T) {
	m := NewPassword(80, PasswordJoin, "locked-room")
	out := m.View(theme.Default(), 80)
	if out == "" {
		t.Fatal("expected rendered modal")
	}
}

func TestPasswordValue(t *testing.T) {
	m := NewPassword(80, PasswordJoin, "room")
	m.Input.SetValue("secret")
	if m.Value() != "secret" {
		t.Fatalf("value = %q", m.Value())
	}
}
