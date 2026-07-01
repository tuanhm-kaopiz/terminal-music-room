package ws

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/hub"
)

func TestClientConnectAndDispatchSnapshot(t *testing.T) {
	srv := hub.New(hub.Config{ListenAddr: ":0", DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	store := state.NewStore()
	client := New(Config{
		ServerURL: ts.URL,
		Nickname:  "host",
		Store:     store,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.connectOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if v := store.Snapshot(); v.SessionID == "" || v.DisplayName != "host" {
		t.Fatalf("session %+v", v)
	}

	corrID, err := client.Send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: "ws-room"})
	if err != nil {
		t.Fatal(err)
	}
	if corrID == "" || len(corrID) != 32 {
		t.Fatalf("correlation id %q", corrID)
	}

	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	go func() { _ = client.readLoop(readCtx) }()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		v := store.Snapshot()
		if v.InRoom && v.Room.Slug == "ws-room" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("store %+v", store.Snapshot())
}

func TestReconnectBackoff(t *testing.T) {
	var sleeps []time.Duration
	var mu sync.Mutex
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	attempts := 0

	client := New(Config{
		ServerURL:          "http://localhost:1",
		SessionID:          "deadbeef",
		Store:              state.NewStore(),
		MinReconnectDelay:  time.Second,
		MaxReconnectDelay:  30 * time.Second,
		MaxReconnectWindow: 10 * time.Second,
		Now:                func() time.Time { return now },
		Sleep: func(ctx context.Context, d time.Duration) error {
			mu.Lock()
			sleeps = append(sleeps, d)
			mu.Unlock()
			return ctx.Err()
		},
		Dial: func(context.Context, string, http.Header) (Conn, error) {
			attempts++
			return nil, errors.New("dial failed")
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := client.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(sleeps) == 0 {
		t.Fatal("expected backoff sleeps")
	}
	if sleeps[0] != time.Second {
		t.Fatalf("first sleep %v", sleeps[0])
	}
	if attempts < 2 {
		t.Fatalf("attempts %d", attempts)
	}
}

func TestReconnectExpired(t *testing.T) {
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	now := start
	client := New(Config{
		ServerURL:          "http://localhost:1",
		SessionID:          "abc",
		Store:              state.NewStore(),
		MinReconnectDelay:  time.Millisecond,
		MaxReconnectDelay:  time.Millisecond,
		MaxReconnectWindow: 5 * time.Millisecond,
		Now:                func() time.Time { return now },
		Sleep: func(context.Context, time.Duration) error {
			now = now.Add(2 * time.Millisecond)
			return nil
		},
		Dial: func(context.Context, string, http.Header) (Conn, error) {
			return nil, errors.New("down")
		},
	})
	client.store.Apply(mustEnv(t, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "gone"}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := client.Run(ctx)
	if !errors.Is(err, ErrReconnectExpired) {
		t.Fatalf("err %v", err)
	}
	if v := client.store.Snapshot(); v.InRoom {
		t.Fatalf("room should be cleared %+v", v)
	}
}

func TestReconnectRestoresSnapshot(t *testing.T) {
	srv := hub.New(hub.Config{ListenAddr: ":0", DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, data, err := conn1.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, ack, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
	if err != nil {
		t.Fatal(err)
	}
	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "reconnect-ws"})
	if err := conn1.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	for {
		_, d, err := conn1.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		env, err := protocol.Decode(d)
		if err != nil {
			continue
		}
		if env.Type == protocol.MsgRoomSnapshot {
			break
		}
	}
	_ = conn1.Close(websocket.StatusNormalClosure, "bye")
	time.Sleep(50 * time.Millisecond)

	store := state.NewStore()
	client := New(Config{
		ServerURL: ts.URL,
		SessionID: ack.SessionID,
		Store:     store,
	})
	if err := client.connectOnce(ctx); err != nil {
		t.Fatal(err)
	}

	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()
	go func() { _ = client.readLoop(readCtx) }()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		v := store.Snapshot()
		if v.InRoom && v.Room.Slug == "reconnect-ws" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("store %+v", store.Snapshot())
}

func mustEnv(t *testing.T, msgType string, payload any) protocol.Envelope {
	t.Helper()
	env, err := protocol.NewEnvelope(msgType, "t", payload)
	if err != nil {
		t.Fatal(err)
	}
	return env
}
