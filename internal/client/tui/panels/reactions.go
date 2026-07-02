package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// QuickReactionEmojis are the 1–4 keyboard shortcuts (ADR-006).
var QuickReactionEmojis = []string{"🔥", "❤️", "😂", "👍"}

// Reactions renders aggregated emoji reaction counts for the current track (AC-034).
func Reactions(tm theme.Theme, v state.View, width int) string {
	return strings.Join(reactionLines(tm, v, width), "\n")
}

func reactionLines(tm theme.Theme, v state.View, width int) []string {
	if v.Room.Playback.Track == nil {
		return []string{tm.Muted().Render("—")}
	}
	if len(v.Room.Reactions) == 0 {
		return []string{tm.Muted().Render("—")}
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
	line := strings.Join(parts, " ")
	hint := tm.Muted().Render("1–4 quick react")
	return []string{
		truncate(line, max(1, width-4)),
		hint,
	}
}
