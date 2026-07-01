package hub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
	"github.com/terminal-music-room/music-room/internal/server/vote"
)

func (s *Server) handleVoteSkip(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	now := time.Now()
	var result vote.Result
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		var castErr error
		result, castErr = vote.CastSkip(rm, sess.ID, now, s.voteCfg)
		return castErr
	}); err != nil {
		s.sendVoteError(ctx, client.conn, env.ID, err)
		return
	}
	s.applyVoteResult(ctx, r.Slug, env.ID, result)
}

func (s *Server) handleVotePriority(ctx context.Context, client *wsClient, sess *Session, env protocol.Envelope) {
	r, err := s.roomForSession(sess)
	if err != nil {
		s.sendRoomError(ctx, client.conn, env.ID, err)
		return
	}

	var payload protocol.VotePriorityPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "invalid vote.priority payload", nil)
		return
	}
	if payload.ItemID == "" {
		_ = s.sendError(ctx, client.conn, env.ID, protocol.ErrInvalidMessage, "item_id required", nil)
		return
	}

	now := time.Now()
	var result vote.Result
	if err := s.rooms.Modify(r.Slug, func(rm *room.Room) error {
		var castErr error
		result, castErr = vote.CastPriority(rm, sess.ID, payload.ItemID, now, s.voteCfg)
		return castErr
	}); err != nil {
		s.sendVoteError(ctx, client.conn, env.ID, err)
		return
	}
	s.applyVoteResult(ctx, r.Slug, env.ID, result)
}

func (s *Server) applyVoteResult(ctx context.Context, slug, corrID string, result vote.Result) {
	if result.Cancelled {
		s.broadcastVote(ctx, slug, corrID, nil, nil)
		s.postSystemChat(ctx, slug, result.CancelMsg, corrID)
		return
	}

	var voteCopy *protocol.Vote
	if result.Vote != nil {
		v := *result.Vote
		voteCopy = &v
	}
	progress := result.Progress
	s.broadcastVote(ctx, slug, corrID, voteCopy, &progress)

	if result.Vote != nil && len(result.Vote.Voters) == 1 {
		s.postVoteStartedChat(ctx, slug, corrID, result.Vote)
	}

	if !result.Passed {
		return
	}

	switch result.PassedKind {
	case protocol.VoteKindSkip:
		s.publishPlayback(ctx, slug, corrID)
		s.broadcastQueue(ctx, slug, corrID)
		s.postNowPlayingChat(ctx, slug, corrID)
	case protocol.VoteKindPriority:
		s.broadcastQueue(ctx, slug, corrID)
		s.postSystemChat(ctx, slug, "priority vote passed", corrID)
	}
}

func (s *Server) postVoteStartedChat(ctx context.Context, slug, corrID string, v *protocol.Vote) {
	if v == nil {
		return
	}
	switch v.Kind {
	case protocol.VoteKindSkip:
		s.postSystemChat(ctx, slug, fmt.Sprintf("skip vote started (%d/%d needed)", len(v.Voters), v.Threshold), corrID)
	case protocol.VoteKindPriority:
		title := v.TargetID
		if r, ok := s.rooms.Get(slug); ok {
			if idx := r.QueueIndex(v.TargetID); idx >= 0 {
				title = r.Queue[idx].Title
			}
		}
		s.postSystemChat(ctx, slug, fmt.Sprintf("priority vote started for %q (%d/%d needed)", title, len(v.Voters), v.Threshold), corrID)
	}
}

func (s *Server) broadcastVote(ctx context.Context, slug, corrID string, v *protocol.Vote, progress *protocol.VoteProgress) {
	env, err := protocol.NewEnvelope(protocol.MsgVoteUpdated, corrID, protocol.VoteUpdatedPayload{
		Vote:     v,
		Progress: progress,
	})
	if err != nil {
		return
	}
	_ = s.broadcastToRoom(ctx, slug, env, "")
}

func (s *Server) checkRoomVotes(ctx context.Context, slug, corrID string) {
	var result vote.Result
	var changed bool
	if err := s.rooms.Modify(slug, func(rm *room.Room) error {
		result, changed = vote.CheckRoomVotes(rm, time.Now(), s.voteCfg)
		return nil
	}); err != nil || !changed {
		return
	}
	s.applyVoteResult(ctx, slug, corrID, result)
}

func (s *Server) clearVoteOnTrackChange(ctx context.Context, slug, corrID string) {
	var result vote.Result
	var changed bool
	if err := s.rooms.Modify(slug, func(rm *room.Room) error {
		result, changed = vote.ClearOnTrackChange(rm)
		return nil
	}); err != nil || !changed {
		return
	}
	s.applyVoteResult(ctx, slug, corrID, result)
}

func (s *Server) sendVoteError(ctx context.Context, conn *websocket.Conn, corrID string, err error) {
	switch {
	case errors.Is(err, vote.ErrNoTrackPlaying):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrNoTrackPlaying, "no track playing", nil)
	case errors.Is(err, vote.ErrQueueTooShort):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "queue too short for priority vote", nil)
	case errors.Is(err, vote.ErrInvalidTarget):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "queue item not found", nil)
	case errors.Is(err, vote.ErrVoteKindConflict):
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, "another vote is in progress", nil)
	default:
		_ = s.sendError(ctx, conn, corrID, protocol.ErrInvalidMessage, err.Error(), nil)
	}
}
