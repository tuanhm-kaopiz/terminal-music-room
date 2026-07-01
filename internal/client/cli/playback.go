package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

var (
	playURL   string
	playQuery string
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play a YouTube URL or search query in the room",
	RunE:  runPlayCmd,
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause playback",
	RunE:  runPauseCmd,
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume playback",
	RunE:  runResumeCmd,
}

var skipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip to the next track",
	RunE:  runSkipCmd,
}

var seekCmd = &cobra.Command{
	Use:   "seek <position_ms>",
	Short: "Seek to a position in milliseconds",
	Args:  cobra.ExactArgs(1),
	RunE:  runSeekCmd,
}

func init() {
	playCmd.Flags().StringVar(&playURL, "url", "", "YouTube URL")
	playCmd.Flags().StringVar(&playQuery, "query", "", "YouTube search query")
	RootCmd.AddCommand(playCmd, pauseCmd, resumeCmd, skipCmd, seekCmd)
}

func runPlayCmd(cmd *cobra.Command, _ []string) error {
	ctx := commandContext(cmd)
	payload, err := buildPlayPayload(playURL, playQuery)
	if err != nil {
		return err
	}
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		if err := rt.send(ctx, protocol.MsgPlaybackPlay, payload); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "play requested")
		return nil
	})
}

func runPauseCmd(cmd *cobra.Command, _ []string) error {
	return sendInRoom(cmd, protocol.MsgPlaybackPause, protocol.PlaybackPausePayload{})
}

func runResumeCmd(cmd *cobra.Command, _ []string) error {
	return sendInRoom(cmd, protocol.MsgPlaybackResume, protocol.PlaybackResumePayload{})
}

func runSkipCmd(cmd *cobra.Command, _ []string) error {
	return sendInRoom(cmd, protocol.MsgPlaybackSkip, protocol.PlaybackSkipPayload{})
}

func runSeekCmd(cmd *cobra.Command, args []string) error {
	ms, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || ms < 0 {
		return fmt.Errorf("invalid position %q — use milliseconds", args[0])
	}
	return sendInRoom(cmd, protocol.MsgPlaybackSeek, protocol.PlaybackSeekPayload{PositionMs: ms})
}

func buildPlayPayload(url, query string) (protocol.PlaybackPlayPayload, error) {
	url = stringsTrim(url)
	query = stringsTrim(query)
	switch {
	case url != "" && query != "":
		return protocol.PlaybackPlayPayload{}, fmt.Errorf("use either --url or --query")
	case url != "":
		return protocol.PlaybackPlayPayload{URL: url}, nil
	case query != "":
		return protocol.PlaybackPlayPayload{Query: query}, nil
	default:
		return protocol.PlaybackPlayPayload{}, fmt.Errorf("provide --url or --query")
	}
}

func sendInRoom(cmd *cobra.Command, msgType string, payload any) error {
	ctx := commandContext(cmd)
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		return rt.send(ctx, msgType, payload)
	})
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}
