package room

import (
	"strings"
	"unicode/utf8"
)

// Reaction validation errors.
var (
	errEmptyEmoji   = errReaction("emoji required")
	errInvalidEmoji = errReaction("invalid emoji")
)

type errReaction string

func (e errReaction) Error() string { return string(e) }

// ValidateEmoji checks reaction input (AC-045, AC-034 pass-through).
func ValidateEmoji(raw string) (string, error) {
	emoji := strings.TrimSpace(raw)
	if emoji == "" {
		return "", errEmptyEmoji
	}
	if utf8.RuneCountInString(emoji) > 8 {
		return "", errInvalidEmoji
	}
	return emoji, nil
}

// ResetReactions clears emoji reactions (AC-046).
func (r *Room) ResetReactions() {
	r.Reactions = make(map[string]int)
}

// AddReaction records an emoji reaction on the current track (AC-045).
func (r *Room) AddReaction(emoji string) error {
	if r.Playback.Track() == nil {
		return ErrNoTrackPlaying
	}
	validated, err := ValidateEmoji(emoji)
	if err != nil {
		return err
	}
	if r.Reactions == nil {
		r.Reactions = make(map[string]int)
	}
	r.Reactions[validated]++
	return nil
}

// ReactionCounts returns a copy of aggregated reaction counts.
func (r *Room) ReactionCounts() map[string]int {
	out := make(map[string]int, len(r.Reactions))
	for k, v := range r.Reactions {
		out[k] = v
	}
	return out
}

// IsEmptyEmojiError reports blank emoji validation failures.
func IsEmptyEmojiError(err error) bool {
	return err == errEmptyEmoji
}

// IsInvalidEmojiError reports malformed emoji validation failures.
func IsInvalidEmojiError(err error) bool {
	return err == errInvalidEmoji
}
