package queuehistory

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestAppendURL(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	at := time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC)
	err := store.Append("my-room", Entry{
		At:      at,
		AddedBy: "alice",
		Source:  "https://www.youtube.com/watch?v=abc123xyz01",
		IsURL:   true,
		Track: protocol.Track{
			VideoID:    "abc123xyz01",
			Title:      "Neon Nights",
			DurationMs: 180_000,
			SourceURL:  "https://www.youtube.com/watch?v=abc123xyz01",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.Path("my-room"))
	if err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(string(data))
	if !strings.Contains(line, "\turl\thttps://www.youtube.com/watch?v=abc123xyz01\tabc123xyz01\tNeon Nights") {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestAppendQuery(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	err := store.Append("team", Entry{
		At:      time.Now(),
		AddedBy: "bob",
		Source:  "rain synthwave",
		IsURL:   false,
		Track: protocol.Track{
			VideoID: "vid999",
			Title:   "Rain Track",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.Path("team"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\tquery\train synthwave\tvid999\tRain Track\t") {
		t.Fatalf("unexpected: %q", string(data))
	}
}

func TestAppendNoOpWhenDisabled(t *testing.T) {
	var store *Store
	if err := store.Append("x", Entry{}); err != nil {
		t.Fatal(err)
	}
	store = NewStore("")
	if err := store.Append("x", Entry{}); err != nil {
		t.Fatal(err)
	}
}
