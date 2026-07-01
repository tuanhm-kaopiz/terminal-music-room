package room

import (
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

func TestValidateEmoji(t *testing.T) {
	t.Run("accepts emoji", func(t *testing.T) {
		got, err := ValidateEmoji("🔥")
		if err != nil || got != "🔥" {
			t.Fatalf("got %q err %v", got, err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		got, err := ValidateEmoji("  👍  ")
		if err != nil || got != "👍" {
			t.Fatalf("got %q err %v", got, err)
		}
	})

	t.Run("rejects empty", func(t *testing.T) {
		_, err := ValidateEmoji("   ")
		if !IsEmptyEmojiError(err) {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("rejects too long", func(t *testing.T) {
		_, err := ValidateEmoji("abcdefghij")
		if !IsInvalidEmojiError(err) {
			t.Fatalf("err %v", err)
		}
	})
}

func TestAddReactionAggregates(t *testing.T) {
	now := time.Now()
	r := NewRoom("test", protocol.Member{SessionID: "h", Nickname: "host"}, now, chat.Options{})
	r.Playback.LoadTrack(protocol.Track{VideoID: "abc123xyz01", Title: "t", DurationMs: 60_000}, now)
	if err := r.AddReaction("🔥"); err != nil {
		t.Fatal(err)
	}
	if err := r.AddReaction("🔥"); err != nil {
		t.Fatal(err)
	}
	if r.Reactions["🔥"] != 2 {
		t.Fatalf("counts %+v", r.Reactions)
	}
}

func TestAddReactionNoTrack(t *testing.T) {
	r := NewRoom("r", protocol.Member{SessionID: "h", Nickname: "h"}, time.Now(), chat.Options{})
	if err := r.AddReaction("🔥"); err != ErrNoTrackPlaying {
		t.Fatalf("got %v", err)
	}
}
