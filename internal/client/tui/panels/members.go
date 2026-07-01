package panels

import (
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Members renders the crew list with host marker *.
func Members(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	innerW, innerH := innerSize(width, height)
	lines := []string{tm.Title().Render("CREW")}
	members := v.Room.Members
	if len(members) == 0 {
		lines = append(lines, tm.Muted().Render("(empty)"))
	} else {
		start := opts.MembersScroll
		if start < 0 {
			start = 0
		}
		if start >= len(members) {
			start = 0
		}
		for _, member := range members[start:] {
			if len(lines) >= innerH {
				break
			}
			name := member.DisplayName
			if name == "" {
				name = member.Nickname
			}
			line := "  " + name
			if member.IsHost {
				line = tm.HostMarker().Render("*") + " " + name
			}
			lines = append(lines, truncate(line, innerW))
		}
		if start > 0 {
			lines = append(lines, tm.Muted().Render("…"))
		}
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
