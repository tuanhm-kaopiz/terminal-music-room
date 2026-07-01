package hub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

func (s *Server) handleChatSend(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	if !s.AllowChat(sess.IP) {
		retry := s.limiter.RetryAfterSeconds(LimitChat, sess.IP)
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrRateLimited, "chat rate limited", &retry)
		return
	}

	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	var payload protocol.ChatSendPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid chat.send payload", nil)
		return
	}

	body, err := chat.ValidateBody(payload.Body)
	if err != nil {
		s.sendChatError(ctx, client.conn, env.ID, err)
		return
	}

	author := sess.DisplayName
	if author == "" {
		author = sess.Nickname
	}
	msg := chat.UserMessage(author, body, time.Now())
	s.appendAndBroadcastChat(ctx, r.Slug, msg, env.ID)
}

func (s *Server) appendAndBroadcastChat(ctx context.Context, slug string, msg protocol.ChatMessage, corrID string) {
	if err := s.rooms.Modify(slug, func(rm *room.Room) error {
		return rm.Chat.Add(msg)
	}); err != nil {
		return
	}
	env, err := protocol.NewEnvelope(protocol.MsgChatMessage, corrID, msg)
	if err != nil {
		return
	}
	_ = s.broadcastToRoom(ctx, slug, env, "")
}

func (s *Server) postSystemChat(ctx context.Context, slug, body, corrID string) {
	msg := chat.SystemMessage(body, time.Now())
	s.appendAndBroadcastChat(ctx, slug, msg, corrID)
}

func (s *Server) postNowPlayingChat(ctx context.Context, slug, corrID string) {
	r, ok := s.rooms.Get(slug)
	if !ok {
		return
	}
	track := r.Playback.Track()
	if track == nil || r.Playback.Status() != protocol.PlaybackPlaying {
		return
	}
	s.postSystemChat(ctx, slug, fmt.Sprintf("now playing: %s", track.Title), corrID)
}

func (s *Server) sendChatError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	switch {
	case errors.Is(err, chat.ErrEmptyMessage):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "message must not be empty", nil)
	case errors.Is(err, chat.ErrBodyTooLong):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "message too long", nil)
	default:
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, err.Error(), nil)
	}
}
