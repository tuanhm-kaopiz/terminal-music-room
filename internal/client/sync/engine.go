package sync

import (
	"context"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

const (
	// DefaultDriftThresholdMs is the seek correction threshold (architecture §sync).
	DefaultDriftThresholdMs int64 = 150
	defaultDriftPoll              = time.Second
)

// Driver controls local mpv playback for the sync engine.
type Driver interface {
	Start(ctx context.Context) error
	Stop() error
	Running() bool
	Play(ctx context.Context, videoID string) error
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
	Seek(ctx context.Context, positionMs int64) error
	PositionMs(ctx context.Context) (int64, error)
}

// Config configures the playback sync engine.
type Config struct {
	Store              *state.Store
	Player             Driver
	DriftThresholdMs   int64
	DriftPollInterval  time.Duration
	Now                func() time.Time
}

// Engine maps server playback state to local mpv (FR-007, NFR-003).
type Engine struct {
	cfg Config

	mu          sync.Mutex
	lastVideoID string
}

// New creates a sync engine.
func New(cfg Config) *Engine {
	if cfg.DriftThresholdMs <= 0 {
		cfg.DriftThresholdMs = DefaultDriftThresholdMs
	}
	if cfg.DriftPollInterval <= 0 {
		cfg.DriftPollInterval = defaultDriftPoll
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Engine{cfg: cfg}
}

// Run subscribes to playback.state/tick via the store and keeps mpv in sync until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	if e.cfg.Store == nil || e.cfg.Player == nil {
		return nil
	}
	playbackCh := e.cfg.Store.SubscribePlayback()
	ticker := time.NewTicker(e.cfg.DriftPollInterval)
	defer ticker.Stop()

	if err := e.Sync(ctx); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-playbackCh:
			if err := e.Sync(ctx); err != nil {
				return err
			}
		case <-ticker.C:
			if err := e.Sync(ctx); err != nil {
				return err
			}
		}
	}
}

// Sync applies the current store playback snapshot to mpv once.
func (e *Engine) Sync(ctx context.Context) error {
	if e.cfg.Store == nil || e.cfg.Player == nil {
		return nil
	}
	view := e.cfg.Store.Snapshot()
	return e.syncView(ctx, view)
}

func (e *Engine) syncView(ctx context.Context, view state.View) error {
	pb := view.Room.Playback
	track := pb.Track

	e.mu.Lock()
	defer e.mu.Unlock()

	if !view.InRoom || track == nil || pb.Status == protocol.PlaybackEnded {
		if e.lastVideoID != "" {
			_ = e.cfg.Player.Stop()
			e.lastVideoID = ""
		}
		return nil
	}

	videoID := track.VideoID
	if videoID != e.lastVideoID {
		if err := e.ensurePlayer(ctx); err != nil {
			return err
		}
		if err := e.cfg.Player.Play(ctx, videoID); err != nil {
			return err
		}
		e.lastVideoID = videoID
		serverMs := EffectiveServerMs(view, e.cfg.Now())
		if err := e.cfg.Player.Seek(ctx, serverMs); err != nil {
			return err
		}
		return e.syncStatus(ctx, pb.Status)
	}

	if err := e.syncStatus(ctx, pb.Status); err != nil {
		return err
	}

	if pb.Status != protocol.PlaybackPlaying {
		return nil
	}

	serverMs := EffectiveServerMs(view, e.cfg.Now())
	localMs, err := e.cfg.Player.PositionMs(ctx)
	if err != nil {
		return err
	}
	if abs64(localMs-serverMs) > e.cfg.DriftThresholdMs {
		return e.cfg.Player.Seek(ctx, serverMs)
	}
	return nil
}

func (e *Engine) syncStatus(ctx context.Context, status protocol.PlaybackStatus) error {
	switch status {
	case protocol.PlaybackPlaying:
		return e.cfg.Player.Resume(ctx)
	case protocol.PlaybackPaused, protocol.PlaybackBuffering, protocol.PlaybackEnded:
		return e.cfg.Player.Pause(ctx)
	default:
		return nil
	}
}

func (e *Engine) ensurePlayer(ctx context.Context) error {
	if e.cfg.Player.Running() {
		return nil
	}
	return e.cfg.Player.Start(ctx)
}

// EffectiveServerMs returns the interpolated server playback position at now.
func EffectiveServerMs(view state.View, now time.Time) int64 {
	pb := view.Room.Playback
	tick := view.LastTick

	if pb.Status == protocol.PlaybackPlaying {
		if !tick.ServerTime.IsZero() && tick.Status == protocol.PlaybackPlaying {
			elapsed := now.Sub(tick.ServerTime).Milliseconds()
			if elapsed < 0 {
				elapsed = 0
			}
			return clampMs(tick.PositionMs+elapsed, pb.DurationMs)
		}
		if !pb.AnchorTime.IsZero() {
			elapsed := now.Sub(pb.AnchorTime).Milliseconds()
			if elapsed < 0 {
				elapsed = 0
			}
			return clampMs(pb.PositionMs+elapsed, pb.DurationMs)
		}
	}
	return pb.PositionMs
}

func clampMs(pos, durationMs int64) int64 {
	if pos < 0 {
		return 0
	}
	if durationMs > 0 && pos > durationMs {
		return durationMs
	}
	return pos
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// LastVideoID returns the video ID currently loaded in mpv (for tests).
func (e *Engine) LastVideoID() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastVideoID
}
