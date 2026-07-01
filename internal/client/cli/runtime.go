package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/client/state"
	clientsync "github.com/terminal-music-room/music-room/internal/client/sync"
	"github.com/terminal-music-room/music-room/internal/client/ws"
)

var (
	// ErrNotLoggedIn is returned when config has no saved session.
	ErrNotLoggedIn = errors.New("not logged in — run: music-room login --name <nickname>")
)

const defaultWait = 5 * time.Second

// WSClient is the subset of ws.Client used by the CLI runtime.
type WSClient interface {
	Run(ctx context.Context) error
	Send(ctx context.Context, msgType string, payload any) (string, error)
	Close() error
	Store() *state.Store
}

// Runtime holds an authenticated CLI session to music-roomd.
type Runtime struct {
	cfgPath string
	cfg     config.Config
	store   *state.Store
	client  WSClient

	cancel context.CancelFunc
	wg     sync.WaitGroup

	playbackMu       sync.Mutex
	playbackCancel   context.CancelFunc
	playbackWg       sync.WaitGroup
	playbackDriver   clientsync.Driver

	newClient func(ws.Config) WSClient
}

func newRuntimeFromConfig() (*Runtime, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	if !cfg.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	return &Runtime{cfgPath: path, cfg: cfg}, nil
}

// Close stops the WebSocket client and local playback.
func (r *Runtime) Close() {
	r.stopLocalPlayback()
	if r.cancel != nil {
		r.cancel()
	}
	if r.client != nil {
		_ = r.client.Close()
	}
	r.wg.Wait()
}

func (r *Runtime) ensureConnected(ctx context.Context) error {
	if r.client != nil {
		if r.store.Snapshot().Status == state.StatusConnected {
			return nil
		}
	}
	r.store = state.NewStore()
	runCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	factory := r.newClient
	if factory == nil {
		factory = func(cfg ws.Config) WSClient { return ws.New(cfg) }
	}
	r.client = factory(ws.Config{
		ServerURL: r.cfg.ServerURL,
		SessionID: r.cfg.SessionID,
		Store:     r.store,
	})

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		_ = r.client.Run(runCtx)
	}()

	if err := waitFor(r.store, defaultWait, func(v state.View) bool {
		return v.Status == state.StatusConnected
	}); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return nil
}

func (r *Runtime) send(ctx context.Context, msgType string, payload any) error {
	if r.client == nil {
		return ws.ErrNotConnected
	}
	_, err := r.client.Send(ctx, msgType, payload)
	return err
}

func (r *Runtime) waitInRoom(slug string, timeout time.Duration) error {
	return waitFor(r.store, timeout, func(v state.View) bool {
		return v.InRoom && v.Room.Slug == slug
	})
}

func (r *Runtime) waitLeftRoom(timeout time.Duration) error {
	return waitFor(r.store, timeout, func(v state.View) bool {
		return !v.InRoom
	})
}

func (r *Runtime) finishLeave(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if msg := r.lastServerError(); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		time.Sleep(20 * time.Millisecond)
	}
	r.store.ClearRoom()
	return nil
}

func (r *Runtime) requireInRoom() error {
	if !r.store.Snapshot().InRoom {
		return fmt.Errorf("not in a room — run: music-room join <slug>")
	}
	return nil
}

func (r *Runtime) lastServerError() string {
	v := r.store.Snapshot()
	if v.LastErr == nil {
		return ""
	}
	return v.LastErr.Message
}

func waitFor(store *state.Store, timeout time.Duration, ok func(state.View) bool) error {
	if timeout <= 0 {
		timeout = defaultWait
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok(store.Snapshot()) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

func withRuntime(ctx context.Context, fn func(context.Context, *Runtime) error) error {
	rt, err := newRuntimeFromConfig()
	if err != nil {
		return err
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		return err
	}
	return fn(ctx, rt)
}

func commandContext(cmd interface{ Context() context.Context }) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
