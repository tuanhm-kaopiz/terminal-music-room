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
	if r.NowPlaying.IsZero() || r.Members.IsZero() {
		t.Fatalf("row1: now=%+v crew=%+v", r.NowPlaying, r.Members)
	}
	if r.Queue.IsZero() || r.Signals.IsZero() || r.Chat.IsZero() {
		t.Fatalf("row2/3: queue=%+v sig=%+v chat=%+v", r.Queue, r.Signals, r.Chat)
	}
	if r.NowPlaying.Height != r.Members.Height {
		t.Fatalf("row1 heights differ: now=%d crew=%d", r.NowPlaying.Height, r.Members.Height)
	}
	if r.Queue.Height != r.Signals.Height {
		t.Fatalf("row2 heights differ: queue=%d sig=%d", r.Queue.Height, r.Signals.Height)
	}
	totalH := r.Header.Height + r.BodyHeight() + r.Input.Height + r.StatusBar.Height
	if totalH > 24 {
		t.Fatalf("layout exceeds height: %d", totalH)
	}
}

func TestCompute120x40(t *testing.T) {
	r := Compute(120, 40, false)
	if r.Degraded {
		t.Fatal("expected full layout")
	}
	if r.Chat.Height < minRow3H {
		t.Fatalf("chat=%+v", r.Chat)
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

func TestRow1WidthsSum(t *testing.T) {
	r := Compute(80, 24, false)
	sum := r.NowPlaying.Width + r.Members.Width
	want := 80 - 2*panelBorder
	if sum != want {
		t.Fatalf("row1 lipgloss widths %d != %d", sum, want)
	}
}

func TestRow2WidthsSum(t *testing.T) {
	r := Compute(80, 24, false)
	sum := r.Queue.Width + r.Signals.Width
	want := 80 - 2*panelBorder
	if sum != want {
		t.Fatalf("row2 lipgloss widths %d != %d", sum, want)
	}
}

func TestBodyHeight(t *testing.T) {
	r := Compute(80, 24, false)
	want := r.NowPlaying.Height + r.Queue.Height + r.Chat.Height
	if r.BodyHeight() != want {
		t.Fatalf("BodyHeight %d != %d", r.BodyHeight(), want)
	}
}
