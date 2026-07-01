package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

const (
	defaultBin         = "yt-dlp"
	defaultTimeout     = 10 * time.Second
	defaultSearchCache = 5 * time.Minute
)

// Config configures the yt-dlp resolver.
type Config struct {
	Bin          string
	Timeout      time.Duration
	SearchCache  time.Duration
	Runner       CommandRunner
}

// Resolver resolves YouTube URLs and keyword searches via yt-dlp (ADR-007).
type Resolver struct {
	bin     string
	timeout time.Duration
	runner  CommandRunner
	cache   *searchCache
}

// NewResolver creates a resolver with defaults.
func NewResolver(cfg Config) *Resolver {
	bin := cfg.Bin
	if bin == "" {
		bin = os.Getenv("MUSIC_ROOM_YTDLP")
	}
	if bin == "" {
		bin = defaultBin
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	cacheTTL := cfg.SearchCache
	if cacheTTL == 0 {
		cacheTTL = defaultSearchCache
	}
	runner := cfg.Runner
	if runner == nil {
		runner = execRunner{}
	}
	return &Resolver{
		bin:     bin,
		timeout: timeout,
		runner:  runner,
		cache:   newSearchCache(cacheTTL),
	}
}

// Resolve returns track metadata for a YouTube URL or search query.
func (r *Resolver) Resolve(ctx context.Context, rawURL, query string) (protocol.Track, error) {
	rawURL = strings.TrimSpace(rawURL)
	query = strings.TrimSpace(query)
	if rawURL == "" && query == "" {
		return protocol.Track{}, ErrInvalidSource
	}
	if rawURL != "" && query != "" {
		return protocol.Track{}, ErrInvalidSource
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if rawURL != "" {
		id, ok := ParseVideoID(rawURL)
		if !ok {
			return protocol.Track{}, ErrInvalidSource
		}
		return r.resolveURL(ctx, id, rawURL)
	}
	if track, ok := r.cache.Get(query); ok {
		return track, nil
	}
	track, err := r.search(ctx, query)
	if err != nil {
		return protocol.Track{}, err
	}
	r.cache.Set(query, track)
	return track, nil
}

func (r *Resolver) resolveURL(ctx context.Context, id, sourceURL string) (protocol.Track, error) {
	entry, err := r.fetchMetadata(ctx, sourceURL)
	if err != nil {
		return protocol.Track{}, err
	}
	if entry.ID == "" {
		entry.ID = id
	}
	return entry.toTrack(), nil
}

func (r *Resolver) search(ctx context.Context, query string) (protocol.Track, error) {
	target := "ytsearch5:" + query
	entry, err := r.fetchMetadata(ctx, target, "-I", "1")
	if err != nil {
		return protocol.Track{}, err
	}
	if entry.ID == "" {
		return protocol.Track{}, ErrSourceUnavailable
	}
	return entry.toTrack(), nil
}

type ytdlpEntry struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Duration   float64 `json:"duration"`
	WebpageURL string  `json:"webpage_url"`
}

func (e ytdlpEntry) toTrack() protocol.Track {
	url := e.WebpageURL
	if url == "" && e.ID != "" {
		url = watchURL(e.ID)
	}
	var durationMs int64
	if e.Duration > 0 {
		durationMs = int64(e.Duration * 1000)
	}
	title := e.Title
	if title == "" {
		title = e.ID
	}
	return protocol.Track{
		VideoID:    e.ID,
		Title:      title,
		DurationMs: durationMs,
		SourceURL:  url,
	}
}

func (r *Resolver) fetchMetadata(ctx context.Context, target string, extraArgs ...string) (ytdlpEntry, error) {
	args := []string{
		"--dump-single-json",
		"--no-playlist",
		"--skip-download",
	}
	args = append(args, extraArgs...)
	args = append(args, target)

	out, err := r.runner.Output(ctx, r.bin, args...)
	if err != nil {
		return ytdlpEntry{}, fmt.Errorf("%w: %v", ErrSourceUnavailable, trimExecErr(err))
	}
	var entry ytdlpEntry
	if err := json.Unmarshal(out, &entry); err != nil {
		return ytdlpEntry{}, fmt.Errorf("%w: invalid metadata", ErrSourceUnavailable)
	}
	if entry.ID == "" && entry.WebpageURL != "" {
		if id, ok := ParseVideoID(entry.WebpageURL); ok {
			entry.ID = id
		}
	}
	if entry.ID == "" {
		return ytdlpEntry{}, ErrSourceUnavailable
	}
	return entry, nil
}

func trimExecErr(err error) string {
	msg := err.Error()
	if idx := strings.Index(msg, ": exit status"); idx >= 0 {
		return msg[:idx]
	}
	return msg
}
