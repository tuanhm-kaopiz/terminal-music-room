package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Chat sends a text chat message.
func (r *Room) Chat(ctx context.Context, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("chat message cannot be empty")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgChatSend, protocol.ChatSendPayload{Body: body})
}

// VoteSkip casts a skip vote on the current track.
func (r *Room) VoteSkip(ctx context.Context) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgVoteSkip, protocol.VoteSkipPayload{})
}

// VotePriority casts a priority vote for a queue item.
func (r *Room) VotePriority(ctx context.Context, itemID string) error {
	if itemID == "" {
		return fmt.Errorf("queue item id is required")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgVotePriority, protocol.VotePriorityPayload{ItemID: itemID})
}

// React sends an emoji reaction on the current track.
func (r *Room) React(ctx context.Context, emoji string) error {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return fmt.Errorf("emoji is required")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgReactionSend, protocol.ReactionSendPayload{Emoji: emoji})
}

// Leave requests leaving the current room (caller handles local teardown).
func (r *Room) Leave(ctx context.Context) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgRoomLeave, protocol.RoomLeavePayload{})
}

// Create sends room.create with an optional password.
func (r *Room) Create(ctx context.Context, slug, password string) error {
	if r.Store != nil && r.Store.Snapshot().InRoom {
		return errors.New("leave current room first")
	}
	return r.send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: slug, Password: password})
}

// Join sends room.join with an optional password.
func (r *Room) Join(ctx context.Context, slug, password string) error {
	if r.Store != nil && r.Store.Snapshot().InRoom {
		return errors.New("leave current room first")
	}
	return r.send(ctx, protocol.MsgRoomJoin, protocol.RoomJoinPayload{Slug: slug, Password: password})
}

// Kick asks the host to remove a member from the room.
func (r *Room) Kick(ctx context.Context, targetSessionID string) error {
	if targetSessionID == "" {
		return errors.New("target session id is required")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgRoomKick, protocol.RoomKickPayload{TargetSessionID: targetSessionID})
}
