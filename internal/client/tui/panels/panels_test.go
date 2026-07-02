package panels

import (
	"fmt"
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/layout"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func fixtureView() state.View {
	return FixtureView()
}

func TestHeaderGolden(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Header(tm, v, reg.Header.Width, reg.Header.Height)
	if !strings.Contains(out, "backend-team") || !strings.Contains(out, "CREW: 2") {
		t.Fatalf("header missing content: %q", out)
	}
	if !strings.Contains(out, "connected") {
		t.Fatalf("conn badge missing: %q", out)
	}
}

func TestNowPlayingGolden(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, RenderOpts{})
	if !strings.Contains(out, "Neon Nights") || !strings.Contains(out, "NOW PLAYING") {
		t.Fatalf("now playing: %q", out)
	}
	if !strings.Contains(out, "█") || !strings.Contains(out, "░") {
		t.Fatalf("progress bar missing: %q", out)
	}
}

func TestNowPlaying_LongTitleTruncate(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	long := strings.Repeat("A", 60)
	v.Room.Playback.Track.Title = long
	reg := layout.Compute(80, 24, false)
	out := NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, RenderOpts{})
	if strings.Contains(out, long) {
		t.Fatal("title should be truncated")
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("expected ellipsis in truncated title: %q", out)
	}
}

func TestMembersHostMarker(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Members(tm, v, reg.Members.Width, reg.Members.Height, RenderOpts{})
	if !strings.Contains(out, "host#1") || !strings.Contains(out, "guest#2") {
		t.Fatalf("members: %q", out)
	}
	if !strings.Contains(out, "*") {
		t.Fatal("expected host marker")
	}
}

func TestMembersScroll(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	for i := 3; i <= 12; i++ {
		v.Room.Members = append(v.Room.Members, protocol.Member{
			SessionID:   fmt.Sprintf("s%d", i),
			DisplayName: fmt.Sprintf("guest#%d", i),
		})
	}
	reg := layout.Compute(80, 24, false)
	out := Members(tm, v, reg.Members.Width, reg.Members.Height, RenderOpts{MembersScroll: 3})
	if strings.Contains(out, "host#1") {
		t.Fatalf("scrolled view should skip early members: %q", out)
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("expected more-below indicator: %q", out)
	}
}

func TestQueueScroll(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Queue(tm, v, reg.Queue.Width, reg.Queue.Height, RenderOpts{QueueScroll: 1})
	if strings.Contains(out, "Track Two") {
		t.Fatalf("scrolled view should skip first item: %q", out)
	}
	if !strings.Contains(out, "Track Three") {
		t.Fatalf("queue: %q", out)
	}
}

func TestChatGolden(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Chat(tm, v, reg.Chat.Width, reg.Chat.Height, RenderOpts{})
	if !strings.Contains(out, "hello") {
		t.Fatalf("chat: %q", out)
	}
}

func TestSignalsVoteAndReactions(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Signals(tm, v, reg.Signals.Width, reg.Signals.Height, RenderOpts{})
	if !strings.Contains(out, "SKIP") || !strings.Contains(out, "1/2") {
		t.Fatalf("vote: %q", out)
	}
	if !strings.Contains(out, "🔥1") {
		t.Fatalf("reactions: %q", out)
	}
}

func TestStatusBarHostHint(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	out := StatusBar(tm, v, 80, true)
	if !strings.Contains(out, "d del") || !strings.Contains(out, "reord") {
		t.Fatalf("host hint: %q", out)
	}
}

func TestStatusBarShowsError(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.LastErr = &protocol.ErrorPayload{Message: "forbidden"}
	out := StatusBar(tm, v, 80, false)
	if !strings.Contains(out, "forbidden") {
		t.Fatalf("error toast: %q", out)
	}
}

func TestReconnectingBadge(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Status = state.StatusReconnecting
	reg := layout.Compute(80, 24, false)
	out := Header(tm, v, reg.Header.Width, reg.Header.Height)
	if !strings.Contains(out, "reconnecting") {
		t.Fatalf("header: %q", out)
	}
}

func TestLeaveConnectionDisconnectedInRoom(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Status = state.StatusDisconnected
	reg := layout.Compute(80, 24, false)
	out := Header(tm, v, reg.Header.Width, reg.Header.Height)
	if !strings.Contains(out, "disconnected") || !strings.Contains(out, "retry") {
		t.Fatalf("header: %q", out)
	}
}

func TestLeaveConnectionRejoinHint(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Status = state.StatusDisconnected
	v.InRoom = false
	v.Room = protocol.RoomSnapshot{}
	reg := layout.Compute(80, 24, false)
	out := Header(tm, v, reg.Header.Width, reg.Header.Height)
	if !strings.Contains(out, "rejoin") {
		t.Fatalf("expected rejoin hint (AC-043): %q", out)
	}
}

func TestConnLabels(t *testing.T) {
	cases := []struct {
		status state.ConnStatus
		want   string
	}{
		{state.StatusConnected, "connected"},
		{state.StatusReconnecting, "reconnecting"},
		{state.StatusDisconnected, "disconnected"},
	}
	for _, tc := range cases {
		if got := ConnLabel(tc.status); got != tc.want {
			t.Fatalf("status %q: got %q want %q", tc.status, got, tc.want)
		}
	}
}

func TestTruncateUtil(t *testing.T) {
	if got := truncate("hello world", 8); got != "hello w…" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatMsUtil(t *testing.T) {
	if got := formatMs(125000); got != "2:05" {
		t.Fatalf("got %q", got)
	}
}
