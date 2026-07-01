package panels

import (
	"fmt"
	"sort"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// QuickReactionEmojis are the 1–4 keyboard shortcuts (ADR-006).
var QuickReactionEmojis = []string{"🔥", "❤️", "😂", "👍"}

// Reactions renders aggregated emoji reaction counts for the current track (AC-034).
func Reactions(tm theme.Theme, v state.View, width int) string {
	if v.Room.Playback.Track == nil {
		return tm.Muted().Render("—")
	}
	if len(v.Room.Reactions) == 0 {
		return tm.Muted().Render("—")
	}
	keys := make([]string, 0, len(v.Room.Reactions))
	for emoji := range v.Room.Reactions {
		keys = append(keys, emoji)
	}
	sort.Strings(keys)
	var parts []string
	for _, emoji := range keys {
		parts = append(parts, fmt.Sprintf("%s%d", emoji, v.Room.Reactions[emoji]))
	}
	line := ""
	for i, p := range parts {
		if i > 0 {
			line += " "
		}
		line += p
	}
	hint := tm.Muted().Render("1–4 quick react")
	return truncate(line, max(1, width-4)) + "\n" + hint
}
