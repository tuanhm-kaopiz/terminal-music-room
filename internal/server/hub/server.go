package hub

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
	"github.com/terminal-music-room/music-room/internal/server/queuehistory"
	"github.com/terminal-music-room/music-room/internal/server/room"
	"github.com/terminal-music-room/music-room/internal/server/vote"
)

// Server is the HTTP + WebSocket bootstrap for music-roomd.
type Server struct {
	log          *slog.Logger
	cfg          Config
	server       *http.Server
	sessions     *SessionStore
	limiter      *Limiter
	rooms        *room.Manager
	clients      *clientRegistry
	resolver     SourceResolver
	queueHistory *queuehistory.Store
	voteCfg      vote.Config
	reconnectTTL time.Duration
	tickOnce     sync.Once
	tickStop     chan struct{}
}

// New constructs a Server. Pass nil logger to use slog.Default().
func New(cfg Config, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		log:      log,
		cfg:      cfg,
		sessions: NewSessionStore(),
		limiter:  NewLimiter(DefaultRateLimitConfig()),
		rooms: room.NewManager(chat.Options{DataDir: cfg.DataDir}),
		clients:  newClientRegistry(),
		resolver: newDefaultResolver(),
		voteCfg:  vote.DefaultConfig(),
		tickStop: make(chan struct{}),
	}
	if cfg.QueueHistoryDir != "" {
		s.queueHistory = queuehistory.NewStore(cfg.QueueHistoryDir)
	}
	s.server = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s
}

// Handler returns the root HTTP handler (for tests).
func (s *Server) Handler() http.Handler {
	return s.routes()
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	s.startPlaybackTicks()
	s.log.Info("music-roomd listening", "addr", s.cfg.ListenAddr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("music-roomd shutting down")
	close(s.tickStop)
	return s.server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /v1/ws", s.handleWebSocket)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter.Allow(LimitConnect, ip) {
		retry := s.limiter.RetryAfterSeconds(LimitConnect, ip)
		w.Header().Set("Retry-After", strconv.Itoa(retry))
		http.Error(w, "rate limited", http.StatusTooManyRequests)
		return
	}

	sessionID := r.Header.Get("X-Session-Id")
	if sessionID == "" {
		if s.sessions.CountByIP(ip) >= maxSessionsPerIP {
			http.Error(w, "too many sessions", http.StatusTooManyRequests)
			return
		}
	} else if _, ok := s.sessions.Get(sessionID); !ok {
		if s.sessions.CountByIP(ip) >= maxSessionsPerIP {
			http.Error(w, "too many sessions", http.StatusTooManyRequests)
			return
		}
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		s.log.Error("websocket accept failed", "err", err)
		return
	}

	ctx := r.Context()
	sess, err := s.establishSession(ctx, conn, r, ip)
	if err != nil {
		s.log.Debug("session establishment failed", "err", err, "ip", ip)
		return
	}
	client := s.clients.register(sess.ID, conn)
	defer func() {
		s.clients.unregisterConn(sess.ID, conn)
		if _, ok := s.clients.get(sess.ID); !ok {
			s.markSessionDisconnected(sess.ID, time.Now())
		}
	}()
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	if err := s.restoreRoomOnReconnect(ctx, client, conn, sess); err != nil {
		s.log.Debug("reconnect restore failed", "err", err, "session_id", sess.ID)
	}

	s.log.Info("websocket session ready",
		"session_id", sess.ID,
		"nickname", sess.Nickname,
		"ip", ip,
	)

	s.serveWS(ctx, client, sess)
}

func (s *Server) establishSession(ctx context.Context, conn *websocket.Conn, r *http.Request, ip string) (*Session, error) {
	now := time.Now()
	sessionID := r.Header.Get("X-Session-Id")
	nickname := r.Header.Get("X-Nickname")

	if sessionID != "" {
		sess, err := s.resumeSession(sessionID, ip, now)
		if err == nil {
			if err := s.sendSessionAck(ctx, conn, sess, ""); err != nil {
				_ = conn.Close(websocket.StatusInternalError, "ack failed")
				return nil, err
			}
			return sess, nil
		}
	}

	if _, hasNick := r.Header["X-Nickname"]; hasNick {
		sess, err := s.createSession(ip, nickname, now)
		if err != nil {
			_ = s.sendError(ctx, conn, "", protocol.ErrInvalidNickname, err.Error(), nil)
			_ = conn.Close(websocket.StatusPolicyViolation, "invalid nickname")
			return nil, err
		}
		if err := s.sendSessionAck(ctx, conn, sess, ""); err != nil {
			s.sessions.Remove(sess.ID)
			_ = conn.Close(websocket.StatusInternalError, "ack failed")
			return nil, err
		}
		return sess, nil
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "hello required")
		return nil, err
	}
	env, err := protocol.Decode(data)
	if err != nil || env.Type != protocol.MsgSessionHello {
		_ = s.sendError(ctx, conn, env.ID, protocol.ErrInvalidMessage, "first message must be session.hello", nil)
		_ = conn.Close(websocket.StatusPolicyViolation, "hello required")
		return nil, err
	}
	var hello protocol.SessionHelloPayload
	if err := env.UnmarshalPayload(&hello); err != nil {
		_ = s.sendError(ctx, conn, env.ID, protocol.ErrInvalidMessage, "invalid session.hello payload", nil)
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid hello")
		return nil, err
	}

	sess, err := s.createSession(ip, hello.Nickname, now)
	if err != nil {
		_ = s.sendError(ctx, conn, env.ID, protocol.ErrInvalidNickname, err.Error(), nil)
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid nickname")
		return nil, err
	}
	if err := s.sendSessionAck(ctx, conn, sess, env.ID); err != nil {
		s.sessions.Remove(sess.ID)
		_ = conn.Close(websocket.StatusInternalError, "ack failed")
		return nil, err
	}
	return sess, nil
}

func (s *Server) serveWS(ctx context.Context, client *wsClient, sess *Session) {
	conn := client.conn
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || err == io.EOF {
				return
			}
			s.log.Debug("websocket read ended", "err", err, "session_id", sess.ID)
			return
		}
		if len(data) == 0 {
			continue
		}
		env, err := protocol.Decode(data)
		if err != nil {
			_ = s.sendError(ctx, conn, "", protocol.ErrInvalidMessage, "invalid message", nil)
			continue
		}
		if env.Type == protocol.MsgSessionHello {
			_ = s.sendError(ctx, conn, env.ID, protocol.ErrInvalidMessage, "already authenticated", nil)
			continue
		}
		s.dispatchMessage(ctx, client, sess, env)
	}
}

func (s *Server) sendSessionAck(ctx context.Context, conn *websocket.Conn, sess *Session, corrID string) error {
	env, err := protocol.NewEnvelope(protocol.MsgSessionAck, corrID, protocol.SessionAckPayload{
		SessionID:   sess.ID,
		DisplayName: sess.DisplayName,
	})
	if err != nil {
		return err
	}
	return s.writeEnvelope(ctx, conn, env)
}

func (s *Server) sendError(ctx context.Context, conn *websocket.Conn, corrID string, code protocol.ErrorCode, message string, retryAfter *int) error {
	env, err := protocol.NewErrorEnvelope(corrID, code, message, retryAfter)
	if err != nil {
		return err
	}
	return s.writeEnvelope(ctx, conn, env)
}

func (s *Server) writeEnvelope(ctx context.Context, conn *websocket.Conn, env protocol.Envelope) error {
	data, err := protocol.Encode(env)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
