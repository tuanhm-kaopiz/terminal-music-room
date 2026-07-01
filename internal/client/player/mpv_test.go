package player

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type mockIPC struct {
	mu    sync.Mutex
	calls [][]any

	timePos float64
	paused  bool
	closed  bool
}

func (m *mockIPC) Call(_ context.Context, command []any) (json.RawMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrNotRunning
	}
	m.calls = append(m.calls, command)
	if len(command) >= 2 && command[0] == "get_property" && command[1] == "time-pos" {
		data, _ := json.Marshal(m.timePos)
		return data, nil
	}
	if len(command) >= 3 && command[0] == "set_property" && command[1] == "pause" {
		if paused, ok := command[2].(bool); ok {
			m.paused = paused
		}
	}
	if len(command) >= 3 && command[0] == "seek" {
		if sec, ok := command[1].(float64); ok {
			m.timePos = sec
		}
	}
	return json.RawMessage(`null`), nil
}

func (m *mockIPC) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockIPC) lastCommands() [][]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([][]any, len(m.calls))
	copy(out, m.calls)
	return out
}

func TestYouTubeURL(t *testing.T) {
	got := YouTubeURL("abc123xyz01")
	want := "https://www.youtube.com/watch?v=abc123xyz01"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestMPVArgs(t *testing.T) {
	args := MPVArgs("/tmp/mpv.sock")
	if len(args) != 7 {
		t.Fatalf("args %v", args)
	}
	if args[len(args)-1] != "--ytdl-format=bestaudio" {
		t.Fatalf("missing ytdl format %v", args)
	}
	if args[4] != "--input-ipc-server=/tmp/mpv.sock" {
		t.Fatalf("ipc arg %q", args[4])
	}
}

func TestStartUsesExpectedArgs(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "mpv.sock")
	var capturedArgs []string
	ipc := &mockIPC{}

	p := New(Config{
		MPVPath: "mpv",
		WorkDir: dir,
		execCommand: func(name string, arg ...string) *exec.Cmd {
			if name != "mpv" {
				t.Fatalf("binary %q", name)
			}
			capturedArgs = append([]string(nil), arg...)
			return exec.Command("sleep", "3600")
		},
		waitSocket: func(path string, _ time.Duration) error {
			if path != socketPath {
				t.Fatalf("socket path %q", path)
			}
			return nil
		},
		dialIPC: func(path string) (ipcClient, error) {
			if path != socketPath {
				t.Fatalf("dial path %q", path)
			}
			return ipc, nil
		},
		killProcess: func(*exec.Cmd) error { return nil },
	})

	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.Running() {
		t.Fatal("expected running")
	}
	wantArgs := MPVArgs(socketPath)
	for i, arg := range wantArgs {
		if capturedArgs[i] != arg {
			t.Fatalf("arg[%d] got %q want %q", i, capturedArgs[i], arg)
		}
	}
	_ = p.Stop()
}

func TestPlayPauseSeekPosition(t *testing.T) {
	ipc := &mockIPC{timePos: 42.0}
	p := newTestPlayer(t, ipc)

	ctx := context.Background()
	if err := p.Play(ctx, "abc123xyz01"); err != nil {
		t.Fatal(err)
	}
	if err := p.Pause(ctx); err != nil {
		t.Fatal(err)
	}
	if !ipc.paused {
		t.Fatal("expected paused")
	}
	if err := p.Resume(ctx); err != nil {
		t.Fatal(err)
	}
	if ipc.paused {
		t.Fatal("expected resumed")
	}
	pos, err := p.PositionMs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pos != 42_000 {
		t.Fatalf("pos %d", pos)
	}
	if err := p.Seek(ctx, 90_000); err != nil {
		t.Fatal(err)
	}
	if ipc.timePos != 90.0 {
		t.Fatalf("seek pos %v", ipc.timePos)
	}

	calls := ipc.lastCommands()
	if calls[0][0] != "loadfile" || calls[0][1] != YouTubeURL("abc123xyz01") {
		t.Fatalf("loadfile call %+v", calls[0])
	}
}

func TestPlayRequiresVideoID(t *testing.T) {
	p := newTestPlayer(t, &mockIPC{})
	if err := p.Play(context.Background(), ""); err != ErrInvalidVideoID {
		t.Fatalf("err %v", err)
	}
}

func TestStopKillsProcess(t *testing.T) {
	dir := t.TempDir()
	killed := false
	ipc := &mockIPC{}

	p := New(Config{
		WorkDir: dir,
		execCommand: func(string, ...string) *exec.Cmd {
			return exec.Command("sleep", "3600")
		},
		waitSocket: func(string, time.Duration) error { return nil },
		dialIPC:    func(string) (ipcClient, error) { return ipc, nil },
		killProcess: func(*exec.Cmd) error {
			killed = true
			return nil
		},
	})
	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if p.Running() {
		t.Fatal("expected stopped")
	}
	if !ipc.closed {
		t.Fatal("ipc should close")
	}
	if !killed {
		t.Fatal("expected mpv process killed")
	}
}

func TestNotRunningErrors(t *testing.T) {
	p := New(Config{})
	ctx := context.Background()
	if err := p.Play(ctx, "abc123xyz01"); err != ErrNotRunning {
		t.Fatalf("play err %v", err)
	}
	if _, err := p.PositionMs(ctx); err != ErrNotRunning {
		t.Fatalf("pos err %v", err)
	}
}

func newTestPlayer(t *testing.T, ipc ipcClient) *Player {
	t.Helper()
	dir := t.TempDir()
	p := New(Config{
		WorkDir: dir,
		execCommand: func(string, ...string) *exec.Cmd {
			return exec.Command("sleep", "3600")
		},
		waitSocket:  func(string, time.Duration) error { return nil },
		dialIPC:     func(string) (ipcClient, error) { return ipc, nil },
		killProcess: func(*exec.Cmd) error { return nil },
	})
	if err := p.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Stop() })
	return p
}

func TestSocketIPCCall(t *testing.T) {
	clientConn, serverConn := netPipe(t)
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		buf := make([]byte, 4096)
		_, _ = serverConn.Read(buf)
		_, _ = serverConn.Write([]byte(`{"request_id":1,"error":"success","data":12.5}` + "\n"))
	}()

	ipc := newSocketIPC(clientConn)
	data, err := ipc.Call(context.Background(), []any{"get_property", "time-pos"})
	if err != nil {
		t.Fatal(err)
	}
	var sec float64
	if err := json.Unmarshal(data, &sec); err != nil || sec != 12.5 {
		t.Fatalf("data %s err %v", data, err)
	}
	_ = ipc.Close()
}

func netPipe(t *testing.T) (*netConn, *netConn) {
	t.Helper()
	left, right := make(chan []byte, 8), make(chan []byte, 8)
	return &netConn{ch: left, peer: right}, &netConn{ch: right, peer: left}
}

type netConn struct {
	ch     chan []byte
	peer   chan []byte
	closed bool
	mu     sync.Mutex
}

func (n *netConn) Read(b []byte) (int, error) {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return 0, os.ErrClosed
	}
	n.mu.Unlock()
	data, ok := <-n.ch
	if !ok {
		return 0, os.ErrClosed
	}
	return copy(b, data), nil
}

func (n *netConn) Write(b []byte) (int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return 0, os.ErrClosed
	}
	buf := append([]byte(nil), b...)
	select {
	case n.peer <- buf:
		return len(b), nil
	default:
		return 0, os.ErrClosed
	}
}

func (n *netConn) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.closed {
		n.closed = true
		close(n.ch)
	}
	return nil
}

func (n *netConn) LocalAddr() net.Addr                { return pipeAddr{} }
func (n *netConn) RemoteAddr() net.Addr               { return pipeAddr{} }
func (n *netConn) SetDeadline(time.Time) error        { return nil }
func (n *netConn) SetReadDeadline(time.Time) error    { return nil }
func (n *netConn) SetWriteDeadline(time.Time) error { return nil }

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

var _ net.Conn = (*netConn)(nil)
