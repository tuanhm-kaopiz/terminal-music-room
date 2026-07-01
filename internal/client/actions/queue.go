package actions

import (
	"context"
	"fmt"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// QueueAdd appends a track from a URL or search query.
func (r *Room) QueueAdd(ctx context.Context, urlOrQuery string) error {
	if err := r.requireInRoom(); err != nil {
		return err
	}
	payload, err := queueAddPayloadFromSource(urlOrQuery)
	if err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgQueueAdd, payload)
}

// QueueRemove removes a queue item (host only; enforced server-side).
func (r *Room) QueueRemove(ctx context.Context, itemID string) error {
	if itemID == "" {
		return fmt.Errorf("queue item id is required")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgQueueRemove, protocol.QueueRemovePayload{ItemID: itemID})
}

// QueueReorder moves itemID after afterID (host only; enforced server-side).
func (r *Room) QueueReorder(ctx context.Context, itemID, afterID string) error {
	if itemID == "" {
		return fmt.Errorf("queue item id is required")
	}
	if err := r.requireInRoom(); err != nil {
		return err
	}
	return r.send(ctx, protocol.MsgQueueReorder, protocol.QueueReorderPayload{
		ItemID:  itemID,
		AfterID: afterID,
	})
}
