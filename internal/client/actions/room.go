package actions

import (
	"context"
	"errors"

	"github.com/terminal-music-room/music-room/internal/client/state"
)

// ErrNotInRoom is returned when a room action requires an active room membership.
var ErrNotInRoom = errors.New("not in a room — run: music-room join <slug>")

// Sender dispatches a protocol message to the sync server.
type Sender func(ctx context.Context, msgType string, payload any) error

// Room performs validated room operations via WebSocket.
type Room struct {
	Send  Sender
	Store *state.Store
}

// New returns a Room bound to send and store.
func New(send Sender, store *state.Store) *Room {
	return &Room{Send: send, Store: store}
}

func (r *Room) requireInRoom() error {
	if r.Store == nil || !r.Store.Snapshot().InRoom {
		return ErrNotInRoom
	}
	return nil
}

func (r *Room) send(ctx context.Context, msgType string, payload any) error {
	if r.Send == nil {
		return errors.New("not connected")
	}
	return r.Send(ctx, msgType, payload)
}
