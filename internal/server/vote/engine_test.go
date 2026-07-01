package vote

import (
	"fmt"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
	"github.com/terminal-music-room/music-room/internal/server/playback"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

func testRoom(members int, queue []protocol.QueueItem) *room.Room {
	now := time.Now()
	r := room.NewRoom("vote", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	for i := 1; i < members; i++ {
		_ = r.AddMember(protocol.Member{SessionID: fmt.Sprintf("m%d", i), Nickname: "u"}, now)
	}
	r.Queue = queue
	r.Playback = playback.NewClock()
	r.Playback.LoadTrack(protocol.Track{VideoID: "cur", Title: "Current", DurationMs: 60_000}, now)
	r.Playback.Play(now)
	return r
}

func TestThreshold(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 4: 3, 5: 3, 20: 11}
	for n, want := range cases {
		if got := Threshold(n); got != want {
			t.Fatalf("N=%d got %d want %d", n, got, want)
		}
	}
}

func TestCastSkipPassesAtMajority(t *testing.T) {
	r := testRoom(3, nil)
	now := time.Now()
	cfg := Config{Timeout: time.Minute}

	res, err := CastSkip(r, "h", now, cfg)
	if err != nil || res.Vote == nil || res.Passed {
		t.Fatalf("first cast %+v err %v", res, err)
	}
	res, err = CastSkip(r, "m1", now, cfg)
	if err != nil || !res.Passed || r.Vote != nil {
		t.Fatalf("second cast %+v vote %v err %v", res, r.Vote, err)
	}
	if r.Playback.Status() != protocol.PlaybackEnded {
		t.Fatalf("status %q", r.Playback.Status())
	}
}

func TestCastSkipDedupesVoter(t *testing.T) {
	r := testRoom(5, nil)
	now := time.Now()
	cfg := Config{Timeout: time.Minute}

	_, _ = CastSkip(r, "h", now, cfg)
	_, _ = CastSkip(r, "h", now, cfg)
	if len(r.Vote.Voters) != 1 {
		t.Fatalf("voters %v", r.Vote.Voters)
	}
}

func TestCastPriorityPromotesItem(t *testing.T) {
	r := testRoom(2, []protocol.QueueItem{
		{ID: "a", Title: "A"},
		{ID: "b", Title: "B"},
	})
	now := time.Now()
	cfg := Config{Timeout: time.Minute}

	_, err := CastPriority(r, "h", "b", now, cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CastPriority(r, "m1", "b", now, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Queue[0].ID != "b" {
		t.Fatalf("queue %+v", r.Queue)
	}
}

func TestCastPriorityCancelMissingTarget(t *testing.T) {
	r := testRoom(2, []protocol.QueueItem{
		{ID: "a", Title: "A"},
		{ID: "b", Title: "B"},
	})
	now := time.Now()
	cfg := Config{Timeout: time.Minute}
	_, _ = CastPriority(r, "h", "b", now, cfg)
	r.Queue = r.Queue[:1]
	res, ok := cancelPriorityIfMissing(r)
	if !ok || !res.Cancelled || r.Vote != nil {
		t.Fatalf("res %+v ok %v vote %v", res, ok, r.Vote)
	}
}

func TestExpireVote(t *testing.T) {
	r := testRoom(3, nil)
	now := time.Now()
	r.Vote = startVote(protocol.VoteKindSkip, "", 3, "h", now.Add(-2*time.Minute))
	cfg := Config{Timeout: time.Minute}
	res, ok := expireIfNeeded(r, now, cfg)
	if !ok || !res.Cancelled || r.Vote != nil {
		t.Fatalf("res %+v ok %v", res, ok)
	}
}

func TestClearOnTrackChange(t *testing.T) {
	r := testRoom(2, nil)
	r.Vote = startVote(protocol.VoteKindSkip, "", 2, "h", time.Now())
	res, ok := ClearOnTrackChange(r)
	if !ok || !res.Cancelled || r.Vote != nil {
		t.Fatalf("res %+v ok %v", res, ok)
	}
}
