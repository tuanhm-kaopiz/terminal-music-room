package hub

import (
	"context"
	"errors"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
	"github.com/terminal-music-room/music-room/internal/server/youtube"
)

func (s *Server) handleQueueAdd(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	var payload protocol.QueueAddPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid queue.add payload", nil)
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

	go s.resolveAndAddQueue(client, sess, r.Slug, env.ID, payload.URL, payload.Query)
}

func (s *Server) resolveAndAddQueue(client *wsClient, sess *Session, slug, corrID, rawURL, query string) {
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
	addedBy := sess.DisplayName
	if addedBy == "" {
		addedBy = sess.Nickname
	}
	item := room.NewQueueItem(track, addedBy, now)
	if err := s.rooms.Modify(slug, func(rm *room.Room) error {
		rm.AddQueueItem(item)
		return nil
	}); err != nil {
		s.sendRoomError(context.Background(), client.conn, corrID, err)
		return
	}
	s.broadcastQueue(context.Background(), slug, corrID)
}

func (s *Server) handleQueueRemove(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if !r.IsHost(sess.ID) {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrForbidden, "host only", nil)
		return
	}

	var payload protocol.QueueRemovePayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid queue.remove payload", nil)
		return
	}
	if payload.ItemID == "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "item_id required", nil)
		return
	}

	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		return rm.RemoveQueueItem(payload.ItemID)
	}); err != nil {
		s.sendQueueError(ctx, client.conn, env.ID, err)
		return
	}
	s.checkRoomVotes(ctx, r.Slug, env.ID)
	s.broadcastQueue(ctx, r.Slug, env.ID)
}

func (s *Server) handleQueueReorder(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}
	if !r.IsHost(sess.ID) {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrForbidden, "host only", nil)
		return
	}

	var payload protocol.QueueReorderPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid queue.reorder payload", nil)
		return
	}
	if payload.ItemID == "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "item_id required", nil)
		return
	}

	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		return rm.ReorderQueue(payload.ItemID, payload.AfterID)
	}); err != nil {
		s.sendQueueError(ctx, client.conn, env.ID, err)
		return
	}
	s.broadcastQueue(ctx, r.Slug, env.ID)
}

func (s *Server) broadcastQueue(ctx context.Context, slug, corrID string) {
	r, ok := s.rooms.Get(slug)
	if !ok {
		return
	}
	items := make([]protocol.QueueItem, len(r.Queue))
	copy(items, r.Queue)
	env, err := protocol.NewEnvelope(protocol.MsgQueueUpdated, corrID, protocol.QueueUpdatedPayload{Items: items})
	if err != nil {
		return
	}
	_ = s.broadcastToRoom(ctx, slug, env, "")
}

func (s *Server) sendQueueError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	switch {
	case errors.Is(err, room.ErrQueueItemNotFound):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "queue item not found", nil)
	case errors.Is(err, room.ErrForbidden):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrForbidden, "forbidden", nil)
	default:
		s.sendRoomError(ctx, conn, corrID, err)
	}
}
