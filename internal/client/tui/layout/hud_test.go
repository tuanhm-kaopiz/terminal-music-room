package layout

import "testing"

func TestCompute80x24(t *testing.T) {
	r := Compute(80, 24, false)
	if r.Degraded {
		t.Fatal("expected full layout at 80x24")
	}
	if r.Header.Height != headerRows {
		t.Fatalf("header h %d", r.Header.Height)
	}
	if r.NowPlaying.IsZero() || r.Members.IsZero() || r.Signals.IsZero() {
		t.Fatalf("top row: now=%+v crew=%+v sig=%+v", r.NowPlaying, r.Members, r.Signals)
	}
	if r.Queue.IsZero() || r.Chat.IsZero() {
		t.Fatalf("queue=%+v chat=%+v", r.Queue, r.Chat)
	}
	if r.Input.Height != inputRows || r.StatusBar.Height != statusRows {
		t.Fatalf("footer input=%+v status=%+v", r.Input, r.StatusBar)
	}
	totalH := r.Header.Height + r.NowPlaying.Height + r.Queue.Height + r.Chat.Height +
		r.Input.Height + r.StatusBar.Height
	if totalH > 24 {
		t.Fatalf("layout exceeds height: %d", totalH)
	}
}

func TestCompute120x40(t *testing.T) {
	r := Compute(120, 40, false)
	if r.Degraded {
		t.Fatal("expected full layout")
	}
	if r.NowPlaying.Width < r.Members.Width {
		// now playing should be the widest top panel
	}
	if r.Queue.Height < minQueueH || r.Chat.Height < minChatH {
		t.Fatalf("queue=%+v chat=%+v", r.Queue, r.Chat)
	}
	if r.Width != 120 || r.Height != 40 {
		t.Fatalf("size %dx%d", r.Width, r.Height)
	}
}

func TestCompute60x20Degraded(t *testing.T) {
	r := Compute(60, 20, true)
	if !r.Degraded {
		t.Fatal("expected degraded layout")
	}
	if !r.AsciiBorders {
		t.Fatal("ascii flag")
	}
	if !r.Signals.IsZero() || !r.Queue.IsZero() || !r.Chat.IsZero() {
		t.Fatalf("hidden panels should be zero: sig=%+v q=%+v chat=%+v",
			r.Signals, r.Queue, r.Chat)
	}
	if r.Header.IsZero() || r.NowPlaying.IsZero() {
		t.Fatalf("header=%+v now=%+v", r.Header, r.NowPlaying)
	}
}

func TestComputeNarrowWidthDegraded(t *testing.T) {
	r := Compute(79, 30, false)
	if !r.Degraded {
		t.Fatal("width < 80 should degrade")
	}
}

func TestComputeShortHeightDegraded(t *testing.T) {
	r := Compute(100, 23, false)
	if !r.Degraded {
		t.Fatal("height < 24 should degrade")
	}
}

func TestTopRowWidthsSum(t *testing.T) {
	r := Compute(80, 24, false)
	sum := r.NowPlaying.Width + r.Members.Width + r.Signals.Width
	if sum != 80 {
		t.Fatalf("top row widths %d != 80", sum)
	}
}
