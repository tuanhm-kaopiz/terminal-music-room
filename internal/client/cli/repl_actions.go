package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

func runPlayArgs(ctx context.Context, rt *Runtime, args []string) error {
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	raw, err := parseSourceArgs(args)
	if err != nil {
		return err
	}
	payload := protocol.PlaybackPlayPayload{}
	if u, ok := raw["url"]; ok {
		payload.URL = u
	}
	if q, ok := raw["query"]; ok {
		payload.Query = q
	}
	return rt.send(ctx, protocol.MsgPlaybackPlay, payload)
}

func runSeekArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /seek <position_ms>")
	}
	var ms int64
	if _, err := fmt.Sscan(args[0], &ms); err != nil || ms < 0 {
		return fmt.Errorf("invalid position %q — use milliseconds", args[0])
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	return rt.send(ctx, protocol.MsgPlaybackSeek, protocol.PlaybackSeekPayload{PositionMs: ms})
}

func runChatArgs(ctx context.Context, rt *Runtime, args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /chat <message>")
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	body := strings.Join(args, " ")
	if err := rt.send(ctx, protocol.MsgChatSend, protocol.ChatSendPayload{Body: body}); err != nil {
		return err
	}
	if out != nil {
		fmt.Fprintln(out, "sent")
	}
	return nil
}

func runReactArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /react <emoji>")
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	return rt.send(ctx, protocol.MsgReactionSend, protocol.ReactionSendPayload{Emoji: args[0]})
}
