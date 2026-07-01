package room

import (
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// QueueItemTrack converts a queue item to a playable track.
func QueueItemTrack(item protocol.QueueItem) protocol.Track {
	return protocol.Track{
		VideoID:    item.VideoID,
		Title:      item.Title,
		DurationMs: item.DurationMs,
		SourceURL:  "https://www.youtube.com/watch?v=" + item.VideoID,
	}
}

// Skip abandons the current track and plays the next queue item, or marks ended (AC-024).
func (r *Room) Skip(now time.Time) {
	r.ResetReactions()
	if len(r.Queue) == 0 {
		r.Playback.MarkEnded()
		return
	}
	item := r.Queue[0]
	r.Queue = r.Queue[1:]
	r.Playback.LoadTrack(QueueItemTrack(item), now)
	r.Playback.Play(now)
}
