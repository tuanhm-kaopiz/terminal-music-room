package room

import (
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
)

// Manager owns the global slug registry and room instances.
type Manager struct {
	mu       sync.RWMutex
	rooms    map[string]*Room
	chatOpts chat.Options
}

// NewManager creates an empty room manager.
func NewManager(chatOpts chat.Options) *Manager {
	return &Manager{rooms: make(map[string]*Room), chatOpts: chatOpts}
}

// Create registers a new room with a unique slug (AC-004, AC-005).
func (m *Manager) Create(rawSlug string, host protocol.Member, now time.Time, password string) (*Room, error) {
	slug, err := ValidateSlug(rawSlug)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.rooms[slug]; ok {
		return nil, ErrSlugTaken
	}
	r := NewRoom(slug, host, now, m.chatOpts)
	if !IsEmptyPassword(password) {
		if err := r.SetPassword(password); err != nil {
			return nil, err
		}
	}
	m.rooms[slug] = r
	return r, nil
}

// Get returns a room by slug.
func (m *Manager) Get(slug string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[slug]
	return r, ok
}

// Join adds a member to an existing room (AC-007–010).
func (m *Manager) Join(rawSlug string, member protocol.Member, now time.Time, password string) (*Room, error) {
	slug, err := ValidateSlug(rawSlug)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[slug]
	if !ok {
		return nil, ErrRoomNotFound
	}
	if r.PasswordProtected() {
		if IsEmptyPassword(password) {
			return nil, ErrAuthRequired
		}
		if !r.CheckPassword(password) {
			return nil, ErrAuthFailed
		}
	}
	if err := r.AddMember(member, now); err != nil {
		return nil, err
	}
	return r, nil
}

// LeaveResult describes the outcome of a member leaving a room.
type LeaveResult struct {
	Destroyed   bool
	HostChanged bool
	Room        *Room
}

// Leave removes a member; destroys the room when empty (AC-011–014).
func (m *Manager) Leave(slug, sessionID string) (LeaveResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[slug]
	if !ok {
		return LeaveResult{}, ErrRoomNotFound
	}
	emptied, hostChanged := r.RemoveMember(sessionID)
	if emptied {
		delete(m.rooms, slug)
		return LeaveResult{Destroyed: true}, nil
	}
	return LeaveResult{HostChanged: hostChanged, Room: r}, nil
}

// KickMember removes a non-host member when requested by the room host.
func (m *Manager) KickMember(slug, hostSessionID, targetSessionID string) (LeaveResult, error) {
	m.mu.RLock()
	r, ok := m.rooms[slug]
	if !ok {
		m.mu.RUnlock()
		return LeaveResult{}, ErrRoomNotFound
	}
	if !r.IsHost(hostSessionID) {
		m.mu.RUnlock()
		return LeaveResult{}, ErrForbidden
	}
	if hostSessionID == targetSessionID || r.IsHost(targetSessionID) {
		m.mu.RUnlock()
		return LeaveResult{}, ErrForbidden
	}
	if _, ok := r.FindMember(targetSessionID); !ok {
		m.mu.RUnlock()
		return LeaveResult{}, ErrNotInRoom
	}
	m.mu.RUnlock()
	return m.Leave(slug, targetSessionID)
}

// Count returns the number of active rooms.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// Slugs returns a snapshot of active room slugs.
func (m *Manager) Slugs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.rooms))
	for slug := range m.rooms {
		out = append(out, slug)
	}
	return out
}

// PlayingSlugs returns slugs where playback is currently playing.
func (m *Manager) PlayingSlugs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.rooms))
	for slug, r := range m.rooms {
		if r.Playback.Status() == protocol.PlaybackPlaying {
			out = append(out, slug)
		}
	}
	return out
}

// Modify runs fn on a room while holding the manager write lock.
func (m *Manager) Modify(slug string, fn func(*Room) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[slug]
	if !ok {
		return ErrRoomNotFound
	}
	return fn(r)
}

// ForEachPlaying invokes fn for each room with playing playback status.
func (m *Manager) ForEachPlaying(fn func(slug string, r *Room)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for slug, r := range m.rooms {
		if r.Playback.Status() == protocol.PlaybackPlaying {
			fn(slug, r)
		}
	}
}

// ForEach invokes fn for every active room.
func (m *Manager) ForEach(fn func(slug string, r *Room)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for slug, r := range m.rooms {
		fn(slug, r)
	}
}
