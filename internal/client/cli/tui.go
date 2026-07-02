package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/tui"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [slug]",
	Short: "Open the sci-fi room TUI",
	Long:  "Open the Bubble Tea HUD for the current room, or join the given slug first.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTUICmd,
}

func init() {
	RootCmd.AddCommand(tuiCmd)
}

func runTUICmd(cmd *cobra.Command, args []string) error {
	ctx := commandContext(cmd)
	rt, err := newRuntimeFromConfig()
	if err != nil {
		return err
	}
	defer rt.Close()
	if err := rt.ensureConnected(ctx); err != nil {
		return err
	}
	slug := ""
	if len(args) == 1 {
		slug = args[0]
	}
	if slug != "" {
		v := rt.store.Snapshot()
		if v.InRoom && v.Room.Slug == slug {
			// already in target room
		} else if roomPassword != "" {
			if err := rt.ensureRoomForUI(ctx, slug, cmd.OutOrStdout()); err != nil {
				return err
			}
		} else {
			rt.startLocalPlayback(ctx)
			return tui.Run(ctx, tui.Config{
				Store:           rt.store,
				Actions:         rt.Actions(),
				PendingJoinSlug: slug,
				Leave: func(ctx context.Context) error {
					return runLeave(ctx, rt, nil)
				},
			})
		}
	} else if err := rt.requireInRoom(); err != nil {
		return err
	}
	rt.startLocalPlayback(ctx)
	return launchTUI(ctx, rt)
}

func launchTUI(ctx context.Context, rt *Runtime) error {
	return tui.Run(ctx, tui.Config{
		Store:   rt.store,
		Actions: rt.Actions(),
		Leave: func(ctx context.Context) error {
			return runLeave(ctx, rt, nil)
		},
	})
}

// ensureRoomForUI joins slug when needed, or requires an active room session.
func (r *Runtime) ensureRoomForUI(ctx context.Context, slug string, out io.Writer) error {
	if slug != "" {
		v := r.store.Snapshot()
		if v.InRoom && v.Room.Slug == slug {
			return nil
		}
		if err := r.send(ctx, protocol.MsgRoomJoin, protocol.RoomJoinPayload{Slug: slug, Password: roomPassword}); err != nil {
			return err
		}
		if err := r.waitInRoom(slug, defaultWait); err != nil {
			if msg := r.lastServerError(); msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("join room: %w", err)
		}
		if out != nil {
			fmt.Fprintf(out, "joined %s\n", slug)
		}
		return nil
	}
	return r.requireInRoom()
}
