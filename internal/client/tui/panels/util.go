package panels

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func innerSize(width, height int) (int, int) {
	return max(1, width-4), max(1, height-2)
}

func wrapPanel(tm theme.Theme, focused bool, width, height int, lines []string) string {
	innerW, innerH := innerSize(width, height)
	trimmed := trimLines(lines, innerW)
	padded := padLines(trimmed, innerH)
	return tm.Panel(focused).Width(width).Height(height).Render(strings.Join(padded[:innerH], "\n"))
}

func trimLines(lines []string, innerW int) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = truncate(line, innerW)
	}
	return out
}

func padLines(lines []string, innerH int) []string {
	out := append([]string(nil), lines...)
	for len(out) < innerH {
		out = append(out, "")
	}
	return out
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > max-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteRune('…')
	return b.String()
}

func formatMs(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	sec := ms / 1000
	min := sec / 60
	sec %= 60
	return fmt.Sprintf("%d:%02d", min, sec)
}

func formatChat(msg protocol.ChatMessage) string {
	if msg.Kind == protocol.ChatKindSystem {
		return "[sys] " + msg.Body
	}
	author := msg.Author
	if author == "" {
		author = "?"
	}
	return author + ": " + msg.Body
}

func connLabel(status state.ConnStatus) string {
	switch status {
	case state.StatusConnected:
		return "connected"
	case state.StatusReconnecting:
		return "reconnecting"
	case state.StatusConnecting:
		return "connecting"
	default:
		return "disconnected"
	}
}

// ConnLabel exports the connection badge text for tests (AC-041).
func ConnLabel(status state.ConnStatus) string {
	return connLabel(status)
}

func connStyle(tm theme.Theme, status state.ConnStatus) themeStyle {
	switch status {
	case state.StatusConnected:
		return tm.Success()
	case state.StatusReconnecting, state.StatusConnecting:
		return tm.Warning()
	default:
		return tm.Error()
	}
}

// themeStyle avoids exporting lipgloss from helpers.
type themeStyle interface {
	Render(...string) string
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
