package room

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// NewQueueItem builds a queue entry from resolved track metadata (AC-027).
func NewQueueItem(track protocol.Track, addedBy string, now time.Time) protocol.QueueItem {
	return protocol.QueueItem{
		ID:         newQueueItemID(),
		VideoID:    track.VideoID,
		Title:      track.Title,
		DurationMs: track.DurationMs,
		AddedBy:    addedBy,
		AddedAt:    now,
	}
}

func newQueueItemID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// AddQueueItem appends an item to the end of the queue.
func (r *Room) AddQueueItem(item protocol.QueueItem) {
	r.Queue = append(r.Queue, item)
}

// RemoveQueueItem removes an item by ID (AC-030).
func (r *Room) RemoveQueueItem(itemID string) error {
	idx := r.queueIndex(itemID)
	if idx < 0 {
		return ErrQueueItemNotFound
	}
	r.Queue = append(r.Queue[:idx], r.Queue[idx+1:]...)
	return nil
}

// ReorderQueue moves itemID to immediately after afterID (AC-031).
// An empty afterID moves the item to the front of the queue.
func (r *Room) ReorderQueue(itemID, afterID string) error {
	from := r.queueIndex(itemID)
	if from < 0 {
		return ErrQueueItemNotFound
	}
	item := r.Queue[from]
	r.Queue = append(r.Queue[:from], r.Queue[from+1:]...)

	insertAt := 0
	if afterID != "" {
		after := r.queueIndex(afterID)
		if after < 0 {
			return ErrQueueItemNotFound
		}
		insertAt = after + 1
	}
	r.Queue = append(r.Queue[:insertAt], append([]protocol.QueueItem{item}, r.Queue[insertAt:]...)...)
	return nil
}

// AdvanceIfEnded starts the next queued track when the current one finishes (AC-029).
func (r *Room) AdvanceIfEnded(now time.Time) bool {
	if r.Playback.Status() != protocol.PlaybackPlaying || r.Playback.Track() == nil {
		return false
	}
	if !r.Playback.ReachedEnd(now) {
		return false
	}
	r.Skip(now)
	return true
}

// PromoteQueueItem moves an item to the front of the queue (AC-042).
func (r *Room) PromoteQueueItem(itemID string) error {
	idx := r.queueIndex(itemID)
	if idx < 0 {
		return ErrQueueItemNotFound
	}
	if idx == 0 {
		return nil
	}
	item := r.Queue[idx]
	r.Queue = append(r.Queue[:idx], r.Queue[idx+1:]...)
	r.Queue = append([]protocol.QueueItem{item}, r.Queue...)
	return nil
}

// QueueIndex returns the index of a queue item or -1.
func (r *Room) QueueIndex(itemID string) int {
	return r.queueIndex(itemID)
}

func (r *Room) queueIndex(itemID string) int {
	for i, item := range r.Queue {
		if item.ID == itemID {
			return i
		}
	}
	return -1
}

// IsHost reports whether sessionID is the room host.
func (r *Room) IsHost(sessionID string) bool {
	return r.HostSessionID == sessionID
}
