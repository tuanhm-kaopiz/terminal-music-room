package hub

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/server/room"
)

const maxSessionsPerIP = 3

// Session is an authenticated client connection identity.
type Session struct {
	ID             string
	Nickname       string
	DisplayName    string
	IP             string
	RoomSlug       string
	ConnectedAt    time.Time
	LastSeen       time.Time
	DisconnectedAt *time.Time
}

// SessionStore tracks active sessions in memory.
type SessionStore struct {
	mu   sync.RWMutex
	byID map[string]*Session
	byIP map[string]map[string]struct{}
}

// NewSessionStore creates an empty session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		byID: make(map[string]*Session),
		byIP: make(map[string]map[string]struct{}),
	}
}

// Register adds or replaces a session (reconnect updates LastSeen).
func (st *SessionStore) Register(sess *Session) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if old, ok := st.byID[sess.ID]; ok {
		delete(st.byIP[old.IP], old.ID)
	}
	st.byID[sess.ID] = sess
	if st.byIP[sess.IP] == nil {
		st.byIP[sess.IP] = make(map[string]struct{})
	}
	st.byIP[sess.IP][sess.ID] = struct{}{}
}

// Get returns a session by ID.
func (st *SessionStore) Get(id string) (*Session, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	sess, ok := st.byID[id]
	return sess, ok
}

// Remove deletes a session on disconnect.
func (st *SessionStore) Remove(id string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	sess, ok := st.byID[id]
	if !ok {
		return
	}
	delete(st.byID, id)
	if ids, ok := st.byIP[sess.IP]; ok {
		delete(ids, id)
		if len(ids) == 0 {
			delete(st.byIP, sess.IP)
		}
	}
}

// CountByIP returns how many sessions an IP currently has.
func (st *SessionStore) CountByIP(ip string) int {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return len(st.byIP[ip])
}

// Disconnect marks a session offline but keeps it for reconnect within TTL (AC-048).
func (st *SessionStore) Disconnect(id string, now time.Time) {
	st.mu.Lock()
	defer st.mu.Unlock()
	sess, ok := st.byID[id]
	if !ok {
		return
	}
	sess.LastSeen = now
	sess.DisconnectedAt = &now
	if ids, ok := st.byIP[sess.IP]; ok {
		delete(ids, id)
		if len(ids) == 0 {
			delete(st.byIP, sess.IP)
		}
	}
}

func newSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *Server) createSession(ip, nickname string, now time.Time) (*Session, error) {
	valid, err := room.ValidateNickname(nickname)
	if err != nil {
		return nil, err
	}
	id, err := newSessionID()
	if err != nil {
		return nil, err
	}
	sess := &Session{
		ID:          id,
		Nickname:    valid,
		DisplayName: valid,
		IP:          ip,
		ConnectedAt: now,
		LastSeen:    now,
	}
	s.sessions.Register(sess)
	return sess, nil
}
