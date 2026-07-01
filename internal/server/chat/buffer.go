package chat

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Options configures an in-memory chat buffer and optional disk log.
type Options struct {
	Capacity int
	DataDir  string
}

// Buffer is a fixed-size ring of recent chat messages with optional append-only log.
type Buffer struct {
	capacity int
	slug     string
	messages []protocol.ChatMessage
	store    *Store
}

// NewBuffer creates a buffer, loading recent lines from disk when DataDir is set.
func NewBuffer(opts Options, slug string) *Buffer {
	capacity := opts.Capacity
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	b := &Buffer{
		capacity: capacity,
		slug:     slug,
	}
	if opts.DataDir != "" {
		b.store = NewStore(opts.DataDir, slug)
		if msgs, err := b.store.LoadRecent(capacity); err == nil {
			b.messages = msgs
		}
	}
	return b
}

// Add appends a message to the ring buffer and persists it when configured.
func (b *Buffer) Add(msg protocol.ChatMessage) error {
	if msg.ID == "" {
		msg.ID = newMessageID()
	}
	b.messages = append(b.messages, msg)
	if len(b.messages) > b.capacity {
		b.messages = b.messages[len(b.messages)-b.capacity:]
	}
	if b.store != nil {
		return b.store.Append(msg)
	}
	return nil
}

// Messages returns a copy of buffered messages oldest-first.
func (b *Buffer) Messages() []protocol.ChatMessage {
	out := make([]protocol.ChatMessage, len(b.messages))
	copy(out, b.messages)
	return out
}

// ValidateBody trims and rejects empty chat input (AC-036).
func ValidateBody(raw string) (string, error) {
	body := strings.TrimSpace(raw)
	if body == "" {
		return "", ErrEmptyMessage
	}
	if utf8.RuneCountInString(body) > maxBodyRunes {
		return "", ErrBodyTooLong
	}
	return body, nil
}

// UserMessage builds a user chat line (AC-033, AC-034).
func UserMessage(author, body string, now time.Time) protocol.ChatMessage {
	return protocol.ChatMessage{
		ID:     newMessageID(),
		Kind:   protocol.ChatKindUser,
		Author: author,
		Body:   body,
		At:     now.UTC(),
	}
}

// SystemMessage builds a system chat line (AC-035).
func SystemMessage(body string, now time.Time) protocol.ChatMessage {
	return protocol.ChatMessage{
		ID:     newMessageID(),
		Kind:   protocol.ChatKindSystem,
		Body:   body,
		At:     now.UTC(),
	}
}

func newMessageID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
