package youtube

import (
	"strings"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

type searchCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]cacheEntry
}

type cacheEntry struct {
	track   protocol.Track
	expires time.Time
}

func newSearchCache(ttl time.Duration) *searchCache {
	return &searchCache{
		ttl:   ttl,
		items: make(map[string]cacheEntry),
	}
}

func (c *searchCache) Get(query string) (protocol.Track, bool) {
	key := normalizeQuery(query)
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expires) {
		return protocol.Track{}, false
	}
	return entry.track, true
}

func (c *searchCache) Set(query string, track protocol.Track) {
	key := normalizeQuery(query)
	c.mu.Lock()
	c.items[key] = cacheEntry{track: track, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func normalizeQuery(q string) string {
	return strings.ToLower(strings.TrimSpace(q))
}
