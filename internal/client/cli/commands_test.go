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

func resetCLIGlobals() {
	configPath = ""
	loginName = ""
	loginServer = ""
	joinTUI = true
	joinRepl = false
	// Commands keep their context across invocations. Some tests set command contexts
	// (e.g. createCmd.SetContext), and the stored context can be canceled after the
	// test finishes, causing flakes in later tests.
	RootCmd.SetContext(context.Background())
	createCmd.SetContext(context.Background())
	joinCmd.SetContext(context.Background())
	leaveCmd.SetContext(context.Background())
	loginCmd.SetContext(context.Background())
	RootCmd.SetArgs(nil)
	RootCmd.SetOut(io.Discard)
	RootCmd.SetErr(io.Discard)
	_ = joinCmd.Flags().Set("tui", "true")
	_ = joinCmd.Flags().Set("repl", "false")
	queueImportDryRun = false
	queueImportDelay = 2 * time.Second
	_ = queueImportCmd.Flags().Set("dry-run", "false")
}

func TestJoinFlagDefaults(t *testing.T) {
	tuiFlag := joinCmd.Flags().Lookup("tui")
	if tuiFlag == nil || tuiFlag.DefValue != "true" {
		t.Fatalf("join --tui default = %q, want true", tuiFlag.DefValue)
	}
	replFlag := joinCmd.Flags().Lookup("repl")
	if replFlag == nil || replFlag.DefValue != "false" {
		t.Fatalf("join --repl default = %q, want false", replFlag.DefValue)
	}
}

func TestJoinOneShotNoUI(t *testing.T) {
	_, ts, hostCfg := testHub(t)
	defer ts.Close()
	guestCfg := filepath.Join(t.TempDir(), "guest.yaml")

	// These tests exercise real websocket traffic (client ↔ hub). In CI/macOS runners,
	// the default timeouts can be too tight and cause flaky "context canceled" writes.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, hostCfg, "join-host", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = hostCfg
	RootCmd.SetOut(io.Discard)
	RootCmd.SetErr(io.Discard)
	RootCmd.SetArgs([]string{"create", "join-shot-room"})
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}

	if err := Login(ctx, io.Discard, guestCfg, "join-guest", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = guestCfg

	var out bytes.Buffer
	RootCmd.SetOut(&out)
	RootCmd.SetErr(&out)
	RootCmd.SetArgs([]string{"join", "join-shot-room", "--tui=false", "--repl=false"})
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "joined join-shot-room") {
		t.Fatalf("out %q", out.String())
	}
	resetCLIGlobals()
}

func TestJoinTUICommandRequiresRoom(t *testing.T) {
	_, ts, cfgPath := testHub(t)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, cfgPath, "tui-user", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	RootCmd.SetArgs([]string{"tui"})
	err := RootCmd.ExecuteContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "not in a room") {
		t.Fatalf("err %v", err)
	}
	resetCLIGlobals()
}

func TestJoinEnsureRoomForUI(t *testing.T) {
	_, ts, hostCfg := testHub(t)
	defer ts.Close()
	guestCfg := filepath.Join(t.TempDir(), "guest.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, hostCfg, "tui-host", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = hostCfg
	hostRT, err := newRuntimeFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer hostRT.Close()
	if err := hostRT.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}
	if err := hostRT.send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: "tui-room"}); err != nil {
		t.Fatal(err)
	}
	if err := hostRT.waitInRoom("tui-room", defaultWait); err != nil {
		t.Fatal(err)
	}

	if err := Login(ctx, io.Discard, guestCfg, "tui-guest", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = guestCfg
	guestRT, err := newRuntimeFromConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer guestRT.Close()
	if err := guestRT.ensureConnected(ctx); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := guestRT.ensureRoomForUI(ctx, "tui-room", &out); err != nil {
		t.Fatal(err)
	}
	if !guestRT.store.Snapshot().InRoom || guestRT.store.Snapshot().Room.Slug != "tui-room" {
		t.Fatalf("snapshot %+v", guestRT.store.Snapshot())
	}
	if !strings.Contains(out.String(), "joined tui-room") {
		t.Fatalf("out %q", out.String())
	}

	if err := guestRT.ensureRoomForUI(ctx, "tui-room", &out); err != nil {
		t.Fatal(err)
	}
}

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

	// This test uses the CLI runtime against an in-process hub over websockets and can
	// be slow on CI/macOS runners. Keep the timeout generous to avoid flakes.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, cfgPath, "cli-host", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	var createOut bytes.Buffer
	createCmd.SetOut(&createOut)
	createCmd.SetErr(&createOut)
	createCmd.SetContext(ctx)
	if err := runCreate(createCmd, []string{"cli-room"}); err != nil {
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
	resetCLIGlobals()
	t.Cleanup(resetCLIGlobals)
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
