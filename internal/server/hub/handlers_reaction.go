package hub

import (
	"context"
	"errors"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

func (s *Server) handleReactionSend(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	var payload protocol.ReactionSendPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid reaction.send payload", nil)
		return
	}

	var counts map[string]int
	err = s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		if err := rm.AddReaction(payload.Emoji); err != nil {
			return err
		}
		counts = rm.ReactionCounts()
		return nil
	})
	if err != nil {
		s.sendReactionError(ctx, client.conn, env.ID, err)
		return
	}
	s.broadcastReactionUpdated(ctx, r.Slug, counts, env.ID)
}

func (s *Server) broadcastReactionUpdated(ctx context.Context, slug string, counts map[string]int, corrID string) {
	env, err := protocol.NewEnvelope(protocol.MsgReactionUpdated, corrID, protocol.ReactionUpdatedPayload{
		Counts: counts,
	})
	if err != nil {
		return
	}
	_ = s.broadcastToRoom(ctx, slug, env, "")
}

func (s *Server) sendReactionError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	switch {
	case errors.Is(err, room.ErrNoTrackPlaying):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "no track playing", nil)
	case room.IsEmptyEmojiError(err):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "emoji required", nil)
	case room.IsInvalidEmojiError(err):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "invalid emoji", nil)
	default:
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, err.Error(), nil)
	}
}
