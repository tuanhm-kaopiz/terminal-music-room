package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errREPLExit = errors.New("repl exit")

// RunREPL reads slash-commands until EOF or /quit.
func RunREPL(ctx context.Context, rt *Runtime, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	fmt.Fprintln(out, "music-room repl — /help for commands")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := ExecuteREPLLine(ctx, rt, line, out); err != nil {
			if errors.Is(err, errREPLExit) {
				return nil
			}
			fmt.Fprintf(out, "error: %v\n", err)
		}
	}
	return scanner.Err()
}

// ExecuteREPLLine handles one REPL input line (AC-052 hints on invalid input).
func ExecuteREPLLine(ctx context.Context, rt *Runtime, line string, out io.Writer) error {
	if !strings.HasPrefix(line, "/") {
		return fmt.Errorf("commands start with / — try /help")
	}
	body := strings.TrimSpace(strings.TrimPrefix(line, "/"))
	if body == "" {
		return fmt.Errorf("empty command — try /help")
	}
	parts := strings.Fields(body)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "?":
		printREPLHelp(out)
		return nil
	case "quit", "exit":
		return errREPLExit
	case "leave":
		return runLeave(ctx, rt, out)
	case "play":
		return runPlayArgs(ctx, rt, args)
	case "pause":
		return rt.Actions().Pause(ctx)
	case "resume":
		return rt.Actions().Resume(ctx)
	case "skip":
		return rt.Actions().Skip(ctx)
	case "seek":
		return runSeekArgs(ctx, rt, args)
	case "queue":
		return runQueueREPL(ctx, rt, args)
	case "chat":
		return runChatArgs(ctx, rt, args, out)
	case "vote":
		return runVoteREPL(ctx, rt, args)
	case "react":
		return runReactArgs(ctx, rt, args)
	default:
		return fmt.Errorf("unknown command %q — try /help", cmd)
	}
}

func printREPLHelp(out io.Writer) {
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  /play <url|search query>")
	fmt.Fprintln(out, "  /pause  /resume  /skip  /seek <ms>")
	fmt.Fprintln(out, "  /queue add <url|query>  |  /queue remove <id>  |  /queue reorder <id> [after <id>]")
	fmt.Fprintln(out, "  /chat <message>")
	fmt.Fprintln(out, "  /vote skip  |  /vote priority <item_id>")
	fmt.Fprintln(out, "  /react <emoji>")
	fmt.Fprintln(out, "  /leave  /quit")
}

func runQueueREPL(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /queue add|remove|reorder — try /help")
	}
	switch strings.ToLower(args[0]) {
	case "add":
		return runQueueAddArgs(ctx, rt, args[1:])
	case "remove":
		return runQueueRemoveArgs(ctx, rt, args[1:])
	case "reorder":
		return runQueueReorderArgs(ctx, rt, args[1:])
	default:
		return fmt.Errorf("unknown /queue subcommand %q — try /help", args[0])
	}
}

func runVoteREPL(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /vote skip|priority — try /help")
	}
	switch strings.ToLower(args[0]) {
	case "skip":
		return rt.Actions().VoteSkip(ctx)
	case "priority":
		return runVotePriorityArgs(ctx, rt, args[1:])
	default:
		return fmt.Errorf("unknown /vote subcommand %q — try /help", args[0])
	}
}
