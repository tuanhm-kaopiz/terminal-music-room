package panels

import (
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Chat renders recent chat messages (newest visible within panel height).
func Chat(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	innerW, innerH := innerSize(width, height)
	lines := []string{tm.Title().Render("COMMS")}
	msgs := v.Room.Chat
	start := 0
	if len(msgs) > innerH-1 {
		start = len(msgs) - (innerH - 1) - opts.ChatScroll
		if start < 0 {
			start = 0
		}
	}
	for _, msg := range msgs[start:] {
		if len(lines) >= innerH {
			break
		}
		lines = append(lines, truncate(formatChat(msg), innerW))
	}
	if start > 0 {
		lines = append(lines, tm.Muted().Render("…"))
	}
	if len(v.Room.Chat) == 0 {
		lines = append(lines, tm.Muted().Render("(no messages yet)"))
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
