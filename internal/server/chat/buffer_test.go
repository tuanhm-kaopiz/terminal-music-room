package chat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestBufferRingCapacity(t *testing.T) {
	b := NewBuffer(Options{Capacity: 3}, "room")
	now := time.Now()
	for i := 0; i < 5; i++ {
		body := string(rune('a' + i))
		if err := b.Add(UserMessage("u", body, now.Add(time.Duration(i)*time.Second))); err != nil {
			t.Fatal(err)
		}
	}
	msgs := b.Messages()
	if len(msgs) != 3 {
		t.Fatalf("len %d", len(msgs))
	}
	if msgs[0].Body != "c" || msgs[2].Body != "e" {
		t.Fatalf("msgs %+v", msgs)
	}
}

func TestValidateBody(t *testing.T) {
	if _, err := ValidateBody("   "); err != ErrEmptyMessage {
		t.Fatalf("got %v", err)
	}
	got, err := ValidateBody(" hello 🔥 ")
	if err != nil || got != "hello 🔥" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestStoreAppendAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir, "backend-team")
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	if err := store.Append(UserMessage("alice", "hi", now)); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(SystemMessage("bob joined", now.Add(time.Second))); err != nil {
		t.Fatal(err)
	}

	msgs, err := store.LoadRecent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0].Author != "alice" || msgs[1].Kind != protocol.ChatKindSystem {
		t.Fatalf("msgs %+v", msgs)
	}
	if _, err := os.Stat(filepath.Join(dir, "backend-team.chat.log")); err != nil {
		t.Fatal(err)
	}
}

func TestBufferPersistsAndReloads(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Capacity: 100, DataDir: dir}
	now := time.Now()

	b1 := NewBuffer(opts, "persist-room")
	if err := b1.Add(UserMessage("host", "saved", now)); err != nil {
		t.Fatal(err)
	}

	b2 := NewBuffer(opts, "persist-room")
	msgs := b2.Messages()
	if len(msgs) != 1 || msgs[0].Body != "saved" {
		t.Fatalf("reload %+v", msgs)
	}
}

func TestSanitizeFieldStripsControlChars(t *testing.T) {
	line := encodeLine(UserMessage("a", "line\nbreak\there", time.Now()))
	if stringsCount(line, "\n") != 0 {
		t.Fatalf("line should be single row: %q", line)
	}
	parts := splitTabs(line)
	if len(parts) != 4 || parts[3] != "line break here" {
		t.Fatalf("line %q parts %v", line, parts)
	}
}

func stringsCount(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}

func splitTabs(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
