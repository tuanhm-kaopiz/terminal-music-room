package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/queuecsv"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

const queueAddWait = 15 * time.Second

var (
	queueImportDelay  time.Duration
	queueImportDryRun bool
)

var queueImportCmd = &cobra.Command{
	Use:   "import <file.csv>",
	Short: "Import YouTube URLs from a CSV file into the queue",
	Long: `Import tracks from a CSV file with a single "url" column.

Example file (playlist.csv):

  url
  https://www.youtube.com/watch?v=abc123
  https://youtu.be/xyz789

Requires an active room session (create or join first).`,
	Args: cobra.ExactArgs(1),
	RunE: runQueueImportCmd,
}

func init() {
	queueImportCmd.Flags().DurationVar(&queueImportDelay, "delay", 2*time.Second, "pause between queue adds")
	queueImportCmd.Flags().BoolVar(&queueImportDryRun, "dry-run", false, "parse file and print URLs without sending")
	queueCmd.AddCommand(queueImportCmd)
}

func runQueueImportCmd(cmd *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	urls, err := queuecsv.ParseURLs(f)
	if err != nil {
		return err
	}

	if queueImportDryRun {
		for i, u := range urls {
			fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\n", i+1, u)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: %d url(s), nothing sent\n", len(urls))
		return nil
	}

	ctx := commandContext(cmd)
	return withRuntime(ctx, func(ctx context.Context, rt *Runtime) error {
		if err := rt.requireInRoom(); err != nil {
			return err
		}
		return rt.importQueueURLs(ctx, cmd, urls)
	})
}

func (r *Runtime) importQueueURLs(ctx context.Context, cmd *cobra.Command, urls []string) error {
	out := cmd.OutOrStdout()
	ok, failed := 0, 0
	for i, url := range urls {
		fmt.Fprintf(out, "[%d/%d] adding %s\n", i+1, len(urls), url)
		if err := r.queueAddURLAndWait(ctx, url); err != nil {
			failed++
			fmt.Fprintf(out, "  failed: %v\n", err)
			continue
		}
		ok++
		if i < len(urls)-1 && queueImportDelay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(queueImportDelay):
			}
		}
	}
	fmt.Fprintf(out, "import done: %d ok, %d failed (of %d)\n", ok, failed, len(urls))
	if ok == 0 {
		return fmt.Errorf("no tracks were added")
	}
	return nil
}

func (r *Runtime) queueAddURLAndWait(ctx context.Context, url string) error {
	prevLen := len(r.store.Snapshot().Room.Queue)
	prevErr := r.lastServerError()

	if err := r.send(ctx, protocol.MsgQueueAdd, protocol.QueueAddPayload{URL: url}); err != nil {
		return err
	}

	deadline := time.Now().Add(queueAddWait)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		v := r.store.Snapshot()
		if len(v.Room.Queue) > prevLen {
			return nil
		}
		if msg := r.lastServerError(); msg != "" && msg != prevErr {
			return fmt.Errorf("%s", msg)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for queue update")
}
