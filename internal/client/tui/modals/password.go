package modals

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// PasswordIntent labels the password modal for create vs join.
type PasswordIntent int

const (
	PasswordJoin PasswordIntent = iota
	PasswordCreate
)

// Password is a masked room-password modal.
type Password struct {
	Input   textinput.Model
	Intent  PasswordIntent
	Slug    string
}

// NewPassword builds a focused masked password prompt.
func NewPassword(width int, intent PasswordIntent, slug string) Password {
	in := textinput.New()
	in.Placeholder = "leave empty for open room"
	if intent == PasswordJoin {
		in.Placeholder = "room password"
	}
	in.CharLimit = 32
	in.EchoMode = textinput.EchoPassword
	in.Focus()
	in.Prompt = "> "
	in.Width = max(0, min(width-16, 40))
	return Password{Input: in, Intent: intent, Slug: slug}
}

// WithWidth updates input width after resize.
func (m Password) WithWidth(width int) Password {
	m.Input.Width = max(0, min(width-16, 40))
	return m
}

// Update handles bubble events for the password field.
func (m Password) Update(msg tea.Msg) (Password, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

// Value returns the entered password (may be empty).
func (m Password) Value() string {
	return m.Input.Value()
}

// View renders the password overlay.
func (m Password) View(tm theme.Theme, width int) string {
	title := "JOIN ROOM"
	hint := "Enter password for " + m.Slug
	if m.Intent == PasswordCreate {
		title = "CREATE ROOM"
		hint = "Optional password for " + m.Slug + " (empty = open)"
	}
	lines := []string{
		tm.Title().Render(title),
		tm.Muted().Render(hint),
		"",
		m.Input.View(),
		tm.Muted().Render("Enter submit · Esc cancel"),
	}
	innerW := max(40, width-6)
	return tm.Panel(true).Width(innerW).Render(strings.Join(lines, "\n"))
}
