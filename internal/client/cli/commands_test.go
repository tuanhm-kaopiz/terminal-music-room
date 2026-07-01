package cli

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/ws"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/hub"
)

func TestREPLUnknownCommandHint(t *testing.T) {
	rt := &Runtime{store: state.NewStore()}
	if err := rt.store.Apply(mustEnv(t, protocol.MsgRoomSnapshot, protocol.RoomSnapshot{Slug: "r"})); err != nil {
		t.Fatal(err)
	}
	err := ExecuteREPLLine(context.Background(), rt, "/nope", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "/help") {
		t.Fatalf("err %v", err)
	}
}

func TestREPLRequiresSlashPrefix(t *testing.T) {
	rt := &Runtime{store: state.NewStore()}
	err := ExecuteREPLLine(context.Background(), rt, "play x", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "/help") {
		t.Fatalf("err %v", err)
	}
}

func TestREPLHelp(t *testing.T) {
	var out bytes.Buffer
	rt := &Runtime{store: state.NewStore()}
	if err := ExecuteREPLLine(context.Background(), rt, "/help", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "/play") || !strings.Contains(out.String(), "/chat") {
		t.Fatalf("help %q", out.String())
	}
}

func TestCreateJoinLeaveCommands(t *testing.T) {
	srv, ts, cfgPath := testHub(t)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, cfgPath, "cli-host", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	var createOut bytes.Buffer
	RootCmd.SetOut(&createOut)
	RootCmd.SetErr(&createOut)
	RootCmd.SetArgs([]string{"create", "cli-room"})
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(createOut.String(), "cli-room") {
		t.Fatalf("create out %q", createOut.String())
	}

	rt, err := newRuntimeFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}
	if !rt.store.Snapshot().InRoom {
		t.Fatal("expected in room after create")
	}

	leaveCmd.SetOut(io.Discard)
	RootCmd.SetArgs([]string{"leave"})
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}
	_ = srv
}

func TestJoinREPLChat(t *testing.T) {
	_, ts, cfgPath := testHub(t)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, cfgPath, "guest", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	rt, err := newRuntimeFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}
	if err := rt.send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: "repl-room"}); err != nil {
		t.Fatal(err)
	}
	if err := rt.waitInRoom("repl-room", defaultWait); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	input := strings.NewReader("/chat hello team\n/quit\n")
	if err := RunREPL(ctx, rt, input, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "sent") {
		t.Fatalf("out %q", out.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, msg := range rt.store.Snapshot().Room.Chat {
			if msg.Body == "hello team" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("chat %+v", rt.store.Snapshot().Room.Chat)
}

func TestNotLoggedIn(t *testing.T) {
	configPath = filepath.Join(t.TempDir(), "missing.yaml")
	RootCmd.SetArgs([]string{"create", "room"})
	if err := RootCmd.ExecuteContext(context.Background()); err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Fatalf("err %v", err)
	}
}

func testHub(t *testing.T) (*hub.Server, *httptest.Server, string) {
	t.Helper()
	t.Setenv("MUSIC_ROOM_NO_PLAYBACK", "1")
	srv := hub.New(hub.Config{ListenAddr: ":0", DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ts := httptest.NewServer(srv.Handler())
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	return srv, ts, cfgPath
}

func mustEnv(t *testing.T, msgType string, payload any) protocol.Envelope {
	t.Helper()
	env, err := protocol.NewEnvelope(msgType, "t", payload)
	if err != nil {
		t.Fatal(err)
	}
	return env
}

func TestRuntimeUsesWSClient(t *testing.T) {
	_, ts, cfgPath := testHub(t)
	defer ts.Close()
	ctx := context.Background()
	if err := Login(ctx, io.Discard, cfgPath, "u", ts.URL); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load(cfgPath)
	rt := &Runtime{
		cfgPath: cfgPath,
		cfg:     cfg,
		newClient: func(c ws.Config) WSClient {
			return ws.New(c)
		},
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}
	if rt.store.Snapshot().Status != state.StatusConnected {
		t.Fatalf("status %q", rt.store.Snapshot().Status)
	}
}
