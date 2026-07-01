package room

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

const (
	minNicknameRunes = 1
	maxNicknameRunes = 32
	minSlugRunes     = 1
	maxSlugRunes     = 64
)

// ValidationError describes invalid nickname or slug input.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateNickname checks AC-001/AC-002 rules.
func ValidateNickname(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ValidationError{Field: "nickname", Message: "must not be empty"}
	}
	if n := utf8.RuneCountInString(trimmed); n < minNicknameRunes || n > maxNicknameRunes {
		return "", ValidationError{Field: "nickname", Message: "must be 1–32 characters"}
	}
	return trimmed, nil
}

// ValidateSlug checks room slug format (AC-006).
func ValidateSlug(raw string) (string, error) {
	slug := strings.TrimSpace(strings.ToLower(raw))
	if slug == "" {
		return "", ValidationError{Field: "slug", Message: "must not be empty"}
	}
	if n := utf8.RuneCountInString(slug); n < minSlugRunes || n > maxSlugRunes {
		return "", ValidationError{Field: "slug", Message: "must be 1–64 characters"}
	}
	for i, r := range slug {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			return "", ValidationError{Field: "slug", Message: "only lowercase letters, digits, and hyphens allowed"}
		}
		if r == '-' && (i == 0 || i == len(slug)-1) {
			return "", ValidationError{Field: "slug", Message: "must not start or end with a hyphen"}
		}
	}
	return slug, nil
}

// DisplayName returns a disambiguated name when nickname already exists (AC-016).
func DisplayName(nickname, sessionID string, members []protocol.Member) string {
	dup := false
	for _, m := range members {
		if m.Nickname == nickname && m.SessionID != sessionID {
			dup = true
			break
		}
	}
	if !dup {
		return nickname
	}
	suffix := sessionID
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	return nickname + "#" + strings.ToLower(suffix)
}

// RecomputeDisplayNames updates display names for all members after membership changes.
func RecomputeDisplayNames(members []protocol.Member) []protocol.Member {
	out := make([]protocol.Member, len(members))
	copy(out, members)
	counts := map[string]int{}
	for _, m := range members {
		counts[m.Nickname]++
	}
	for i := range out {
		if counts[out[i].Nickname] <= 1 {
			out[i].DisplayName = out[i].Nickname
			continue
		}
		out[i].DisplayName = DisplayName(out[i].Nickname, out[i].SessionID, members)
	}
	return out
}
