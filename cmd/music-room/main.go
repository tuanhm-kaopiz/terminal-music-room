// music-room is the Terminal Music Room CLI/TUI client for Ubuntu.
package main

import (
	"fmt"
	"os"

	"github.com/terminal-music-room/music-room/internal/client/cli"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(version)
		return
	}
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
