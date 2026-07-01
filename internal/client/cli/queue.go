package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
	payload, err := buildQueueAddPayload(queueURL, queueQuery)
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
	return sendInRoom(cmd, protocol.MsgQueueRemove, protocol.QueueRemovePayload{ItemID: args[0]})
}

func runQueueReorderCmd(cmd *cobra.Command, args []string) error {
	return sendInRoom(cmd, protocol.MsgQueueReorder, protocol.QueueReorderPayload{
		ItemID:  args[0],
		AfterID: queueAfter,
	})
}

func runQueueAddArgs(ctx context.Context, rt *Runtime, args []string) error {
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	raw, err := parseSourceArgs(args)
	if err != nil {
		return err
	}
	payload := protocol.QueueAddPayload{}
	if u, ok := raw["url"]; ok {
		payload.URL = u
	}
	if q, ok := raw["query"]; ok {
		payload.Query = q
	}
	return rt.send(ctx, protocol.MsgQueueAdd, payload)
}

func runQueueRemoveArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /queue remove <item_id>")
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	return rt.send(ctx, protocol.MsgQueueRemove, protocol.QueueRemovePayload{ItemID: args[0]})
}

func runQueueReorderArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /queue reorder <item_id> [after <id>]")
	}
	afterID := ""
	if len(args) >= 3 && strings.EqualFold(args[1], "after") {
		afterID = args[2]
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	return rt.send(ctx, protocol.MsgQueueReorder, protocol.QueueReorderPayload{
		ItemID:  args[0],
		AfterID: afterID,
	})
}

func buildQueueAddPayload(url, query string) (protocol.QueueAddPayload, error) {
	url = stringsTrim(url)
	query = stringsTrim(query)
	switch {
	case url != "" && query != "":
		return protocol.QueueAddPayload{}, fmt.Errorf("use either --url or --query")
	case url != "":
		return protocol.QueueAddPayload{URL: url}, nil
	case query != "":
		return protocol.QueueAddPayload{Query: query}, nil
	default:
		return protocol.QueueAddPayload{}, fmt.Errorf("provide --url or --query")
	}
}
