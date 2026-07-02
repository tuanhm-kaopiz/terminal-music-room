package cli

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

type mockPlaybackDriver struct {
	mu      sync.Mutex
	running bool
	stops   int
	plays   []string
}

func (m *mockPlaybackDriver) Start(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

func (m *mockPlaybackDriver) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stops++
	m.running = false
	return nil
}

func (m *mockPlaybackDriver) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *mockPlaybackDriver) Play(_ context.Context, videoID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plays = append(m.plays, videoID)
	return nil
}

func (m *mockPlaybackDriver) Pause(context.Context) error  { return nil }
func (m *mockPlaybackDriver) Resume(context.Context) error { return nil }
func (m *mockPlaybackDriver) Seek(context.Context, int64) error {
	return nil
}
func (m *mockPlaybackDriver) Seekable(context.Context) (bool, error) { return true, nil }
func (m *mockPlaybackDriver) PositionMs(context.Context) (int64, error) { return 0, nil }

func TestRuntimeLocalPlaybackStartStop(t *testing.T) {
	t.Setenv("MUSIC_ROOM_NO_PLAYBACK", "")

	store := state.NewStore()
	track := protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000}
	if err := store.Apply(mustEnv(t, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{
		Slug: "r",
		Playback: protocol.PlaybackState{
			Status:     protocol.PlaybackPlaying,
			Track:      &track,
			PositionMs: 0,
		},
	})); err != nil {
		t.Fatal(err)
	}

	drv := &mockPlaybackDriver{}
	rt := &Runtime{store: store, playbackDriver: drv}
	rt.startLocalPlayback(context.Background())
	defer rt.stopLocalPlayback()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		drv.mu.Lock()
		n := len(drv.plays)
		drv.mu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected sync engine to load track")
}

func TestRuntimeLeaveStopsLocalPlayback(t *testing.T) {
	t.Setenv("MUSIC_ROOM_NO_PLAYBACK", "")
	_, ts, cfgPath := testHub(t)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Login(ctx, ioDiscard, cfgPath, "host", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	drv := &mockPlaybackDriver{}
	rt, err := newRuntimeFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	rt.playbackDriver = drv
	defer rt.Close()

	if err := rt.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}
	if err := rt.send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: "leave-audio"}); err != nil {
		t.Fatal(err)
	}
	if err := rt.waitInRoom("leave-audio", defaultWait); err != nil {
		t.Fatal(err)
	}
	track := protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000}
	_ = rt.store.Apply(mustEnv(t, protocol.MsgPlaybackState, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &track,
		PositionMs: 0,
	}))
	rt.startLocalPlayback(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		drv.mu.Lock()
		n := len(drv.plays)
		drv.mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if err := runLeave(ctx, rt, nil); err != nil {
		t.Fatalf("leave: %v", err)
	}
	if drv.stops == 0 {
		t.Fatal("expected mpv stopped on leave")
	}
}

// ioDiscard avoids importing io in test-only helper name clash.
var ioDiscard = discardWriter{}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
