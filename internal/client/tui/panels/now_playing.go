package panels

import (
	"fmt"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// NowPlaying renders track title, progress bar, and playback status.
func NowPlaying(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	innerW, _ := innerSize(width, height)
	pb := v.Room.Playback
	title := "(nothing playing)"
	if pb.Track != nil && pb.Track.Title != "" {
		title = pb.Track.Title
	}
	dur := pb.DurationMs
	if dur <= 0 && pb.Track != nil {
		dur = pb.Track.DurationMs
	}
	pos := pb.PositionMs
	if dur > 0 {
		bar := tm.ProgressBar(int(pos), int(dur), innerW)
		status := fmt.Sprintf("%s  %s / %s", pb.Status, formatMs(pos), formatMs(dur))
		lines := []string{
			tm.Title().Render("NOW PLAYING"),
			"▶ " + title,
			bar,
			status,
		}
		return wrapPanel(tm, opts.Focused, width, height, lines)
	}
	lines := []string{
		tm.Title().Render("NOW PLAYING"),
		"▶ " + title,
		tm.Muted().Render(string(pb.Status)),
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
