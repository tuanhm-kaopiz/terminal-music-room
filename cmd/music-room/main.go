// music-room is the Terminal Music Room CLI/TUI client.
package main

import (
	"fmt"
	"os"

	"github.com/terminal-music-room/music-room/internal/client/cli"
)

// version is set at link time: -ldflags "-X main.version=0.2.1"
var version = "dev"

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
