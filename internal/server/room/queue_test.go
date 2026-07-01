package room

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

func TestQueueAddRemoveReorder(t *testing.T) {
	now := time.Now()
	r := NewRoom("q", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})

	a := protocol.QueueItem{ID: "a", VideoID: "v1", Title: "A", AddedBy: "host", AddedAt: now}
	b := protocol.QueueItem{ID: "b", VideoID: "v2", Title: "B", AddedBy: "host", AddedAt: now}
	c := protocol.QueueItem{ID: "c", VideoID: "v3", Title: "C", AddedBy: "guest", AddedAt: now}
	r.AddQueueItem(a)
	r.AddQueueItem(b)
	r.AddQueueItem(c)

	if len(r.Queue) != 3 || r.Queue[2].ID != "c" {
		t.Fatalf("queue %+v", r.Queue)
	}

	if err := r.RemoveQueueItem("b"); err != nil {
		t.Fatal(err)
	}
	if len(r.Queue) != 2 || r.Queue[0].ID != "a" || r.Queue[1].ID != "c" {
		t.Fatalf("after remove %+v", r.Queue)
	}

	if err := r.ReorderQueue("c", "a"); err != nil {
		t.Fatal(err)
	}
	if r.Queue[0].ID != "a" || r.Queue[1].ID != "c" {
		t.Fatalf("after reorder %+v", r.Queue)
	}
}

func TestQueueRemoveNotFound(t *testing.T) {
	r := NewRoom("q", protocol.Member{SessionID: "h", Nickname: "host"}, time.Now(), chat.Options{})
	if err := r.RemoveQueueItem("missing"); err != ErrQueueItemNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestQueueReorderToFront(t *testing.T) {
	now := time.Now()
	r := NewRoom("q", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	r.AddQueueItem(protocol.QueueItem{ID: "a", VideoID: "v1", Title: "A"})
	r.AddQueueItem(protocol.QueueItem{ID: "b", VideoID: "v2", Title: "B"})

	if err := r.ReorderQueue("b", ""); err != nil {
		t.Fatal(err)
	}
	if r.Queue[0].ID != "b" || r.Queue[1].ID != "a" {
		t.Fatalf("queue %+v", r.Queue)
	}
}

func TestQueueNewItemMetadata(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	item := NewQueueItem(protocol.Track{
		VideoID:    "vid12345678",
		Title:      "Song",
		DurationMs: 180_000,
	}, "alice", now)
	if item.ID == "" || item.AddedBy != "alice" || item.DurationMs != 180_000 {
		t.Fatalf("item %+v", item)
	}
}

func TestQueueAdvanceIfEnded(t *testing.T) {
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	r := NewRoom("q", protocol.Member{SessionID: "h", Nickname: "host"}, start, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "cur", Title: "Current", DurationMs: 1000}, start)
	r.Playback.Play(start)
	r.AddQueueItem(protocol.QueueItem{ID: "next", VideoID: "nxt", Title: "Next", DurationMs: 2000})
	r.Reactions = map[string]int{"🔥": 2}

	now := start.Add(1500 * time.Millisecond)
	if !r.AdvanceIfEnded(now) {
		t.Fatal("expected advance")
	}
	if r.Playback.Track().VideoID != "nxt" {
		t.Fatalf("track %+v", r.Playback.Track())
	}
	if len(r.Reactions) != 0 {
		t.Fatalf("reactions should reset %+v", r.Reactions)
	}
}

func TestQueueAdvanceIfEndedNoQueue(t *testing.T) {
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	r := NewRoom("q", protocol.Member{SessionID: "h", Nickname: "host"}, start, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "cur", Title: "Current", DurationMs: 1000}, start)
	r.Playback.Play(start)

	now := start.Add(1500 * time.Millisecond)
	if !r.AdvanceIfEnded(now) {
		t.Fatal("expected ended")
	}
	if r.Playback.Status() != protocol.PlaybackEnded {
		t.Fatalf("status %q", r.Playback.Status())
	}
}

func TestRoomIsHost(t *testing.T) {
	r := NewRoom("q", protocol.Member{SessionID: "host", Nickname: "host"}, time.Now(), chat.Options{})
	if !r.IsHost("host") || r.IsHost("guest") {
		t.Fatal("host check failed")
	}
}
