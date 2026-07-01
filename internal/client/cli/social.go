package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/actions"
)

var chatCmd = &cobra.Command{
	Use:   "chat <message...>",
	Short: "Send a chat message",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runChatCmd,
}

var voteCmd = &cobra.Command{
	Use:   "vote",
	Short: "Vote on skip or queue priority",
}

var voteSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Vote to skip the current track",
	RunE:  runVoteSkipCmd,
}

var votePriorityCmd = &cobra.Command{
	Use:   "priority <item_id>",
	Short: "Vote to prioritize a queue item",
	Args:  cobra.ExactArgs(1),
	RunE:  runVotePriorityCmd,
}

var reactCmd = &cobra.Command{
	Use:   "react <emoji>",
	Short: "Send an emoji reaction on the current track",
	Args:  cobra.ExactArgs(1),
	RunE:  runReactCmd,
}

func init() {
	voteCmd.AddCommand(voteSkipCmd, votePriorityCmd)
	RootCmd.AddCommand(chatCmd, voteCmd, reactCmd)
}

func runChatCmd(cmd *cobra.Command, args []string) error {
	body := strings.Join(args, " ")
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		if err := room.Chat(ctx, body); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "sent")
		return nil
	})
}

func runVoteSkipCmd(cmd *cobra.Command, _ []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.VoteSkip(ctx)
	})
}

func runVotePriorityCmd(cmd *cobra.Command, args []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.VotePriority(ctx, args[0])
	})
}

func runReactCmd(cmd *cobra.Command, args []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.React(ctx, args[0])
	})
}

func runVotePriorityArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /vote priority <item_id>")
	}
	return rt.Actions().VotePriority(ctx, args[0])
}
