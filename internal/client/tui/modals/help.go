package modals

import (
	"strings"

	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Help renders the keyboard shortcut overlay (ADR-006).
func Help(tm theme.Theme, width int, isHost bool) string {
	innerW := max(40, width-6)
	sections := []string{
		tm.Title().Render("KEYBOARD SHORTCUTS"),
		tm.Muted().Render("? or Esc close"),
		"",
		helpSection(tm, "Playback", []string{
			"Space     pause / resume",
			"s         skip track",
			"S         seek (milliseconds)",
		}),
		helpSection(tm, "Queue & source", []string{
			"a         add URL or search (Tab: play / queue)",
		}),
		helpSection(tm, "Social", []string{
			"Enter     send chat",
			"v         vote skip",
			"V         vote priority (selected queue item)",
			"1–4       quick react 🔥 ❤️ 😂 👍",
		}),
		helpSection(tm, "Navigation", []string{
			"Tab       cycle focus panel",
			"Shift+Tab cycle focus panel",
		}),
		helpSection(tm, "Session", []string{
			"q         exit TUI (stay in room)",
			"l         leave room (confirm)",
			"?         this help overlay",
		}),
	}
	if isHost {
		sections = append(sections, helpSection(tm, "Host only", []string{
			"d         remove selected queue item",
			"Ctrl+↑/↓  reorder selected queue item",
		}))
	}
	return tm.Panel(true).Width(innerW).Render(strings.Join(sections, "\n"))
}

func helpSection(tm theme.Theme, title string, lines []string) string {
	out := tm.Title().Render(title) + "\n"
	for _, line := range lines {
		out += tm.Muted().Render("  "+line) + "\n"
	}
	return strings.TrimSuffix(out, "\n")
}
