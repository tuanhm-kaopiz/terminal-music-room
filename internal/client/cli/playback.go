package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

const defaultPlayWait = 90 * time.Second

var (
	playURL    string
	playQuery  string
	playDetach bool
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
	playCmd.Flags().BoolVar(&playDetach, "detach", false, "return after playback starts (default: keep listening until Ctrl+C)")
	RootCmd.AddCommand(playCmd, pauseCmd, resumeCmd, skipCmd, seekCmd)
}

func runPlayCmd(cmd *cobra.Command, _ []string) error {
	ctx := commandContext(cmd)
	payload, err := actions.PlaybackPlayPayload(playURL, playQuery)
	if err != nil {
		return err
	}
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		rt.startLocalPlayback(ctx)
		if err := rt.send(ctx, protocol.MsgPlaybackPlay, payload); err != nil {
			return err
		}
		if err := waitForPlayback(rt.store, defaultPlayWait, protocol.PlaybackPlaying); err != nil {
			if msg := rt.lastServerError(); msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("wait for playing: %w", err)
		}
		track := rt.store.Snapshot().Room.Playback.Track
		if track != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "playing %q\n", track.Title)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "playing")
		}
		if playDetach {
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "listening — press Ctrl+C to stop")
		<-ctx.Done()
		return nil
	})
}

func waitForPlayback(store *state.Store, timeout time.Duration, want protocol.PlaybackStatus) error {
	return waitFor(store, timeout, func(v state.View) bool {
		return v.InRoom && v.Room.Playback.Status == want
	})
}

func runPauseCmd(cmd *cobra.Command, _ []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.Pause(ctx)
	})
}

func runResumeCmd(cmd *cobra.Command, _ []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.Resume(ctx)
	})
}

func runSkipCmd(cmd *cobra.Command, _ []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.Skip(ctx)
	})
}

func runSeekCmd(cmd *cobra.Command, args []string) error {
	return actionInRoom(cmd, func(ctx context.Context, room *actions.Room) error {
		return room.SeekFromString(ctx, args[0])
	})
}
