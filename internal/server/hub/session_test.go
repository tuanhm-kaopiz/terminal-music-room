package hub

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func newTestServer(limiter *Limiter) *Server {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := New(Config{ListenAddr: ":0"}, log)
	if limiter != nil {
		s.limiter = limiter
	}
	return s
}

func dialWS(t *testing.T, ts *httptest.Server, headers http.Header) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(websocket.StatusNormalClosure, "done") })
	return conn
}

func readSessionAck(t *testing.T, conn *websocket.Conn) protocol.SessionAckPayload {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read ack: %v", err)
	}
	env, payload, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
	if err != nil {
		t.Fatalf("decode ack: %v", err)
	}
	if env.Type != protocol.MsgSessionAck {
		t.Fatalf("type %q", env.Type)
	}
	return payload
}

func TestSessionHelloViaHeader(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	conn := dialWS(t, ts, http.Header{"X-Nickname": []string{"kaopiz"}})
	ack := readSessionAck(t, conn)
	if ack.DisplayName != "kaopiz" || ack.SessionID == "" {
		t.Fatalf("ack %+v", ack)
	}
}

func TestSessionHelloViaMessage(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	conn := dialWS(t, ts, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hello, err := protocol.EncodeMessage(protocol.MsgSessionHello, "req-1", protocol.SessionHelloPayload{Nickname: "dev01"})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageText, hello); err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, conn)
	if ack.DisplayName != "dev01" || ack.SessionID == "" {
		t.Fatalf("ack %+v", ack)
	}
}

func TestSessionReconnect(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"reconnect"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack1 := readSessionAck(t, conn1)
	_ = conn1.Close(websocket.StatusNormalClosure, "bye")

	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Session-Id": []string{ack1.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")
	ack2 := readSessionAck(t, conn2)
	if ack2.SessionID != ack1.SessionID {
		t.Fatalf("session id changed: %s -> %s", ack1.SessionID, ack2.SessionID)
	}
}

func TestSessionInvalidNickname(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{""}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	readCtx, readCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer readCancel()
	_, data, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	env, payload, err := protocol.DecodePayload[protocol.ErrorPayload](data)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != protocol.MsgError || payload.Code != protocol.ErrInvalidNickname {
		t.Fatalf("got type=%q code=%q", env.Type, payload.Code)
	}
}

func TestSessionMaxPerIP(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	conns := make([]*websocket.Conn, 0, maxSessionsPerIP)
	for i := 0; i < maxSessionsPerIP; i++ {
		conns = append(conns, dialWS(t, ts, http.Header{"X-Nickname": []string{"user"}}))
		readSessionAck(t, conns[i])
	}
	t.Cleanup(func() {
		for _, c := range conns {
			_ = c.Close(websocket.StatusNormalClosure, "done")
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	_, resp2, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"fourth"}},
	})
	if err == nil {
		t.Fatal("expected fourth connection to fail")
	}
	if resp2 != nil && resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status %d", resp2.StatusCode)
	}
}

func TestSessionRateLimitConnect(t *testing.T) {
	lim := NewLimiter(RateLimitConfig{ConnectPerMinute: 2, CreateRoomPerHour: 5, ChatPerMinute: 20})
	s := newTestServer(lim)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	dialWS(t, ts, http.Header{"X-Nickname": []string{"a"}})
	dialWS(t, ts, http.Header{"X-Nickname": []string{"b"}})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	_, resp, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"c"}},
	})
	if err == nil {
		t.Fatal("expected rate limit")
	}
	if resp != nil && resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestLimiterCreateRoomAndChat(t *testing.T) {
	lim := NewLimiter(RateLimitConfig{ConnectPerMinute: 10, CreateRoomPerHour: 1, ChatPerMinute: 1})
	ip := "127.0.0.1"
	if !lim.Allow(LimitCreateRoom, ip) {
		t.Fatal("first create should pass")
	}
	if lim.Allow(LimitCreateRoom, ip) {
		t.Fatal("second create should fail")
	}
	if !lim.Allow(LimitChat, ip) {
		t.Fatal("first chat should pass")
	}
	if lim.Allow(LimitChat, ip) {
		t.Fatal("second chat should fail")
	}
}
