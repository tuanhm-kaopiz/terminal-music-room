package modals

import (
	"fmt"

	"github.com/terminal-music-room/music-room/internal/client/tui/panels"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// ConfirmLeave is the leave-room confirmation overlay.
type ConfirmLeave struct {
	Room string
}

// NewConfirmLeave builds a confirmation modal for the current room slug.
func NewConfirmLeave(room string) ConfirmLeave {
	if room == "" {
		room = "this room"
	}
	return ConfirmLeave{Room: room}
}

// View renders the leave confirmation overlay card.
func (c ConfirmLeave) View(tm theme.Theme, width int) string {
	lines := []string{
		tm.Title().Render("LEAVE ROOM"),
		tm.Warning().Render(fmt.Sprintf("Leave %s?", c.Room)),
		tm.Muted().Render("y / Enter confirm · Esc cancel"),
	}
	return panels.OverlayCard(tm, width, lines)
}
