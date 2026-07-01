// Command wait connects to music-roomd and waits for room snapshot criteria (e2e smoke).
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func main() {
	cfgPath := flag.String("config", "", "client config file (required)")
	members := flag.Int("members", 0, "minimum member count in room snapshot (0 = skip)")
	playbackStatus := flag.String("playback-status", "", "expected playback status (empty = skip)")
	timeout := flag.Duration("timeout", 30*time.Second, "overall wait timeout")
	flag.Parse()

	if *cfgPath == "" {
		fmt.Fprintln(os.Stderr, "usage: wait --config <path> [--members N] [--playback-status playing]")
		os.Exit(2)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fail(err)
	}
	if !cfg.LoggedIn() {
		fail(fmt.Errorf("config not logged in: %s", *cfgPath))
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	wsURL, err := config.WebSocketURL(cfg.ServerURL)
	if err != nil {
		fail(err)
	}

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Session-Id": []string{cfg.SessionID}},
	})
	if err != nil {
		fail(fmt.Errorf("connect: %w", err))
	}
	defer conn.Close(websocket.StatusNormalClosure, "wait done")

	_, data, err := conn.Read(ctx)
	if err != nil {
		fail(fmt.Errorf("read session ack: %w", err))
	}
	env, _, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
	if err != nil || env.Type != protocol.MsgSessionAck {
		fail(fmt.Errorf("expected session.ack, got %q", env.Type))
	}

	deadline := time.Now().Add(*timeout)
	for time.Now().Before(deadline) {
		readCtx, readCancel := context.WithDeadline(ctx, deadline)
		_, data, err := conn.Read(readCtx)
		readCancel()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			continue
		}
		env, err := protocol.Decode(data)
		if err != nil {
			continue
		}
		switch env.Type {
		case protocol.MsgRoomSnapshot:
			var snap protocol.RoomSnapshot
			if err := env.UnmarshalPayload(&snap); err != nil {
				continue
			}
			if snapshotOK(snap, *members, *playbackStatus) {
				fmt.Printf("ok: room=%s members=%d playback=%s\n", snap.Slug, len(snap.Members), snap.Playback.Status)
				return
			}
		case protocol.MsgPlaybackState:
			var state protocol.PlaybackState
			if err := env.UnmarshalPayload(&state); err != nil {
				continue
			}
			if *members == 0 && playbackStatusMatches(state.Status, *playbackStatus) {
				fmt.Printf("ok: playback=%s track=%q\n", state.Status, trackTitle(state.Track))
				return
			}
		}
	}
	fail(fmt.Errorf("timeout waiting for snapshot (members>=%d playback=%q)", *members, *playbackStatus))
}

func snapshotOK(snap protocol.RoomSnapshot, minMembers int, wantStatus string) bool {
	if minMembers > 0 && len(snap.Members) < minMembers {
		return false
	}
	if wantStatus != "" && !playbackStatusMatches(snap.Playback.Status, wantStatus) {
		return false
	}
	if wantStatus == "playing" && snap.Playback.Track == nil {
		return false
	}
	return minMembers > 0 || wantStatus != ""
}

func playbackStatusMatches(got protocol.PlaybackStatus, want string) bool {
	if want == "" {
		return true
	}
	return strings.EqualFold(string(got), strings.TrimSpace(want))
}

func trackTitle(t *protocol.Track) string {
	if t == nil {
		return ""
	}
	return t.Title
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "e2e wait: %v\n", err)
	os.Exit(1)
}
