package panels

import (
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Members renders the crew list with host marker *.
func Members(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	innerW, innerH := panelInnerSize(width, height)
	lines := []string{tm.Title().Render("CREW")}
	members := v.Room.Members
	if len(members) == 0 {
		lines = append(lines, tm.Muted().Render("(empty)"))
		return wrapPanel(tm, opts.Focused, width, height, lines)
	}
	start := opts.MembersScroll
	if start < 0 {
		start = 0
	}
	if start >= len(members) {
		start = 0
	}
	memberSlots := innerH - 1
	moreBelow := start+memberSlots < len(members)
	if moreBelow {
		memberSlots--
	}
	for i, member := range members[start:] {
		if len(lines)-1 >= memberSlots {
			break
		}
		absoluteIdx := start + i
		marker := " "
		if absoluteIdx == opts.MembersSelectedIdx {
			marker = "›"
		}
		name := member.DisplayName
		if name == "" {
			name = member.Nickname
		}
		prefix := marker + " "
		if member.IsHost {
			prefix += tm.HostMarker().Render("*") + " "
		} else {
			prefix += "  "
		}
		lines = append(lines, truncate(prefix+name, innerW))
	}
	if moreBelow {
		lines = append(lines, tm.Muted().Render("…"))
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
