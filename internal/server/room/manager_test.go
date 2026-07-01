package room

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

func TestManagerCreateAndDuplicateSlug(t *testing.T) {
	m := NewManager(chat.Options{})
	now := time.Now()
	host := protocol.Member{SessionID: "h1", Nickname: "host"}

	r, err := m.Create("backend-team", host, now)
	if err != nil || r.Slug != "backend-team" {
		t.Fatalf("create: %v room %+v", err, r)
	}
	_, err = m.Create("backend-team", host, now)
	if err != ErrSlugTaken {
		t.Fatalf("got %v", err)
	}
}

func TestManagerCreateInvalidSlug(t *testing.T) {
	m := NewManager(chat.Options{})
	_, err := m.Create("-bad", protocol.Member{SessionID: "h", Nickname: "h"}, time.Now())
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestManagerJoinNotFound(t *testing.T) {
	m := NewManager(chat.Options{})
	_, err := m.Join("missing", protocol.Member{SessionID: "a", Nickname: "a"}, time.Now())
	if err != ErrRoomNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestManagerJoinFull(t *testing.T) {
	m := NewManager(chat.Options{})
	now := time.Now()
	_, err := m.Create("full", protocol.Member{SessionID: "h", Nickname: "host"}, now)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 19; i++ {
		_, err = m.Join("full", protocol.Member{
			SessionID: "m" + string(rune('a'+i)),
			Nickname:  "u",
		}, now)
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = m.Join("full", protocol.Member{SessionID: "overflow", Nickname: "x"}, now)
	if err != ErrRoomFull {
		t.Fatalf("got %v", err)
	}
}

func TestManagerLeaveDestroysEmptyRoom(t *testing.T) {
	m := NewManager(chat.Options{})
	now := time.Now()
	_, err := m.Create("solo", protocol.Member{SessionID: "only", Nickname: "solo"}, now)
	if err != nil {
		t.Fatal(err)
	}
	res, err := m.Leave("solo", "only")
	if err != nil || !res.Destroyed {
		t.Fatalf("res %+v err %v", res, err)
	}
	if m.Count() != 0 {
		t.Fatal("room should be removed from registry")
	}
	_, ok := m.Get("solo")
	if ok {
		t.Fatal("slug should be free")
	}
}

func TestManagerLeaveHostTransfer(t *testing.T) {
	m := NewManager(chat.Options{})
	now := time.Now()
	_, err := m.Create("team", protocol.Member{SessionID: "host", Nickname: "host"}, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Join("team", protocol.Member{SessionID: "guest", Nickname: "guest"}, now)
	if err != nil {
		t.Fatal(err)
	}
	res, err := m.Leave("team", "host")
	if err != nil || !res.HostChanged || res.Room.HostSessionID != "guest" {
		t.Fatalf("res %+v err %v", res, err)
	}
}

func TestManagerJoinSnapshotReady(t *testing.T) {
	m := NewManager(chat.Options{})
	now := time.Now()
	r, err := m.Create("snap", protocol.Member{SessionID: "h", Nickname: "host"}, now)
	if err != nil {
		t.Fatal(err)
	}
	snap := r.Snapshot(now)
	if snap.Slug != "snap" || len(snap.Members) != 1 {
		t.Fatalf("snap %+v", snap)
	}
}
