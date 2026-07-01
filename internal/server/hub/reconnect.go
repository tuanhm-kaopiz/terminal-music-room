package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

const defaultReconnectWindow = 5 * time.Minute

func (s *Server) reconnectWindow() time.Duration {
	if s.reconnectTTL > 0 {
		return s.reconnectTTL
	}
	return defaultReconnectWindow
}

func (s *Server) markSessionDisconnected(id string, now time.Time) {
	s.sessions.Disconnect(id, now)
}

func (s *Server) restoreRoomOnReconnect(ctx context.Context, client *wsClient, conn *websocket.Conn, sess *Session) error {
	if sess.RoomSlug == "" {
		return nil
	}
	r, ok := s.rooms.Get(sess.RoomSlug)
	if !ok {
		sess.RoomSlug = ""
		return nil
	}
	if _, ok := r.FindMember(sess.ID); !ok {
		sess.RoomSlug = ""
		return nil
	}
	client.setRoom(sess.RoomSlug)
	return s.sendSnapshot(ctx, conn, r, time.Now(), "")
}

func (s *Server) expireSessionRoomMembership(ctx context.Context, sess *Session, corrID string) {
	if sess.RoomSlug == "" {
		return
	}
	slug := sess.RoomSlug
	displayName := sess.DisplayName
	if displayName == "" {
		displayName = sess.Nickname
	}
	res, err := s.rooms.Leave(slug, sess.ID)
	sess.RoomSlug = ""
	if err != nil {
		return
	}
	if res.Destroyed {
		return
	}
	s.postSystemChat(ctx, slug, displayName+" left the room", corrID)
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

func (s *Server) resumeSession(id, ip string, now time.Time) (*Session, error) {
	sess, ok := s.sessions.Get(id)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if sess.DisconnectedAt != nil && now.Sub(*sess.DisconnectedAt) > s.reconnectWindow() {
		s.expireSessionRoomMembership(context.Background(), sess, "")
	}
	sess.IP = ip
	sess.LastSeen = now
	sess.DisconnectedAt = nil
	s.sessions.Register(sess)
	return sess, nil
}
