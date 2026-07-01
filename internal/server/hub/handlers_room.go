package hub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

func (s *Server) handleRoomCreate(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	if !s.AllowCreateRoom(sess.IP) {
		retry := s.limiter.RetryAfterSeconds(LimitCreateRoom, sess.IP)
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrRateLimited, "create room rate limited", &retry)
		return
	}
	if sess.RoomSlug != "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "leave current room first", nil)
		return
	}

	var payload protocol.RoomCreatePayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid room.create payload", nil)
		return
	}

	now := time.Now()
	member := protocol.Member{
		SessionID:   sess.ID,
		Nickname:    sess.Nickname,
		DisplayName: sess.DisplayName,
	}
	r, err := s.rooms.Create(payload.Slug, member, now)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	displayName := member.DisplayName
	if m, ok := r.FindMember(sess.ID); ok {
		sess.DisplayName = m.DisplayName
		displayName = m.DisplayName
	}
	sess.RoomSlug = r.Slug
	client.setRoom(r.Slug)

	_ = s.sendSnapshot(ctx, client.conn, r, now, env.ID)
	s.postSystemChat(ctx, r.Slug, fmt.Sprintf("%s joined the room", displayName), env.ID)
}

func (s *Server) handleRoomJoin(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	if sess.RoomSlug != "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "leave current room first", nil)
		return
	}

	var payload protocol.RoomJoinPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid room.join payload", nil)
		return
	}

	now := time.Now()
	member := protocol.Member{
		SessionID:   sess.ID,
		Nickname:    sess.Nickname,
		DisplayName: sess.DisplayName,
	}
	r, err := s.rooms.Join(payload.Slug, member, now)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	joined, ok := r.FindMember(sess.ID)
	if !ok {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "join failed", nil)
		return
	}
	sess.DisplayName = joined.DisplayName
	sess.RoomSlug = r.Slug
	client.setRoom(r.Slug)

	_ = s.sendSnapshot(ctx, client.conn, r, now, env.ID)
	s.postSystemChat(ctx, r.Slug, fmt.Sprintf("%s joined the room", joined.DisplayName), env.ID)
	joinEnv, err := protocol.NewEnvelope(protocol.MsgRoomMemberJoined, env.ID, protocol.RoomMemberJoinedPayload{
		Member: joined,
	})
	if err == nil {
		_ = s.broadcastToRoom(ctx, r.Slug, joinEnv, sess.ID)
	}
}

func (s *Server) handleRoomLeave(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	if sess.RoomSlug == "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrNotInRoom, "not in a room", nil)
		return
	}
	s.leaveRoom(ctx, client, sess, env.ID)
}

func (s *Server) leaveRoom(ctx context.Context, client *wsClient, sess *Session, corrID string) {
	slug := sess.RoomSlug
	displayName := sess.DisplayName
	if displayName == "" {
		displayName = sess.Nickname
	}
	res, err := s.rooms.Leave(slug, sess.ID)
	if err != nil {
		s.sendRoomError(ctx, client.conn, corrID, err)
		return
	}

	sess.RoomSlug = ""
	client.clearRoom()

	if res.Destroyed {
		return
	}

	s.postSystemChat(ctx, slug, fmt.Sprintf("%s left the room", displayName), corrID)

	leftEnv, err := protocol.NewEnvelope(protocol.MsgRoomMemberLeft, corrID, protocol.RoomMemberLeftPayload{
		SessionID: sess.ID,
	})
	if err == nil {
		_ = s.broadcastToRoom(ctx, slug, leftEnv, sess.ID)
	}

	if res.HostChanged {
		hostEnv, err := protocol.NewEnvelope(protocol.MsgRoomHostChanged, corrID, protocol.RoomHostChangedPayload{
			HostSessionID: res.Room.HostSessionID,
		})
		if err == nil {
			_ = s.broadcastToRoom(ctx, slug, hostEnv, "")
		}
	}
}

func (s *Server) sendSnapshot(ctx context.Context, conn *websocket.Conn, r *room.Room, now time.Time, corrID string) error {
	snap := r.Snapshot(now)
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, corrID, snap)
	if err != nil {
		return err
	}
	return s.writeEnvelope(ctx, conn, env)
}

func (s *Server) broadcastToRoom(ctx context.Context, slug string, env protocol.Envelope, excludeSession string) error {
	r, ok := s.rooms.Get(slug)
	if !ok {
		return nil
	}
	for _, m := range r.Members {
		if m.SessionID == excludeSession {
			continue
		}
		if c, ok := s.clients.get(m.SessionID); ok {
			_ = s.writeEnvelope(ctx, c.conn, env)
		}
	}
	return nil
}

func (s *Server) sendRoomError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	code := roomErrorCode(err)
	msg := err.Error()
	if v, ok := err.(room.ValidationError); ok {
		msg = v.Error()
	}
	_ = s.sendError(ctx, conn, corrID, code, msg, nil)
}

func roomErrorCode(err error) protocol.ErrorCode {
	switch {
	case errors.Is(err, room.ErrSlugTaken):
		return protocol.ErrSlugTaken
	case errors.Is(err, room.ErrRoomNotFound):
		return protocol.ErrRoomNotFound
	case errors.Is(err, room.ErrRoomFull):
		return protocol.ErrRoomFull
	case errors.Is(err, room.ErrNotInRoom):
		return protocol.ErrNotInRoom
	case errors.Is(err, room.ErrAlreadyMember):
		return protocol.ErrInvalidMessage
	case errors.Is(err, room.ErrForbidden):
		return protocol.ErrForbidden
	case errors.Is(err, room.ErrQueueItemNotFound):
		return protocol.ErrInvalidMessage
	default:
		var v room.ValidationError
		if errors.As(err, &v) {
			if v.Field == "slug" {
				return protocol.ErrInvalidSlug
			}
		}
		return protocol.ErrInvalidMessage
	}
}

func (s *Server) dispatchMessage(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	switch env.Type {
	case protocol.MsgRoomCreate:
		s.handleRoomCreate(ctx, client, sess, env)
	case protocol.MsgRoomJoin:
		s.handleRoomJoin(ctx, client, sess, env)
	case protocol.MsgRoomLeave:
		s.handleRoomLeave(ctx, client, sess, env)
	case protocol.MsgPlaybackPlay:
		s.handlePlaybackPlay(ctx, client, sess, env)
	case protocol.MsgPlaybackPause:
		s.handlePlaybackPause(ctx, client, sess, env)
	case protocol.MsgPlaybackResume:
		s.handlePlaybackResume(ctx, client, sess, env)
	case protocol.MsgPlaybackSkip:
		s.handlePlaybackSkip(ctx, client, sess, env)
	case protocol.MsgPlaybackSeek:
		s.handlePlaybackSeek(ctx, client, sess, env)
	case protocol.MsgQueueAdd:
		s.handleQueueAdd(ctx, client, sess, env)
	case protocol.MsgQueueRemove:
		s.handleQueueRemove(ctx, client, sess, env)
	case protocol.MsgQueueReorder:
		s.handleQueueReorder(ctx, client, sess, env)
	case protocol.MsgChatSend:
		s.handleChatSend(ctx, client, sess, env)
	case protocol.MsgVoteSkip:
		s.handleVoteSkip(ctx, client, sess, env)
	case protocol.MsgVotePriority:
		s.handleVotePriority(ctx, client, sess, env)
	case protocol.MsgReactionSend:
		s.handleReactionSend(ctx, client, sess, env)
	default:
		s.log.Debug("unhandled message", "type", env.Type, "session_id", sess.ID)
	}
}
