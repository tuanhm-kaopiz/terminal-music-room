package hub

import (
	"sync"

	"github.com/coder/websocket"
)

type wsClient struct {
	mu        sync.Mutex
	conn      *websocket.Conn
	sessionID string
	roomSlug  string
}

type clientRegistry struct {
	mu        sync.RWMutex
	bySession map[string]*wsClient
}

func newClientRegistry() *clientRegistry {
	return &clientRegistry{bySession: make(map[string]*wsClient)}
}

func (r *clientRegistry) register(sessionID string, conn *websocket.Conn) *wsClient {
	c := &wsClient{conn: conn, sessionID: sessionID}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bySession[sessionID] = c
	return c
}

func (r *clientRegistry) unregister(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.bySession, sessionID)
}

func (r *clientRegistry) unregisterConn(sessionID string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.bySession[sessionID]
	if !ok || c.conn != conn {
		return
	}
	delete(r.bySession, sessionID)
}

func (r *clientRegistry) get(sessionID string) (*wsClient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.bySession[sessionID]
	return c, ok
}

func (c *wsClient) setRoom(slug string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomSlug = slug
}

func (c *wsClient) clearRoom() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomSlug = ""
}

func (c *wsClient) currentRoom() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.roomSlug
}
