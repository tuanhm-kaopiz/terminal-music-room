package room

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

func TestValidateNickname(t *testing.T) {
	got, err := ValidateNickname("kaopiz")
	if err != nil || got != "kaopiz" {
		t.Fatalf("got %q err %v", got, err)
	}
	_, err = ValidateNickname("")
	if err == nil {
		t.Fatal("expected error for empty")
	}
	_, err = ValidateNickname("   ")
	if err == nil {
		t.Fatal("expected error for whitespace")
	}
	long := make([]rune, 33)
	for i := range long {
		long[i] = 'a'
	}
	_, err = ValidateNickname(string(long))
	if err == nil {
		t.Fatal("expected error for >32 runes")
	}
}

func TestValidateSlug(t *testing.T) {
	cases := []struct {
		in    string
		want  string
		valid bool
	}{
		{"backend-team", "backend-team", true},
		{"Backend-Team", "backend-team", true},
		{"", "", false},
		{"-bad", "", false},
		{"bad-", "", false},
		{"bad slug", "", false},
		{"team_1", "", false},
	}
	for _, tc := range cases {
		got, err := ValidateSlug(tc.in)
		if tc.valid {
			if err != nil || got != tc.want {
				t.Fatalf("%q: got %q err %v", tc.in, got, err)
			}
		} else if err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
	}
}

func TestDisplayNameDisambiguation(t *testing.T) {
	members := []protocol.Member{
		{SessionID: "sess-aaaa1111", Nickname: "kaopiz"},
		{SessionID: "sess-bbbb2222", Nickname: "kaopiz"},
	}
	updated := RecomputeDisplayNames(members)
	if updated[0].DisplayName == updated[1].DisplayName {
		t.Fatalf("display names should differ: %q vs %q", updated[0].DisplayName, updated[1].DisplayName)
	}
}

func TestRoomAddRemoveHostTransfer(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	host := protocol.Member{SessionID: "host-1", Nickname: "host"}
	r := NewRoom("team", host, now, chat.Options{})

	m2 := protocol.Member{SessionID: "mem-2", Nickname: "guest"}
	if err := r.AddMember(m2, now); err != nil {
		t.Fatal(err)
	}

	emptied, hostChanged := r.RemoveMember("host-1")
	if emptied || !hostChanged {
		t.Fatalf("emptied=%v hostChanged=%v", emptied, hostChanged)
	}
	if r.HostSessionID != "mem-2" {
		t.Fatalf("host %q", r.HostSessionID)
	}
	if !r.Members[0].IsHost {
		t.Fatal("mem-2 should be host")
	}
}

func TestRoomFull(t *testing.T) {
	now := time.Now()
	host := protocol.Member{SessionID: "h", Nickname: "host"}
	r := NewRoom("full", host, now, chat.Options{})
	for i := 0; i < 19; i++ {
		_ = r.AddMember(protocol.Member{
			SessionID: "m" + string(rune('a'+i)),
			Nickname:  "u",
		}, now)
	}
	if !r.IsFull() {
		t.Fatal("expected full at 20")
	}
	err := r.AddMember(protocol.Member{SessionID: "overflow", Nickname: "x"}, now)
	if err != ErrRoomFull {
		t.Fatalf("got %v", err)
	}
}

func TestSnapshotPlaybackPosition(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	host := protocol.Member{SessionID: "h", Nickname: "host"}
	r := NewRoom("snap", host, now, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "v", Title: "t", DurationMs: 60_000}, now)
	r.Playback.Play(now)

	snap := r.Snapshot(now.Add(2 * time.Second))
	if snap.Playback.PositionMs != 2000 {
		t.Fatalf("position %d", snap.Playback.PositionMs)
	}
	if snap.Playback.Status != protocol.PlaybackPlaying {
		t.Fatalf("status %q", snap.Playback.Status)
	}
}

func TestRoomPasswordProtected(t *testing.T) {
	now := time.Now()
	r := NewRoom("locked", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	if r.PasswordProtected() {
		t.Fatal("new room should be open")
	}
	if err := r.SetPassword("secret"); err != nil {
		t.Fatal(err)
	}
	if !r.PasswordProtected() || !r.CheckPassword("secret") {
		t.Fatal("expected password match")
	}
	if r.CheckPassword("wrong") {
		t.Fatal("expected password mismatch")
	}
	snap := r.Snapshot(now)
	if !snap.PasswordProtected {
		t.Fatal("snapshot password_protected")
	}
}

func TestRemoveLastMemberEmptiesRoom(t *testing.T) {
	now := time.Now()
	r := NewRoom("solo", protocol.Member{SessionID: "only", Nickname: "solo"}, now, chat.Options{})
	emptied, _ := r.RemoveMember("only")
	if !emptied {
		t.Fatal("expected room to empty")
	}
}
