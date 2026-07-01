package actions

import (
	"context"
	"fmt"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Play requests playback of a YouTube URL or search query.
func (r *Room) Play(ctx context.Context, urlOrQuery string) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	payload, err := playPayloadFromSource(urlOrQuery)
	if err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgPlaybackPlay, payload)
}

// Pause requests a room-wide pause.
func (r *Room) Pause(ctx context.Context) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgPlaybackPause, protocol.PlaybackPausePayload{})
}

// Resume requests a room-wide resume.
func (r *Room) Resume(ctx context.Context) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgPlaybackResume, protocol.PlaybackResumePayload{})
}

// Skip requests skipping to the next queue item.
func (r *Room) Skip(ctx context.Context) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgPlaybackSkip, protocol.PlaybackSkipPayload{})
}

// Seek requests a seek to positionMs (milliseconds).
func (r *Room) Seek(ctx context.Context, positionMs int64) error {
	if positionMs < 0 {
		return fmt.Errorf("invalid position %d — use milliseconds >= 0", positionMs)
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgPlaybackSeek, protocol.PlaybackSeekPayload{PositionMs: positionMs})
}

// SeekFromString parses a millisecond position and seeks.
func (r *Room) SeekFromString(ctx context.Context, position string) error {
	var ms int64
	if _, err := fmt.Sscan(position, &ms); err != nil || ms < 0 {
		return fmt.Errorf("invalid position %q — use milliseconds", position)
	}
	return r.Seek(ctx, ms)
}
