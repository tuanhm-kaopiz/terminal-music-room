package tui

import "testing"

func TestTruncate(t *testing.T) {
	t.Parallel()
	if got := truncate("hello", 10); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got := truncate("hello world", 8); got != "hello w…" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatMs(t *testing.T) {
	t.Parallel()
	if got := formatMs(125000); got != "2:05" {
		t.Fatalf("got %q", got)
	}
}
