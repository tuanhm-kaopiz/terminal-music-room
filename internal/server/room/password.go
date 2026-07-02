package room

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordMinRunes = 1
	passwordMaxRunes = 32
)

// ValidatePassword checks a non-empty room password (1–32 runes after trim).
func ValidatePassword(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ValidationError{Field: "password", Message: "must not be empty or whitespace only"}
	}
	if n := utf8.RuneCountInString(trimmed); n < passwordMinRunes || n > passwordMaxRunes {
		return "", ValidationError{Field: "password", Message: "must be 1–32 characters"}
	}
	return trimmed, nil
}

// IsEmptyPassword reports whether raw means no room password (open room).
func IsEmptyPassword(raw string) bool {
	return strings.TrimSpace(raw) == ""
}

// HashPassword returns a bcrypt hash for plain.
func HashPassword(plain string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
}

// CheckPassword reports whether plain matches hash.
func CheckPassword(hash []byte, plain string) bool {
	if len(hash) == 0 {
		return true
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(plain)) == nil
}
