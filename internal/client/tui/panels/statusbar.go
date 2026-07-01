package panels

import (
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// StatusBar renders the bottom shortcut hint row.
func StatusBar(tm theme.Theme, v state.View, width int, isHost bool) string {
	hint := "? help · space pause · s skip · a add · v vote · q quit"
	if isHost {
		hint += " · d del · ctrl+↑↓ reorder"
	}
	if v.LastErr != nil && v.LastErr.Message != "" {
		errLine := tm.Error().Render(truncate(v.LastErr.Message, width))
		return errLine + "\n" + tm.Muted().Render(truncate(hint, width))
	}
	return tm.Muted().Render(truncate(hint, width))
}
