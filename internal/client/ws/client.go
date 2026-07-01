package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

const (
	defaultMinReconnect = time.Second
	defaultMaxReconnect = 30 * time.Second
	defaultReconnectWin = 5 * time.Minute
)

var (
	// ErrNotConnected is returned when Send is called without an active connection.
	ErrNotConnected = errors.New("websocket not connected")
	// ErrReconnectExpired is returned when reconnect attempts exceed the 5-minute window (AC-050).
	ErrReconnectExpired = errors.New("reconnect window expired")
)

// Conn is a WebSocket connection abstraction for tests.
type Conn interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	Close(code websocket.StatusCode, reason string) error
}

type wsConn struct {
	*websocket.Conn
}

func (c wsConn) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.Conn.Read(ctx)
	return data, err
}

func (c wsConn) Write(ctx context.Context, data []byte) error {
	return c.Conn.Write(ctx, websocket.MessageText, data)
}

func (c wsConn) Close(code websocket.StatusCode, reason string) error {
	return c.Conn.Close(code, reason)
}

// Config configures the WebSocket client.
type Config struct {
	ServerURL string
	SessionID string
	Nickname  string
	Store     *state.Store

	MinReconnectDelay  time.Duration
	MaxReconnectDelay  time.Duration
	MaxReconnectWindow time.Duration

	Dial  func(ctx context.Context, url string, hdr http.Header) (Conn, error)
	Now   func() time.Time
	Sleep func(context.Context, time.Duration) error
}

// Client maintains a WebSocket session with automatic reconnect (AC-048).
type Client struct {
	cfg   Config
	store *state.Store

	connMu sync.Mutex
	conn   Conn

	disconnectAt time.Time
	backoff      time.Duration
}

// New creates a Client. Store defaults to a new store when nil.
func New(cfg Config) *Client {
	if cfg.Store == nil {
		cfg.Store = state.NewStore()
	}
	if cfg.MinReconnectDelay <= 0 {
		cfg.MinReconnectDelay = defaultMinReconnect
	}
	if cfg.MaxReconnectDelay <= 0 {
		cfg.MaxReconnectDelay = defaultMaxReconnect
	}
	if cfg.MaxReconnectWindow <= 0 {
		cfg.MaxReconnectWindow = defaultReconnectWin
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleepContext
	}
	if cfg.Dial == nil {
		cfg.Dial = defaultDial
	}
	return &Client{cfg: cfg, store: cfg.Store}
}

// Store returns the shared state store.
func (c *Client) Store() *state.Store {
	return c.store
}

// Run connects, reads server events into the store, and reconnects with backoff until ctx is done or the reconnect window expires.
func (c *Client) Run(ctx context.Context) error {
	c.disconnectAt = time.Time{}
	for {
		if err := c.connectOnce(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if c.disconnectAt.IsZero() {
				c.disconnectAt = c.cfg.Now()
			}
			if expiredErr := c.waitReconnect(ctx); expiredErr != nil {
				return expiredErr
			}
			continue
		}

		err := c.readLoop(ctx)
		c.closeConn()

		if ctx.Err() != nil {
			return ctx.Err()
		}
		if c.disconnectAt.IsZero() {
			c.disconnectAt = c.cfg.Now()
		}
		if expiredErr := c.waitReconnect(ctx); expiredErr != nil {
			return expiredErr
		}
		if err != nil {
			_ = err
		}
	}
}

func (c *Client) connectOnce(ctx context.Context) error {
	c.store.SetStatus(state.StatusConnecting)
	wsURL, err := config.WebSocketURL(c.cfg.ServerURL)
	if err != nil {
		c.store.SetStatus(state.StatusDisconnected)
		return err
	}
	hdr := http.Header{}
	if c.cfg.SessionID != "" {
		hdr.Set("X-Session-Id", c.cfg.SessionID)
	} else if c.cfg.Nickname != "" {
		hdr.Set("X-Nickname", c.cfg.Nickname)
	}
	conn, err := c.cfg.Dial(ctx, wsURL, hdr)
	if err != nil {
		c.store.SetStatus(state.StatusDisconnected)
		return fmt.Errorf("connect: %w", err)
	}

	data, err := conn.Read(ctx)
	if err != nil {
		_ = conn.Close(websocket.StatusInternalError, "ack read failed")
		c.store.SetStatus(state.StatusDisconnected)
		return fmt.Errorf("read session ack: %w", err)
	}
	env, err := protocol.Decode(data)
	if err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid ack")
		c.store.SetStatus(state.StatusDisconnected)
		return fmt.Errorf("decode session ack: %w", err)
	}
	if env.Type != protocol.MsgSessionAck {
		_ = conn.Close(websocket.StatusPolicyViolation, "expected session.ack")
		c.store.SetStatus(state.StatusDisconnected)
		return fmt.Errorf("expected session.ack, got %q", env.Type)
	}
	if err := c.store.Apply(env); err != nil {
		_ = conn.Close(websocket.StatusInternalError, "apply ack failed")
		c.store.SetStatus(state.StatusDisconnected)
		return err
	}

	c.setConn(conn)
	c.resetBackoff()
	c.disconnectAt = time.Time{}
	c.store.SetStatus(state.StatusConnected)
	return nil
}

func (c *Client) readLoop(ctx context.Context) error {
	for {
		data, err := c.read(ctx)
		if err != nil {
			return err
		}
		env, err := protocol.Decode(data)
		if err != nil {
			continue
		}
		_ = c.store.Apply(env)
	}
}

func (c *Client) read(ctx context.Context) ([]byte, error) {
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()
	if conn == nil {
		return nil, ErrNotConnected
	}
	return conn.Read(ctx)
}

// Send encodes a client message with a correlation ID and writes it to the server.
func (c *Client) Send(ctx context.Context, msgType string, payload any) (string, error) {
	id, err := newCorrelationID()
	if err != nil {
		return "", err
	}
	data, err := protocol.EncodeMessage(msgType, id, payload)
	if err != nil {
		return "", err
	}
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()
	if conn == nil {
		return "", ErrNotConnected
	}
	if err := conn.Write(ctx, data); err != nil {
		return "", err
	}
	return id, nil
}

// Close shuts down the active WebSocket connection.
func (c *Client) Close() error {
	c.closeConn()
	c.store.SetStatus(state.StatusDisconnected)
	return nil
}

func (c *Client) waitReconnect(ctx context.Context) error {
	if !c.shouldReconnect() {
		c.store.ClearRoom()
		c.store.SetStatus(state.StatusDisconnected)
		return ErrReconnectExpired
	}
	delay := c.nextBackoff()
	c.store.SetStatus(state.StatusReconnecting)
	if err := c.cfg.Sleep(ctx, delay); err != nil {
		return err
	}
	return nil
}

func (c *Client) shouldReconnect() bool {
	if c.disconnectAt.IsZero() {
		return true
	}
	return c.cfg.Now().Sub(c.disconnectAt) < c.cfg.MaxReconnectWindow
}

func (c *Client) nextBackoff() time.Duration {
	if c.backoff <= 0 {
		c.backoff = c.cfg.MinReconnectDelay
		return c.backoff
	}
	next := c.backoff * 2
	if next > c.cfg.MaxReconnectDelay {
		next = c.cfg.MaxReconnectDelay
	}
	c.backoff = next
	return c.backoff
}

func (c *Client) resetBackoff() {
	c.backoff = 0
}

func (c *Client) setConn(conn Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.conn = conn
}

func (c *Client) closeConn() {
	c.connMu.Lock()
	conn := c.conn
	c.conn = nil
	c.connMu.Unlock()
	if conn != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "client closing")
	}
}

func defaultDial(ctx context.Context, url string, hdr http.Header) (Conn, error) {
	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: hdr})
	if err != nil {
		return nil, err
	}
	return wsConn{conn}, nil
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func newCorrelationID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
