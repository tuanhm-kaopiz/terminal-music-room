package cli

import (
	"github.com/spf13/cobra"
)

var (
	configPath string
)

// RootCmd is the music-room CLI root.
var RootCmd = &cobra.Command{
	Use:   "music-room",
	Short: "Terminal Music Room — sync YouTube playback with your team",
	Long:  "CLI client for Terminal Music Room. Log in, create or join rooms, and listen in sync.",
}

// Execute runs the CLI and returns any error from the invoked command.
func Execute(version string) error {
	RootCmd.Version = version
	RootCmd.SetVersionTemplate("music-room {{.Version}}\n")
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (default ~/.config/music-room/config.yaml)")
	RootCmd.SilenceUsage = true
}
