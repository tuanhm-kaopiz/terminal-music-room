//go:build integration

package youtube

import (
	"context"
	"os/exec"
	"testing"
)

func TestResolveURLIntegration(t *testing.T) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not installed")
	}

	r := NewResolver(Config{})
	track, err := r.Resolve(context.Background(), "https://www.youtube.com/watch?v=jNQXAC9IVRw", "")
	if err != nil {
		t.Fatal(err)
	}
	if track.VideoID != "jNQXAC9IVRw" || track.Title == "" {
		t.Fatalf("track %+v", track)
	}
}

func TestResolveSearchIntegration(t *testing.T) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not installed")
	}

	r := NewResolver(Config{})
	track, err := r.Resolve(context.Background(), "", "lofi hip hop")
	if err != nil {
		t.Fatal(err)
	}
	if track.VideoID == "" || track.Title == "" {
		t.Fatalf("track %+v", track)
	}
}
