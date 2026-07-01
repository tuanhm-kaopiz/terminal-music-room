package youtube

import (
	"net/url"
	"regexp"
	"strings"
)

var videoIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

// ParseVideoID extracts a YouTube video ID from a URL or bare ID string.
func ParseVideoID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if videoIDPattern.MatchString(raw) {
		return raw, true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	switch host {
	case "youtube.com", "m.youtube.com", "music.youtube.com":
		if u.Path == "/watch" {
			if id := u.Query().Get("v"); videoIDPattern.MatchString(id) {
				return id, true
			}
		}
		if strings.HasPrefix(u.Path, "/shorts/") {
			id := strings.TrimPrefix(u.Path, "/shorts/")
			id = strings.Trim(id, "/")
			if videoIDPattern.MatchString(id) {
				return id, true
			}
		}
	case "youtu.be":
		id := strings.Trim(u.Path, "/")
		if videoIDPattern.MatchString(id) {
			return id, true
		}
	}
	return "", false
}

func watchURL(id string) string {
	return "https://www.youtube.com/watch?v=" + id
}
