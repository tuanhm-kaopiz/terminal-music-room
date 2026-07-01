package playback

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func testTrack() protocol.Track {
	return protocol.Track{
		VideoID:    "abc123",
		Title:      "Lofi Mix",
		DurationMs: 120_000,
		SourceURL:  "https://youtube.com/watch?v=abc123",
	}
}

func TestEffectivePositionMsWhilePlaying(t *testing.T) {
	clk := NewClock()
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	clk.LoadTrack(testTrack(), start)
	clk.Play(start)

	now := start.Add(2500 * time.Millisecond)
	got := clk.EffectivePositionMs(now)
	if got != 2500 {
		t.Fatalf("position: got %d want 2500", got)
	}
}

func TestPauseCapturesEffectivePosition(t *testing.T) {
	clk := NewClock()
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	clk.LoadTrack(testTrack(), start)
	clk.Play(start)

	pauseAt := start.Add(1500 * time.Millisecond)
	clk.Pause(pauseAt)

	if clk.positionMs != 1500 {
		t.Fatalf("frozen position: got %d want 1500", clk.positionMs)
	}
	if clk.EffectivePositionMs(pauseAt.Add(time.Second)) != 1500 {
		t.Fatal("position should not advance while paused")
	}
	if clk.Status() != protocol.PlaybackPaused {
		t.Fatalf("status %q", clk.Status())
	}
}

func TestSeekWhilePlaying(t *testing.T) {
	clk := NewClock()
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	clk.LoadTrack(testTrack(), start)
	clk.Play(start)
	clk.SeekTo(30_000, start.Add(time.Second))

	now := start.Add(2 * time.Second)
	if clk.EffectivePositionMs(now) != 31_000 {
		t.Fatalf("got %d want 31000", clk.EffectivePositionMs(now))
	}
}

func TestSeekClampsToDuration(t *testing.T) {
	clk := NewClock()
	start := time.Now()
	clk.LoadTrack(testTrack(), start)
	clk.SeekTo(999_000, start)
	if clk.positionMs != 120_000 {
		t.Fatalf("got %d want 120000", clk.positionMs)
	}
}

func TestMarkEnded(t *testing.T) {
	clk := NewClock()
	start := time.Now()
	clk.LoadTrack(testTrack(), start)
	clk.Play(start)
	clk.MarkEnded()

	if clk.Status() != protocol.PlaybackEnded {
		t.Fatalf("status %q", clk.Status())
	}
	if clk.positionMs != 120_000 {
		t.Fatalf("position %d", clk.positionMs)
	}
}

func TestStateRoundTripFields(t *testing.T) {
	clk := NewClock()
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	clk.LoadTrack(testTrack(), start)
	clk.Play(start)

	st := clk.State()
	if st.Status != protocol.PlaybackPlaying {
		t.Fatalf("status %q", st.Status)
	}
	if st.Track == nil || st.Track.VideoID != "abc123" {
		t.Fatalf("track %+v", st.Track)
	}
	if st.DurationMs != 120_000 {
		t.Fatalf("duration %d", st.DurationMs)
	}
}
