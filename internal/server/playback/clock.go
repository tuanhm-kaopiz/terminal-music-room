package playback

import (
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

const maxDurationMs = int64((24 * time.Hour) / time.Millisecond)

// Clock is the server-authoritative playback state machine.
type Clock struct {
	status     protocol.PlaybackStatus
	track      *protocol.Track
	positionMs int64
	anchorTime time.Time
	durationMs int64
}

// NewClock returns a clock in ended state with no track.
func NewClock() *Clock {
	return &Clock{status: protocol.PlaybackEnded}
}

// State returns a snapshot suitable for protocol messages.
func (c *Clock) State() protocol.PlaybackState {
	var track *protocol.Track
	if c.track != nil {
		t := *c.track
		track = &t
	}
	return protocol.PlaybackState{
		Status:     c.status,
		Track:      track,
		PositionMs: c.positionMs,
		AnchorTime: c.anchorTime,
		DurationMs: c.durationMs,
	}
}

// EffectivePositionMs returns the interpolated playback position at now.
func (c *Clock) EffectivePositionMs(now time.Time) int64 {
	if c.status != protocol.PlaybackPlaying || c.anchorTime.IsZero() {
		return c.positionMs
	}
	elapsed := now.Sub(c.anchorTime).Milliseconds()
	if elapsed < 0 {
		return c.positionMs
	}
	return clampPosition(c.positionMs+elapsed, c.durationMs)
}

// LoadTrack sets the current track and resets position without starting playback.
func (c *Clock) LoadTrack(track protocol.Track, now time.Time) {
	t := track
	c.track = &t
	c.durationMs = track.DurationMs
	c.positionMs = 0
	c.anchorTime = time.Time{}
	c.status = protocol.PlaybackPaused
	_ = now
}

// Play starts or resumes playback from the current position.
func (c *Clock) Play(now time.Time) {
	if c.track == nil {
		return
	}
	if c.status == protocol.PlaybackPlaying {
		return
	}
	if c.status == protocol.PlaybackEnded {
		c.positionMs = 0
	}
	c.status = protocol.PlaybackPlaying
	c.anchorTime = now
}

// Pause freezes playback at the effective position.
func (c *Clock) Pause(now time.Time) {
	if c.status != protocol.PlaybackPlaying {
		return
	}
	c.positionMs = c.EffectivePositionMs(now)
	c.status = protocol.PlaybackPaused
	c.anchorTime = time.Time{}
}

// SeekTo moves to positionMs and keeps play/pause status (paused stays paused).
func (c *Clock) SeekTo(positionMs int64, now time.Time) {
	if c.track == nil {
		return
	}
	wasPlaying := c.status == protocol.PlaybackPlaying
	c.positionMs = clampPosition(positionMs, c.durationMs)
	if wasPlaying {
		c.status = protocol.PlaybackPlaying
		c.anchorTime = now
	} else {
		c.status = protocol.PlaybackPaused
		c.anchorTime = time.Time{}
	}
}

// SetBuffering marks the clock as buffering without changing position.
func (c *Clock) SetBuffering(now time.Time) {
	if c.track == nil {
		return
	}
	if c.status == protocol.PlaybackPlaying {
		c.positionMs = c.EffectivePositionMs(now)
		c.anchorTime = time.Time{}
	}
	c.status = protocol.PlaybackBuffering
}

// ReachedEnd reports whether playing position reached track duration.
func (c *Clock) ReachedEnd(now time.Time) bool {
	if c.status != protocol.PlaybackPlaying || c.track == nil || c.durationMs <= 0 {
		return false
	}
	return c.EffectivePositionMs(now) >= c.durationMs
}

// MarkEnded sets playback to ended at the track duration.
func (c *Clock) MarkEnded() {
	if c.durationMs > 0 {
		c.positionMs = c.durationMs
	}
	c.status = protocol.PlaybackEnded
	c.anchorTime = time.Time{}
}

// Clear removes the current track and resets the clock.
func (c *Clock) Clear() {
	*c = *NewClock()
}

// Status returns the current playback status.
func (c *Clock) Status() protocol.PlaybackStatus {
	return c.status
}

// Track returns the loaded track or nil.
func (c *Clock) Track() *protocol.Track {
	return c.track
}

func clampPosition(pos, durationMs int64) int64 {
	if pos < 0 {
		return 0
	}
	if durationMs > 0 && pos > durationMs {
		return durationMs
	}
	if pos > maxDurationMs {
		return maxDurationMs
	}
	return pos
}
