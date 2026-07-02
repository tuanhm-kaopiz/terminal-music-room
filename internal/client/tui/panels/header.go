package panels

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// Header renders a single-line room strip (no bordered panel — saves vertical space).
func Header(tm theme.Theme, v state.View, width, _ int) string {
	room := v.Room.Slug
	if room == "" {
		room = "(no room)"
	}
	online := len(v.Room.Members)
	conn := connBadge(tm, v)
	title := tm.Header().Render(fmt.Sprintf("◈ ROOM: %s", room))
	meta := tm.Muted().Render(fmt.Sprintf("CREW: %d", online))
	line := truncate(fmt.Sprintf("%s  %s  %s", title, meta, conn), max(1, width))
	return lipgloss.NewStyle().Width(width).Render(line)
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
