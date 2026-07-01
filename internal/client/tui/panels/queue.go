package panels

import (
	"fmt"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Queue renders upcoming tracks with optional scroll offset.
func Queue(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	innerW, innerH := innerSize(width, height)
	lines := []string{tm.Title().Render("QUEUE")}
	items := v.Room.Queue
	if len(items) == 0 {
		lines = append(lines, tm.Muted().Render("(empty)"))
		return wrapPanel(tm, opts.Focused, width, height, lines)
	}
	start := opts.QueueScroll
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		start = 0
	}
	for i, item := range items[start:] {
		if len(lines) >= innerH {
			break
		}
		absoluteIdx := start + i
		marker := " "
		if absoluteIdx == opts.QueueSelectedIdx {
			marker = "›"
		}
		line := fmt.Sprintf("%s %d. %s", marker, absoluteIdx+1, item.Title)
		if item.AddedBy != "" {
			line += " · " + item.AddedBy
		}
		lines = append(lines, truncate(line, innerW))
	}
	if start+innerH-1 < len(items) {
		lines = append(lines, tm.Muted().Render("…"))
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
