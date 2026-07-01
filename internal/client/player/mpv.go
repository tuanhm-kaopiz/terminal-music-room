package player

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMPVPath      = "mpv"
	socketWaitTimeout   = 5 * time.Second
	socketPollInterval  = 50 * time.Millisecond
	defaultIPCCallTimeout = 3 * time.Second
)

var (
	// ErrNotRunning is returned when the player is not started.
	ErrNotRunning = errors.New("mpv player not running")
	// ErrInvalidVideoID is returned for blank video IDs.
	ErrInvalidVideoID = errors.New("video id is required")
)

// Config configures mpv subprocess and IPC behavior.
type Config struct {
	MPVPath string
	WorkDir string

	execCommand func(name string, arg ...string) *exec.Cmd
	dialIPC     func(path string) (ipcClient, error)
	waitSocket  func(path string, timeout time.Duration) error
	killProcess func(cmd *exec.Cmd) error
}

// Player drives a local mpv subprocess for YouTube audio playback (ADR-002).
type Player struct {
	cfg Config

	mu         sync.Mutex
	cmd        *exec.Cmd
	ipc        ipcClient
	socketPath string
	running    bool
}

// New creates a Player with defaults.
func New(cfg Config) *Player {
	if cfg.MPVPath == "" {
		cfg.MPVPath = defaultMPVPath
	}
	if cfg.execCommand == nil {
		cfg.execCommand = exec.Command
	}
	if cfg.dialIPC == nil {
		cfg.dialIPC = dialSocketIPC
	}
	if cfg.waitSocket == nil {
		cfg.waitSocket = waitForSocket
	}
	if cfg.killProcess == nil {
		cfg.killProcess = killProcess
	}
	return &Player{cfg: cfg}
}

// YouTubeURL builds a watch URL from a YouTube video ID.
func YouTubeURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

// Start spawns mpv with IPC, ytdl, and bestaudio format.
func (p *Player) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		return nil
	}

	workDir := p.cfg.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "music-room-mpv-*")
		if err != nil {
			return fmt.Errorf("create mpv work dir: %w", err)
		}
	}
	socketPath := filepath.Join(workDir, "mpv.sock")
	args := MPVArgs(socketPath)

	cmd := p.cfg.execCommand(p.cfg.MPVPath, args...)
	cmd.Dir = workDir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, socketWaitTimeout)
	defer cancel()
	if err := p.cfg.waitSocket(socketPath, time.Until(deadlineFromCtx(waitCtx))); err != nil {
		_ = p.cfg.killProcess(cmd)
		return fmt.Errorf("wait for mpv ipc socket: %w", err)
	}

	ipc, err := p.cfg.dialIPC(socketPath)
	if err != nil {
		_ = p.cfg.killProcess(cmd)
		return fmt.Errorf("connect mpv ipc: %w", err)
	}

	p.cmd = cmd
	p.ipc = ipc
	p.socketPath = socketPath
	p.running = true
	return nil
}

// Stop kills mpv and closes IPC (AC-011: stop audio on leave).
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stopLocked()
}

func (p *Player) stopLocked() error {
	if !p.running {
		return nil
	}
	if p.ipc != nil {
		_ = p.ipc.Close()
		p.ipc = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cfg.killProcess(p.cmd)
	}
	if p.socketPath != "" {
		_ = os.Remove(p.socketPath)
	}
	p.cmd = nil
	p.socketPath = ""
	p.running = false
	return nil
}

// Running reports whether mpv is active.
func (p *Player) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// Play loads a YouTube track by video ID and starts playback.
func (p *Player) Play(ctx context.Context, videoID string) error {
	if videoID == "" {
		return ErrInvalidVideoID
	}
	ipc, err := p.ipcOrErr()
	if err != nil {
		return err
	}
	url := YouTubeURL(videoID)
	if _, err := ipc.Call(ctx, []any{"loadfile", url, "replace"}); err != nil {
		return fmt.Errorf("mpv loadfile: %w", err)
	}
	if _, err := ipc.Call(ctx, []any{"set_property", "pause", false}); err != nil {
		return fmt.Errorf("mpv resume after load: %w", err)
	}
	return nil
}

// Pause pauses playback.
func (p *Player) Pause(ctx context.Context) error {
	return p.setPause(ctx, true)
}

// Resume unpauses playback.
func (p *Player) Resume(ctx context.Context) error {
	return p.setPause(ctx, false)
}

func (p *Player) setPause(ctx context.Context, paused bool) error {
	ipc, err := p.ipcOrErr()
	if err != nil {
		return err
	}
	_, err = ipc.Call(ctx, []any{"set_property", "pause", paused})
	if err != nil {
		return fmt.Errorf("mpv set pause: %w", err)
	}
	return nil
}

// Seek moves playback to an absolute position in milliseconds.
func (p *Player) Seek(ctx context.Context, positionMs int64) error {
	ipc, err := p.ipcOrErr()
	if err != nil {
		return err
	}
	seconds := float64(positionMs) / 1000.0
	_, err = ipc.Call(ctx, []any{"seek", seconds, "absolute"})
	if err != nil {
		return fmt.Errorf("mpv seek: %w", err)
	}
	return nil
}

// PositionMs returns the current playback position in milliseconds.
func (p *Player) PositionMs(ctx context.Context) (int64, error) {
	ipc, err := p.ipcOrErr()
	if err != nil {
		return 0, err
	}
	data, err := ipc.Call(ctx, []any{"get_property", "time-pos"})
	if err != nil {
		return 0, fmt.Errorf("mpv get time-pos: %w", err)
	}
	var seconds float64
	if err := json.Unmarshal(data, &seconds); err != nil {
		return 0, fmt.Errorf("decode time-pos: %w", err)
	}
	if seconds < 0 {
		return 0, nil
	}
	return int64(seconds * 1000), nil
}

func (p *Player) ipcOrErr() (ipcClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running || p.ipc == nil {
		return nil, ErrNotRunning
	}
	return p.ipc, nil
}

func waitForSocket(path string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = socketWaitTimeout
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(socketPollInterval)
	}
	return fmt.Errorf("socket %s not ready", path)
}

func dialSocketIPC(path string) (ipcClient, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return newSocketIPC(conn), nil
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
	return nil
}

func deadlineFromCtx(ctx context.Context) time.Time {
	if d, ok := ctx.Deadline(); ok {
		return d
	}
	return time.Now().Add(socketWaitTimeout)
}

// ipcClient sends mpv JSON IPC commands and reads responses.
type ipcClient interface {
	Call(ctx context.Context, command []any) (json.RawMessage, error)
	Close() error
}

type socketIPC struct {
	conn net.Conn

	mu      sync.Mutex
	pending map[uint64]chan ipcResult
	nextID  uint64

	closed atomic.Bool
	done   chan struct{}
}

type ipcResult struct {
	data  json.RawMessage
	error string
}

func newSocketIPC(conn net.Conn) *socketIPC {
	s := &socketIPC{
		conn:    conn,
		pending: make(map[uint64]chan ipcResult),
		done:    make(chan struct{}),
	}
	go s.readLoop()
	return s
}

func (s *socketIPC) Call(ctx context.Context, command []any) (json.RawMessage, error) {
	if s.closed.Load() {
		return nil, errors.New("ipc closed")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, defaultIPCCallTimeout)
	defer cancel()

	id := s.registerPending()
	defer s.dropPending(id)

	msg := map[string]any{
		"command":    command,
		"request_id": id,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	payload = append(payload, '\n')

	s.mu.Lock()
	_, err = s.conn.Write(payload)
	ch := s.pending[id]
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.error != "success" {
			if resp.error == "" {
				return nil, errors.New("mpv ipc command failed")
			}
			return nil, fmt.Errorf("mpv ipc: %s", resp.error)
		}
		return resp.data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *socketIPC) registerPending() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	id := s.nextID
	s.pending[id] = make(chan ipcResult, 1)
	return id
}

func (s *socketIPC) dropPending(id uint64) {
	s.mu.Lock()
	delete(s.pending, id)
	s.mu.Unlock()
}

func (s *socketIPC) readLoop() {
	defer close(s.done)
	scanner := bufio.NewScanner(s.conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var msg struct {
			RequestID uint64          `json:"request_id"`
			Error     string          `json:"error"`
			Data      json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.RequestID == 0 {
			continue
		}
		s.mu.Lock()
		ch, ok := s.pending[msg.RequestID]
		s.mu.Unlock()
		if !ok {
			continue
		}
		ch <- ipcResult{data: msg.Data, error: msg.Error}
	}
}

func (s *socketIPC) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	err := s.conn.Close()
	<-s.done
	return err
}

// MPVArgs returns the argument list used to spawn mpv (for tests).
func MPVArgs(socketPath string) []string {
	return []string{
		"--no-video",
		"--really-quiet",
		"--idle=yes",
		"--keep-open=yes",
		"--input-ipc-server=" + socketPath,
		"--ytdl",
		"--ytdl-format=bestaudio",
	}
}
