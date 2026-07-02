package sync

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

type mockDriver struct {
	mu sync.Mutex

	running  bool
	videoID  string
	paused   bool
	position int64

	starts int
	stops  int
	plays  []string
	seeks  []int64
	playErr   error
	seekable  bool
}

func (m *mockDriver) Seekable(context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seekable, nil
}

func (m *mockDriver) Start(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	m.starts++
	return nil
}

func (m *mockDriver) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	m.videoID = ""
	m.stops++
	return nil
}

func (m *mockDriver) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *mockDriver) Play(_ context.Context, videoID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.playErr != nil {
		return m.playErr
	}
	m.videoID = videoID
	m.plays = append(m.plays, videoID)
	m.paused = false
	return nil
}

func (m *mockDriver) Pause(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = true
	return nil
}

func (m *mockDriver) Resume(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = false
	return nil
}

func (m *mockDriver) Seek(_ context.Context, positionMs int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.position = positionMs
	m.seeks = append(m.seeks, positionMs)
	return nil
}

func (m *mockDriver) PositionMs(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.position, nil
}

func (m *mockDriver) lastSeek() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.seeks) == 0 {
		return -1
	}
	return m.seeks[len(m.seeks)-1]
}

func TestEffectiveServerMsFromTick(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 2, 0, time.UTC)
	view := state.View{
		Room: protocol.RoomSnapshot{
			Playback: protocol.PlaybackState{
				Status:     protocol.PlaybackPlaying,
				PositionMs: 1000,
				DurationMs: 120_000,
			},
		},
		LastTick: protocol.PlaybackTickPayload{
			PositionMs: 1000,
			Status:     protocol.PlaybackPlaying,
			ServerTime: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	got := EffectiveServerMs(view, now)
	if got != 3000 {
		t.Fatalf("got %d", got)
	}
}

func TestSyncLoadsNewTrack(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{seekable: true}
	engine := New(Config{Store: store, Player: driver, Now: time.Now})

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000},
		PositionMs: 500,
	})

	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if driver.videoID != "abc123xyz01" {
		t.Fatalf("video %q", driver.videoID)
	}
	if driver.starts != 1 {
		t.Fatalf("starts %d", driver.starts)
	}
	if got := driver.lastSeek(); got != 500 {
		t.Fatalf("seek %d", got)
	}
	if driver.paused {
		t.Fatal("expected playing")
	}
}

func TestSyncPauseResume(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{running: true, videoID: "abc123xyz01", seekable: true}
	engine := New(Config{Store: store, Player: driver})
	engine.lastVideoID = "abc123xyz01"

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPaused,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t"},
		PositionMs: 1000,
	})
	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !driver.paused {
		t.Fatal("expected paused")
	}

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t"},
		PositionMs: 1000,
		AnchorTime: time.Now(),
	})
	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if driver.paused {
		t.Fatal("expected resumed")
	}
}

func TestSyncDriftSeek(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 2, 0, time.UTC)
	store := state.NewStore()
	driver := &mockDriver{running: true, videoID: "abc123xyz01", position: 10_000, seekable: true}
	engine := New(Config{Store: store, Player: driver, DriftThresholdMs: 150, Now: func() time.Time { return now }})
	engine.lastVideoID = "abc123xyz01"

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 120_000},
		PositionMs: 5000,
	})
	applyPlaybackTick(t, store, protocol.PlaybackTickPayload{
		PositionMs: 5000,
		Status:     protocol.PlaybackPlaying,
		ServerTime: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	})

	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := driver.lastSeek(); got != 7000 {
		t.Fatalf("seek %d want 7000", got)
	}
}

func TestSyncNoSeekWithinThreshold(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 2, 0, time.UTC)
	store := state.NewStore()
	driver := &mockDriver{running: true, videoID: "abc123xyz01", position: 7050, seekable: true}
	engine := New(Config{Store: store, Player: driver, DriftThresholdMs: 150, Now: func() time.Time { return now }})
	engine.lastVideoID = "abc123xyz01"

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 120_000},
		PositionMs: 5000,
	})
	applyPlaybackTick(t, store, protocol.PlaybackTickPayload{
		PositionMs: 5000,
		Status:     protocol.PlaybackPlaying,
		ServerTime: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
	})

	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(driver.seeks) != 0 {
		t.Fatalf("unexpected seeks %v", driver.seeks)
	}
}

func TestSyncStopOnLeave(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{running: true, videoID: "abc123xyz01"}
	engine := New(Config{Store: store, Player: driver})
	engine.lastVideoID = "abc123xyz01"

	store.ClearRoom()
	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if driver.stops != 1 {
		t.Fatalf("stops %d", driver.stops)
	}
	if engine.LastVideoID() != "" {
		t.Fatal("expected cleared track")
	}
}

func TestRunReactsToPlaybackNotify(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{}
	engine := New(Config{Store: store, Player: driver, DriftPollInterval: time.Hour})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- engine.Run(ctx) }()

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status: protocol.PlaybackPlaying,
		Track:  &protocol.Track{VideoID: "abc123xyz01", Title: "t"},
	})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if driver.videoID == "abc123xyz01" {
			cancel()
			if err := <-done; err != nil && err != context.Canceled {
				t.Fatal(err)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatal("timeout waiting for playback sync")
}

func TestSyncSkipsSeekWhileNotSeekable(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{running: true, videoID: "abc123xyz01", position: 0, seekable: false}
	engine := New(Config{Store: store, Player: driver, DriftThresholdMs: 150, Now: time.Now})
	engine.lastVideoID = "abc123xyz01"

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000},
		PositionMs: 5000,
	})

	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := driver.lastSeek(); got != -1 {
		t.Fatalf("seek %d, want no seek while buffering", got)
	}
}

func TestSyncRetriesAfterPlayFailure(t *testing.T) {
	store := state.NewStore()
	driver := &mockDriver{seekable: true}
	engine := New(Config{Store: store, Player: driver, Now: time.Now})

	applyRoomPlayback(t, store, protocol.PlaybackState{
		Status:     protocol.PlaybackPlaying,
		Track:      &protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000},
		PositionMs: 500,
	})

	driver.playErr = fmt.Errorf("mpv busy")
	if err := engine.Sync(context.Background()); err == nil {
		t.Fatal("expected play error")
	}
	if engine.LastVideoID() != "" {
		t.Fatal("lastVideoID should stay empty after failed play")
	}

	driver.playErr = nil
	driver.seekable = true
	if err := engine.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if engine.LastVideoID() != "abc123xyz01" {
		t.Fatalf("video %q", engine.LastVideoID())
	}
}

func applyRoomPlayback(t *testing.T, store *state.Store, pb protocol.PlaybackState) {
	t.Helper()
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "t", protocol.RoomSnapshot{
		Slug:     "sync-room",
		Playback: pb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
}

func applyPlaybackState(t *testing.T, store *state.Store, pb protocol.PlaybackState) {
	t.Helper()
	env, err := protocol.NewEnvelope(protocol.MsgPlaybackState, "t", pb)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
}

func applyPlaybackTick(t *testing.T, store *state.Store, tick protocol.PlaybackTickPayload) {
	t.Helper()
	env, err := protocol.NewEnvelope(protocol.MsgPlaybackTick, "", tick)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
}
