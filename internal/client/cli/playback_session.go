package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/terminal-music-room/music-room/internal/client/deps"
	"github.com/terminal-music-room/music-room/internal/client/player"
	clientsync "github.com/terminal-music-room/music-room/internal/client/sync"
)

func localPlaybackDisabled() bool {
	return os.Getenv("MUSIC_ROOM_NO_PLAYBACK") != ""
}

func ensurePlaybackDeps() error {
	if localPlaybackDisabled() {
		return nil
	}
	return deps.EnsurePlayback()
}

// startLocalPlayback runs the sync engine and mpv until stopLocalPlayback or Close.
func (r *Runtime) startLocalPlayback(ctx context.Context) {
	if localPlaybackDisabled() || r.store == nil {
		return
	}
	r.playbackMu.Lock()
	defer r.playbackMu.Unlock()
	if r.playbackCancel != nil {
		return
	}

	drv := r.playbackDriver
	if drv == nil {
		if err := ensurePlaybackDeps(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		drv = player.New(player.Config{})
		r.playbackDriver = drv
	}

	engine := clientsync.New(clientsync.Config{Store: r.store, Player: drv})
	playbackCtx, cancel := context.WithCancel(context.Background())
	r.playbackCancel = cancel
	r.playbackWg.Add(1)
	go func() {
		defer r.playbackWg.Done()
		_ = engine.Run(playbackCtx)
	}()
}

// stopLocalPlayback stops mpv and the sync engine goroutine.
func (r *Runtime) stopLocalPlayback() {
	r.playbackMu.Lock()
	cancel := r.playbackCancel
	r.playbackCancel = nil
	drv := r.playbackDriver
	r.playbackDriver = nil
	r.playbackMu.Unlock()

	if cancel != nil {
		cancel()
	}
	r.playbackWg.Wait()
	if drv != nil {
		_ = drv.Stop()
	}
}
