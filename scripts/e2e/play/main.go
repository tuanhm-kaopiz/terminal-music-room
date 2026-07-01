// Command play sends playback.play and waits for playing state (e2e smoke).
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
	url := flag.String("url", "", "YouTube URL to play (required)")
	timeout := flag.Duration("timeout", 90*time.Second, "wait for playing")
	flag.Parse()

	if *cfgPath == "" || strings.TrimSpace(*url) == "" {
		fmt.Fprintln(os.Stderr, "usage: play --config <path> --url <youtube-url>")
		os.Exit(2)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fail(err)
	}
	if !cfg.LoggedIn() {
		fail(fmt.Errorf("config not logged in"))
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
	defer conn.Close(websocket.StatusNormalClosure, "play done")

	_ = readSessionAck(ctx, conn)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](ctx, conn)

	playID := "e2e-play"
	data, err := protocol.EncodeMessage(protocol.MsgPlaybackPlay, playID, protocol.PlaybackPlayPayload{URL: *url})
	if err != nil {
		fail(err)
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		fail(fmt.Errorf("send play: %w", err))
	}

	state, tick, err := readPlaybackUpdate(ctx, conn)
	if err != nil {
		fail(err)
	}
	if state.Status != protocol.PlaybackPlaying || state.Track == nil {
		fail(fmt.Errorf("unexpected state %+v", state))
	}
	if tick.Status != protocol.PlaybackPlaying {
		fail(fmt.Errorf("unexpected tick %+v", tick))
	}
	fmt.Printf("ok: playing %q\n", state.Track.Title)
}

func readSessionAck(ctx context.Context, conn *websocket.Conn) protocol.SessionAckPayload {
	_, data, err := conn.Read(ctx)
	if err != nil {
		fail(fmt.Errorf("read session ack: %w", err))
	}
	env, ack, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
	if err != nil || env.Type != protocol.MsgSessionAck {
		fail(fmt.Errorf("expected session.ack, got %q", env.Type))
	}
	return ack
}

func readEnvelope[T any](ctx context.Context, conn *websocket.Conn) (protocol.Envelope, T, error) {
	var zero T
	_, data, err := conn.Read(ctx)
	if err != nil {
		return protocol.Envelope{}, zero, err
	}
	return protocol.DecodePayload[T](data)
}

func readPlaybackUpdate(ctx context.Context, conn *websocket.Conn) (protocol.PlaybackState, protocol.PlaybackTickPayload, error) {
	_, state, err := readEnvelope[protocol.PlaybackState](ctx, conn)
	if err != nil {
		return protocol.PlaybackState{}, protocol.PlaybackTickPayload{}, err
	}
	_, tick, err := readEnvelope[protocol.PlaybackTickPayload](ctx, conn)
	if err != nil {
		return protocol.PlaybackState{}, protocol.PlaybackTickPayload{}, err
	}
	return state, tick, nil
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "e2e play: %v\n", err)
	os.Exit(1)
}
