package keys

import (
	"errors"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func playingView() state.View {
	return state.View{
		InRoom: true,
		Room: protocol.RoomSnapshot{
			Playback: protocol.PlaybackState{
				Status: protocol.PlaybackPlaying,
				Track:  &protocol.Track{Title: "Neon Nights"},
			},
		},
	}
}

func pausedView() state.View {
	v := playingView()
	v.Room.Playback.Status = protocol.PlaybackPaused
	return v
}

func emptyView() state.View {
	return state.View{InRoom: true, Room: protocol.RoomSnapshot{}}
}

func TestPlaybackTogglePlaying(t *testing.T) {
	action, err := PlaybackToggle(playingView())
	if err != nil {
		t.Fatal(err)
	}
	if action != PlaybackPause {
		t.Fatalf("action = %v, want pause", action)
	}
}

func TestPlaybackTogglePaused(t *testing.T) {
	action, err := PlaybackToggle(pausedView())
	if err != nil {
		t.Fatal(err)
	}
	if action != PlaybackResume {
		t.Fatalf("action = %v, want resume", action)
	}
}

func TestPlaybackNoTrackGuard(t *testing.T) {
	_, err := PlaybackToggle(emptyView())
	if !errors.Is(err, ErrNoTrack) {
		t.Fatalf("err = %v, want ErrNoTrack", err)
	}
	if err := RequireTrack(emptyView()); !errors.Is(err, ErrNoTrack) {
		t.Fatalf("RequireTrack err = %v", err)
	}
}
