package panels

import (
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/tui/layout"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestQueueSelectedMarker(t *testing.T) {
	tm := theme.Default()
	v := fixtureView()
	reg := layout.Compute(80, 24, false)
	out := Queue(tm, v, reg.Queue.Width, reg.Queue.Height, RenderOpts{QueueSelectedIdx: 1})
	if !strings.Contains(out, "› 2. Track Three") {
		t.Fatalf("expected selection marker on item 2: %q", out)
	}
	if strings.Contains(out, "› 1. Track Two") {
		t.Fatal("first item should not be selected")
	}
}

func TestQueueReorderTargetsUp(t *testing.T) {
	items := fixtureView().Room.Queue
	// idx 1 (q2) moves to front
	itemID, afterID, ok := queueReorderTargets(items, 1, -1)
	if !ok || itemID != "q2" || afterID != "" {
		t.Fatalf("got %q %q %v", itemID, afterID, ok)
	}
}

func TestQueueReorderTargetsDown(t *testing.T) {
	items := fixtureView().Room.Queue
	itemID, afterID, ok := queueReorderTargets(items, 0, 1)
	if !ok || itemID != "q1" || afterID != "q2" {
		t.Fatalf("got %q after %q ok=%v", itemID, afterID, ok)
	}
}

// queueReorderTargets mirrors keys.QueueReorderTargets for panel-local tests.
func queueReorderTargets(items []protocol.QueueItem, idx, direction int) (string, string, bool) {
	if idx < 0 || idx >= len(items) || direction == 0 {
		return "", "", false
	}
	itemID := items[idx].ID
	switch direction {
	case -1:
		if idx == 0 {
			return "", "", false
		}
		if idx == 1 {
			return itemID, "", true
		}
		return itemID, items[idx-2].ID, true
	case 1:
		if idx >= len(items)-1 {
			return "", "", false
		}
		return itemID, items[idx+1].ID, true
	default:
		return "", "", false
	}
}
