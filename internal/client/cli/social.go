package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/protocol"
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
	ctx := commandContext(cmd)
	body := strings.Join(args, " ")
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		if err := rt.send(ctx, protocol.MsgChatSend, protocol.ChatSendPayload{Body: body}); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "sent")
		return nil
	})
}

func runVoteSkipCmd(cmd *cobra.Command, _ []string) error {
	return sendInRoom(cmd, protocol.MsgVoteSkip, protocol.VoteSkipPayload{})
}

func runVotePriorityCmd(cmd *cobra.Command, args []string) error {
	return sendInRoom(cmd, protocol.MsgVotePriority, protocol.VotePriorityPayload{ItemID: args[0]})
}

func runReactCmd(cmd *cobra.Command, args []string) error {
	return sendInRoom(cmd, protocol.MsgReactionSend, protocol.ReactionSendPayload{Emoji: args[0]})
}

func runVotePriorityArgs(ctx context.Context, rt *Runtime, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: /vote priority <item_id>")
	}
	if err := rt.requireInRoom(); err != nil {
		return err
	}
	return rt.send(ctx, protocol.MsgVotePriority, protocol.VotePriorityPayload{ItemID: args[0]})
}
