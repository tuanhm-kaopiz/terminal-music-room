package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/tui"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

var joinRepl bool
var joinTUI bool

var createCmd = &cobra.Command{
	Use:   "create <slug>",
	Short: "Create a room",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

var joinCmd = &cobra.Command{
	Use:   "join <slug>",
	Short: "Join a room",
	Args:  cobra.ExactArgs(1),
	RunE:  runJoin,
}

var leaveCmd = &cobra.Command{
	Use:   "leave",
	Short: "Leave the current room",
	RunE:  runLeaveCmd,
}

func init() {
	joinCmd.Flags().BoolVar(&joinRepl, "repl", true, "start interactive REPL after joining")
	joinCmd.Flags().BoolVar(&joinTUI, "tui", false, "start Bubble Tea TUI after joining")
	RootCmd.AddCommand(createCmd, joinCmd, leaveCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	slug := args[0]
	ctx := commandContext(cmd)
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.send(ctx, protocol.MsgRoomCreate, protocol.RoomCreatePayload{Slug: slug}); err != nil {
			return err
		}
		if err := rt.waitInRoom(slug, defaultWait); err != nil {
			if msg := rt.lastServerError(); msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("create room: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created and joined %s\n", slug)
		return nil
	})
}

func runJoin(cmd *cobra.Command, args []string) error {
	slug := args[0]
	ctx := commandContext(cmd)
	rt, err := newRuntimeFromConfig()
	if err != nil {
		return err
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		return err
	}
	if err := rt.send(ctx, protocol.MsgRoomJoin, protocol.RoomJoinPayload{Slug: slug}); err != nil {
		return err
	}
	if err := rt.waitInRoom(slug, defaultWait); err != nil {
		if msg := rt.lastServerError(); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("join room: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "joined %s\n", slug)
	if joinTUI {
		return tui.Run(ctx, tui.Config{
			Store: rt.store,
			Send:  rt.send,
			Leave: func(ctx context.Context) error {
				return runLeave(ctx, rt, nil)
			},
		})
	}
	if joinRepl {
		return RunREPL(ctx, rt, cmd.InOrStdin(), cmd.OutOrStdout())
	}
	return nil
}

func runLeaveCmd(cmd *cobra.Command, _ []string) error {
	ctx := commandContext(cmd)
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		return runLeave(ctx, rt, cmd.OutOrStdout())
	})
}

func runLeave(ctx context.Context, rt *Runtime, out io.Writer) error {
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	if err := rt.send(ctx, protocol.MsgRoomLeave, protocol.RoomLeavePayload{}); err != nil {
		return err
	}
	if err := rt.finishLeave(defaultWait); err != nil {
		return err
	}
	if out != nil {
		fmt.Fprintln(out, "left room")
	}
	return nil
}
