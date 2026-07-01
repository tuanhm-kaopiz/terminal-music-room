package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/terminal-music-room/music-room/internal/client/actions"
)

func runPlayArgs(ctx context.Context, rt *Runtime, args []string) error {
	url, query, err := parseSourceArgs(args)
	if err != nil {
		return err
	}
	if url != "" {
		return rt.Actions().Play(ctx, url)
	}
	return rt.Actions().Play(ctx, query)
}

func runSeekArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /seek <position_ms>")
	}
	return rt.Actions().SeekFromString(ctx, args[0])
}

func runChatArgs(ctx context.Context, rt *Runtime, args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /chat <message>")
	}
	body := strings.Join(args, " ")
	if err := rt.Actions().Chat(ctx, body); err != nil {
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
	return rt.Actions().React(ctx, args[0])
}

func parseSourceArgs(args []string) (url, query string, err error) {
	return actions.ParseSourceArgs(args)
}
