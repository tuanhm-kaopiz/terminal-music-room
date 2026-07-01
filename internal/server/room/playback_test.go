package room

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

func TestRoomSkipAdvancesQueue(t *testing.T) {
	now := time.Now()
	r := NewRoom("test", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "cur", Title: "Current", DurationMs: 60_000}, now)
	r.Playback.Play(now)
	r.Queue = []protocol.QueueItem{
		{ID: "q1", VideoID: "next", Title: "Next", DurationMs: 90_000},
	}

	r.Skip(now)

	if r.Playback.Track() == nil || r.Playback.Track().VideoID != "next" {
		t.Fatalf("track %+v", r.Playback.Track())
	}
	if r.Playback.Status() != protocol.PlaybackPlaying {
		t.Fatalf("status %q", r.Playback.Status())
	}
	if len(r.Queue) != 0 {
		t.Fatalf("queue len %d", len(r.Queue))
	}
}

func TestRoomSkipEndsWhenQueueEmpty(t *testing.T) {
	now := time.Now()
	r := NewRoom("test", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "cur", Title: "Current", DurationMs: 60_000}, now)
	r.Playback.Play(now)

	r.Skip(now)

	if r.Playback.Status() != protocol.PlaybackEnded {
		t.Fatalf("status %q", r.Playback.Status())
	}
}
