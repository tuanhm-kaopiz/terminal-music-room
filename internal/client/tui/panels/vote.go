package panels

import (
	"fmt"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Vote renders in-progress skip or priority vote progress (AC-030–033).
func Vote(tm theme.Theme, v state.View, width int) string {
	vote := v.Room.Vote
	if vote == nil {
		return tm.Muted().Render("no active vote")
	}
	votes := len(vote.Voters)
	threshold := vote.Threshold
	if threshold <= 0 {
		threshold = 1
	}
	label := "SKIP"
	suffix := ""
	if vote.Kind == protocol.VoteKindPriority {
		label = "PRIORITY"
		if title := queueItemTitle(v.Room.Queue, vote.TargetID); title != "" {
			suffix = " · " + title
		}
	}
	barW := max(8, min(width-12, 24))
	line := fmt.Sprintf("%s %d/%d%s", label, votes, threshold, suffix)
	bar := tm.ProgressBar(votes, threshold, barW)
	return tm.Warning().Render(truncate(line, max(1, width-4))) + "\n" + bar
}

func queueItemTitle(items []protocol.QueueItem, id string) string {
	for _, item := range items {
		if item.ID == id {
			return item.Title
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Signals renders vote + reactions in the signals panel slot.
func Signals(tm theme.Theme, v state.View, width, height int, opts RenderOpts) string {
	lines := []string{
		tm.Title().Render("SIGNALS"),
		Vote(tm, v, width),
		Reactions(tm, v, width),
	}
	return wrapPanel(tm, opts.Focused, width, height, lines)
}
