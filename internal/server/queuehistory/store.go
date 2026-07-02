package queuehistory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Entry is one queue-add event written to disk.
type Entry struct {
	At       time.Time
	AddedBy  string
	Source   string // original URL or search query submitted by the user
	IsURL    bool
	Track    protocol.Track
}

// Store appends queue-add history per room slug.
type Store struct {
	dir string
	mu  sync.Mutex
}

// NewStore creates a history store under dataDir (e.g. ./data/queue).
func NewStore(dataDir string) *Store {
	return &Store{dir: dataDir}
}

// Path returns the log file path for a room slug.
func (s *Store) Path(slug string) string {
	return filepath.Join(s.dir, slug+".queue.log")
}

// Append records a successful queue add. Errors are returned to the caller.
func (s *Store) Append(slug string, e Entry) error {
	if s == nil || s.dir == "" || slug == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	path := s.Path(slug)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, encodeLine(e))
	return err
}

func encodeLine(e Entry) string {
	at := e.At.UTC().Format(time.RFC3339Nano)
	kind := "query"
	if e.IsURL {
		kind = "url"
	}
	youtubeURL := e.Track.SourceURL
	if youtubeURL == "" && e.Track.VideoID != "" {
		youtubeURL = "https://www.youtube.com/watch?v=" + e.Track.VideoID
	}
	fields := []string{
		at,
		sanitize(e.AddedBy),
		kind,
		sanitize(e.Source),
		sanitize(e.Track.VideoID),
		sanitize(e.Track.Title),
		sanitize(youtubeURL),
	}
	return strings.Join(fields, "\t")
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
