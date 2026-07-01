package modals

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// AddIntent selects play-now vs queue-add on submit.
type AddIntent int

const (
	IntentPlay AddIntent = iota
	IntentQueue
)

// AddSource is the URL/search modal overlay (AC-017, FR-005).
type AddSource struct {
	Input  textinput.Model
	Intent AddIntent
}

// NewAddSource builds a focused add-source modal.
func NewAddSource(width int) AddSource {
	in := textinput.New()
	in.Placeholder = "YouTube URL or search query"
	in.CharLimit = 500
	in.Focus()
	in.Prompt = "> "
	in.Width = max(0, width-6)
	return AddSource{Input: in, Intent: IntentPlay}
}

// WithWidth updates the input width after terminal resize.
func (m AddSource) WithWidth(width int) AddSource {
	m.Input.Width = max(0, width-6)
	return m
}

// ToggleIntent switches between play-now and queue-add.
func ToggleIntent(intent AddIntent) AddIntent {
	if intent == IntentPlay {
		return IntentQueue
	}
	return IntentPlay
}

func (m AddSource) intentLabel() string {
	if m.Intent == IntentQueue {
		return "ADD TO QUEUE"
	}
	return "PLAY NOW"
}

// View renders the modal overlay panel.
func (m AddSource) View(tm theme.Theme, width int) string {
	innerW := max(40, width-6)
	hint := fmt.Sprintf("Tab: %s · Enter submit · Esc cancel", m.intentLabel())
	lines := []string{
		tm.Title().Render("ADD SOURCE"),
		tm.Muted().Render(hint),
		m.Input.View(),
	}
	return tm.Panel(true).Width(innerW).Render(strings.Join(lines, "\n"))
}

// Update forwards messages to the modal text input.
func (m AddSource) Update(msg tea.Msg) (AddSource, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

// Submit validates input and dispatches play or queue-add via actions.
func (m AddSource) Submit(ctx context.Context, act *actions.Room) error {
	if act == nil {
		return fmt.Errorf("not connected")
	}
	source := strings.TrimSpace(m.Input.Value())
	if source == "" {
		return fmt.Errorf("provide a YouTube URL or search query")
	}
	switch m.Intent {
	case IntentPlay:
		return act.Play(ctx, source)
	case IntentQueue:
		return act.QueueAdd(ctx, source)
	default:
		return act.Play(ctx, source)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
