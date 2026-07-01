// Command tui runs a headless join+HUD smoke: connect, join room, render sci-fi HUD, assert panels.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui"
	"github.com/terminal-music-room/music-room/internal/client/ws"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func main() {
	cfgPath := flag.String("config", "", "client config file (required)")
	slug := flag.String("room", "", "room slug to join (required)")
	timeout := flag.Duration("timeout", 20*time.Second, "overall timeout")
	width := flag.Int("width", 80, "terminal width")
	height := flag.Int("height", 24, "terminal height")
	flag.Parse()

	if *cfgPath == "" || *slug == "" {
		fmt.Fprintln(os.Stderr, "usage: tui --config <path> --room <slug>")
		os.Exit(2)
	}

	_ = os.Setenv("MUSIC_ROOM_NO_PLAYBACK", "1")

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fail(err)
	}
	if !cfg.LoggedIn() {
		fail(fmt.Errorf("config not logged in: %s", *cfgPath))
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	store := state.NewStore()
	client := ws.New(ws.Config{
		ServerURL: cfg.ServerURL,
		SessionID: cfg.SessionID,
		Store:     store,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = client.Run(ctx)
	}()
	defer func() {
		cancel()
		_ = client.Close()
		wg.Wait()
	}()

	if err := waitFor(store, 10*time.Second, func(v state.View) bool {
		return v.Status == state.StatusConnected
	}); err != nil {
		fail(fmt.Errorf("connect: %w", err))
	}

	if _, err := client.Send(ctx, protocol.MsgRoomJoin, protocol.RoomJoinPayload{Slug: *slug}); err != nil {
		fail(fmt.Errorf("join: %w", err))
	}
	if err := waitFor(store, 10*time.Second, func(v state.View) bool {
		return v.InRoom && v.Room.Slug == *slug
	}); err != nil {
		fail(fmt.Errorf("wait in room: %w", err))
	}

	act := actions.New(func(ctx context.Context, msgType string, payload any) error {
		_, err := client.Send(ctx, msgType, payload)
		return err
	}, store)

	model := tui.NewModel(ctx, tui.Config{
		Store:   store,
		Actions: act,
	})

	// Exercise Bubble Tea init + resize (join --tui path) without blocking on a TTY.
	program := tea.NewProgram(
		&model,
		tea.WithContext(ctx),
		tea.WithInput(strings.NewReader("q")),
	)
	done := make(chan error, 1)
	go func() {
		_, err := program.Run()
		done <- err
	}()
	time.Sleep(100 * time.Millisecond)
	program.Send(tea.WindowSizeMsg{Width: *width, Height: *height})
	select {
	case err := <-done:
		if err != nil {
			fail(fmt.Errorf("tui program: %w", err))
		}
	case <-time.After(3 * time.Second):
		program.Quit()
		if err := <-done; err != nil {
			fail(fmt.Errorf("tui program: %w", err))
		}
	case <-ctx.Done():
		fail(fmt.Errorf("tui program: %w", ctx.Err()))
	}

	// Assert HUD composition from live room snapshot (AC-009).
	hud := tui.NewModel(ctx, tui.Config{Store: store, Actions: act})
	next, _ := (&hud).Update(tea.WindowSizeMsg{Width: *width, Height: *height})
	rendered := stripANSI(next.View())
	if err := assertHUD(rendered); err != nil {
		fail(err)
	}

	fmt.Printf("ok: tui hud rendered for room=%s (%dx%d)\n", *slug, *width, *height)
}

func assertHUD(s string) error {
	required := []string{"QUEUE", "COMMS", "NOW PLAYING", "ROOM"}
	var missing []string
	for _, want := range required {
		if !strings.Contains(s, want) {
			missing = append(missing, want)
		}
	}
	if len(missing) > 0 {
		if len(s) > 500 {
			s = s[:500] + "..."
		}
		return fmt.Errorf("hud missing %v in output: %q", missing, s)
	}
	return nil
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	esc := false
	for i := 0; i < len(s); i++ {
		if esc {
			if s[i] == 'm' {
				esc = false
			}
			continue
		}
		if s[i] == '\x1b' {
			esc = true
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func waitFor(store *state.Store, timeout time.Duration, ok func(state.View) bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok(store.Snapshot()) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "e2e tui: %v\n", err)
	os.Exit(1)
}
