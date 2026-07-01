package modals

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/keys"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Seek is the position-ms modal overlay (AC-015, FR-004).
type Seek struct {
	Input textinput.Model
}

// NewSeek builds a focused seek modal.
func NewSeek(width int) Seek {
	in := textinput.New()
	in.Placeholder = "position in milliseconds"
	in.CharLimit = 12
	in.Focus()
	in.Prompt = "> "
	in.Width = max(0, width-6)
	return Seek{Input: in}
}

// WithWidth updates the input width after terminal resize.
func (m Seek) WithWidth(width int) Seek {
	m.Input.Width = max(0, width-6)
	return m
}

// View renders the seek modal overlay panel.
func (m Seek) View(tm theme.Theme, width int) string {
	innerW := max(40, width-6)
	lines := []string{
		tm.Title().Render("SEEK"),
		tm.Muted().Render("Enter position (ms) · Enter submit · Esc cancel"),
		m.Input.View(),
	}
	return tm.Panel(true).Width(innerW).Render(strings.Join(lines, "\n"))
}

// Update forwards messages to the modal text input.
func (m Seek) Update(msg tea.Msg) (Seek, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

// Submit validates input and dispatches seek via actions.
func (m Seek) Submit(ctx context.Context, act *actions.Room, v state.View) error {
	if act == nil {
		return fmt.Errorf("not connected")
	}
	if err := keys.RequireTrack(v); err != nil {
		return err
	}
	position := strings.TrimSpace(m.Input.Value())
	if position == "" {
		return fmt.Errorf("position required — use milliseconds")
	}
	return act.SeekFromString(ctx, position)
}
