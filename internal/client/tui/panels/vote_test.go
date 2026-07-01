package panels

import (
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestVoteProgressSkip(t *testing.T) {
	tm := theme.Default()
	out := Vote(tm, fixtureView(), 40)
	if !strings.Contains(out, "SKIP") || !strings.Contains(out, "1/2") {
		t.Fatalf("skip vote progress: %q", out)
	}
	if !strings.Contains(out, "█") {
		t.Fatalf("expected progress bar: %q", out)
	}
}

func TestVoteProgressPriority(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Room.Vote = &protocol.Vote{
		Kind:      protocol.VoteKindPriority,
		TargetID:  "q2",
		Voters:    []string{"sess-host"},
		Threshold: 2,
	}
	out := Vote(tm, v, 50)
	if !strings.Contains(out, "PRIORITY") || !strings.Contains(out, "Track Three") {
		t.Fatalf("priority vote: %q", out)
	}
}

func TestVoteNoActive(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Room.Vote = nil
	out := Vote(tm, v, 40)
	if !strings.Contains(out, "no active vote") {
		t.Fatalf("vote: %q", out)
	}
}

func TestVoteReactionsNoTrack(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	v.Room.Playback.Track = nil
	v.Room.Reactions = map[string]int{"🔥": 1}
	out := Reactions(tm, v, 40)
	if out != tm.Muted().Render("—") {
		t.Fatalf("reactions without track: %q", out)
	}
}

func TestVoteReactionsQuickHint(t *testing.T) {
	tm := theme.Default()
	out := Reactions(tm, fixtureView(), 40)
	if !strings.Contains(out, "🔥1") || !strings.Contains(out, "quick react") {
		t.Fatalf("reactions: %q", out)
	}
}
