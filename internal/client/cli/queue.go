package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

var (
	queueURL    string
	queueQuery  string
	queueAfter  string
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage the room queue",
}

var queueAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a track to the queue",
	RunE:  runQueueAddCmd,
}

var queueRemoveCmd = &cobra.Command{
	Use:   "remove <item_id>",
	Short: "Remove a queue item (host only)",
	Args:  cobra.ExactArgs(1),
	RunE:  runQueueRemoveCmd,
}

var queueReorderCmd = &cobra.Command{
	Use:   "reorder <item_id>",
	Short: "Reorder a queue item (host only)",
	Args:  cobra.ExactArgs(1),
	RunE:  runQueueReorderCmd,
}

func init() {
	queueAddCmd.Flags().StringVar(&queueURL, "url", "", "YouTube URL")
	queueAddCmd.Flags().StringVar(&queueQuery, "query", "", "YouTube search query")
	queueReorderCmd.Flags().StringVar(&queueAfter, "after", "", "place after item id")
	queueCmd.AddCommand(queueAddCmd, queueRemoveCmd, queueReorderCmd)
	RootCmd.AddCommand(queueCmd)
}

func runQueueAddCmd(cmd *cobra.Command, _ []string) error {
	payload, err := actions.QueueAddPayload(queueURL, queueQuery)
	if err != nil {
		return err
	}
	ctx := commandContext(cmd)
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		if err := rt.send(ctx, protocol.MsgQueueAdd, payload); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "queue add requested")
		return nil
	})
}

func runQueueRemoveCmd(cmd *cobra.Command, args []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.QueueRemove(ctx, args[0])
	})
}

func runQueueReorderCmd(cmd *cobra.Command, args []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.QueueReorder(ctx, args[0], queueAfter)
	})
}

func runQueueAddArgs(ctx context.Context, rt *Runtime, args []string) error {
	url, query, err := parseSourceArgs(args)
	if err != nil {
		return err
	}
	if url != "" {
		return rt.Actions().QueueAdd(ctx, url)
	}
	return rt.Actions().QueueAdd(ctx, query)
}

func runQueueRemoveArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /queue remove <item_id>")
	}
	return rt.Actions().QueueRemove(ctx, args[0])
}

func runQueueReorderArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /queue reorder <item_id> [after <id>]")
	}
	afterID := ""
	if len(args) >= 3 && strings.EqualFold(args[1], "after") {
		afterID = args[2]
	}
	return rt.Actions().QueueReorder(ctx, args[0], afterID)
}
