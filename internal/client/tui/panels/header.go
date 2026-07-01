package panels

import (
	"fmt"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Header renders the top HUD strip with room slug, online count, and conn badge.
func Header(tm theme.Theme, v state.View, width, height int) string {
	room := v.Room.Slug
	if room == "" {
		room = "(no room)"
	}
	online := len(v.Room.Members)
	conn := connBadge(tm, v)
	title := tm.Header().Render(fmt.Sprintf("◈ ROOM: %s", room))
	meta := tm.Muted().Render(fmt.Sprintf("CREW: %d", online))
	line := truncate(fmt.Sprintf("%s  %s  %s", title, meta, conn), max(1, width-4))
	lines := []string{line, "", ""}
	return wrapPanel(tm, false, width, height, lines)
}

func connBadge(tm theme.Theme, v state.View) string {
	switch {
	case v.Status == state.StatusConnected:
		return connStyle(tm, v.Status).Render(connLabel(v.Status))
	case v.Status == state.StatusReconnecting, v.Status == state.StatusConnecting:
		return tm.Warning().Render("reconnecting…")
	case v.Status == state.StatusDisconnected && v.SessionID != "" && !v.InRoom:
		return tm.Error().Render("disconnected — rejoin")
	case v.Status == state.StatusDisconnected && v.InRoom:
		return tm.Warning().Render("disconnected (retry)")
	default:
		return connStyle(tm, v.Status).Render(connLabel(v.Status))
	}
}
