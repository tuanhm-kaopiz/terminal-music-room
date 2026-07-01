package hub

import (
	"context"
	"errors"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
	"github.com/terminal-music-room/music-room/internal/server/vote"
	"github.com/terminal-music-room/music-room/internal/server/youtube"
)

func (s *Server) handlePlaybackPlay(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	var payload protocol.PlaybackPlayPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid playback.play payload", nil)
		return
	}
	if payload.URL == "" && payload.Query == "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidSource, "url or query required", nil)
		return
	}
	if payload.URL != "" && payload.Query != "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidSource, "provide url or query, not both", nil)
		return
	}
	if payload.URL != "" {
		if _, ok := youtube.ParseVideoID(payload.URL); !ok {
			_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidSource, "invalid source", nil)
			return
		}
	}

	go s.resolveAndPlay(client, sess, r.Slug, env.ID, payload.URL, payload.Query)
}

func (s *Server) resolveAndPlay(client *wsClient, sess *Session, slug, corrID, rawURL, query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if sess.RoomSlug != slug {
		return
	}

	track, err := s.resolver.Resolve(ctx, rawURL, query)
	if err != nil {
		s.sendPlaybackResolveError(context.Background(), client.conn, corrID, err)
		return
	}

	now := time.Now()
	if err := s.rooms.Modify(slug, func(rm *room.Room) error {
		cur := rm.Playback.Track()
		if cur != nil && cur.VideoID != track.VideoID {
			vote.ClearOnTrackChange(rm)
		}
		rm.ResetReactions()
		rm.Playback.LoadTrack(track, now)
		rm.Playback.Play(now)
		return nil
	}); err != nil {
		s.sendRoomError(context.Background(), client.conn, corrID, err)
		return
	}
	s.publishPlayback(context.Background(), slug, corrID)
	s.postNowPlayingChat(context.Background(), slug, corrID)
}

func (s *Server) handlePlaybackPause(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if r.Playback.Track() == nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrNoTrackPlaying, "no track loaded", nil)
		return
	}

	now := time.Now()
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		rm.Playback.Pause(now)
		return nil
	}); err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	s.publishPlayback(ctx, r.Slug, env.ID)
}

func (s *Server) handlePlaybackResume(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if r.Playback.Track() == nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrNoTrackPlaying, "no track loaded", nil)
		return
	}

	now := time.Now()
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		rm.Playback.Play(now)
		return nil
	}); err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	s.publishPlayback(ctx, r.Slug, env.ID)
}

func (s *Server) handlePlaybackSkip(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if r.Playback.Track() == nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrNoTrackPlaying, "no track loaded", nil)
		return
	}

	now := time.Now()
	s.clearVoteOnTrackChange(ctx, r.Slug, env.ID)
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		rm.Skip(now)
		return nil
	}); err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	s.publishPlayback(ctx, r.Slug, env.ID)
	s.postNowPlayingChat(ctx, r.Slug, env.ID)
}

func (s *Server) handlePlaybackSeek(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if r.Playback.Track() == nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrNoTrackPlaying, "no track loaded", nil)
		return
	}

	var payload protocol.PlaybackSeekPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid playback.seek payload", nil)
		return
	}

	now := time.Now()
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		rm.Playback.SeekTo(payload.PositionMs, now)
		return nil
	}); err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	s.publishPlayback(ctx, r.Slug, env.ID)
}

func (s *Server) roomForSession(sess *Session) (*room.Room, error) {
	if sess.RoomSlug == "" {
		return nil, room.ErrNotInRoom
	}
	r, ok := s.rooms.Get(sess.RoomSlug)
	if !ok {
		return nil, room.ErrRoomNotFound
	}
	return r, nil
}

func (s *Server) publishPlayback(ctx context.Context, slug, corrID string) {
	r, ok := s.rooms.Get(slug)
	if !ok {
		return
	}
	now := time.Now()
	state := r.Playback.State()
	if r.Playback.Status() == protocol.PlaybackPlaying {
		state.PositionMs = r.Playback.EffectivePositionMs(now)
	}

	stateEnv, err := protocol.NewEnvelope(protocol.MsgPlaybackState, corrID, state)
	if err == nil {
		_ = s.broadcastToRoom(ctx, slug, stateEnv, "")
	}
	s.broadcastTick(ctx, slug, r, now)
}

func (s *Server) broadcastTick(ctx context.Context, slug string, r *room.Room, now time.Time) {
	pos := r.Playback.State().PositionMs
	if r.Playback.Status() == protocol.PlaybackPlaying {
		pos = r.Playback.EffectivePositionMs(now)
	}
	tickEnv, err := protocol.NewEnvelope(protocol.MsgPlaybackTick, "", protocol.PlaybackTickPayload{
		PositionMs: pos,
		Status:     r.Playback.Status(),
		ServerTime: now,
	})
	if err != nil {
		return
	}
	_ = s.broadcastToRoom(ctx, slug, tickEnv, "")
}

func (s *Server) sendPlaybackResolveError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	switch {
	case errors.Is(err, youtube.ErrInvalidSource):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidSource, "invalid source", nil)
	case errors.Is(err, youtube.ErrSourceUnavailable):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrSourceUnavailable, err.Error(), nil)
	default:
		_ = s.sendError(ctx, conn, corrID, protocol.ErrSourceUnavailable, err.Error(), nil)
	}
}

func (s *Server) startPlaybackTicks() {
	s.tickOnce.Do(func() {
		go s.runPlaybackTicks()
	})
}

func (s *Server) runPlaybackTicks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.tickStop:
			return
		case now := <-ticker.C:
			ctx := context.Background()
			for _, slug := range s.rooms.Slugs() {
				s.checkRoomVotes(ctx, slug, "")
			}
			for _, slug := range s.rooms.PlayingSlugs() {
				var advanced bool
				if err := s.rooms.Modify(slug, func(rm *room.Room) error {
					if rm.AdvanceIfEnded(now) {
						advanced = true
					}
					return nil
				}); err != nil {
					continue
				}
				if advanced {
					s.clearVoteOnTrackChange(ctx, slug, "")
					s.publishPlayback(ctx, slug, "")
					s.broadcastQueue(ctx, slug, "")
					s.postNowPlayingChat(ctx, slug, "")
					continue
				}
				r, ok := s.rooms.Get(slug)
				if ok {
					s.broadcastTick(ctx, slug, r, now)
				}
			}
		}
	}
}
