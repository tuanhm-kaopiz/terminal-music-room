package deps

import (
	"fmt"
	"os/exec"
	"strings"
)

// Required binaries for client playback (mpv uses yt-dlp via --ytdl).
var Required = []string{"mpv", "yt-dlp"}

// CheckResult reports missing binaries and OS-specific install hints.
type CheckResult struct {
	Missing []string
	Hints   map[string]string
}

var lookPath = exec.LookPath

// Check returns binaries not found on PATH.
func Check() CheckResult {
	var missing []string
	hints := make(map[string]string)
	for _, name := range Required {
		if _, err := lookPath(name); err != nil {
			missing = append(missing, name)
			hints[name] = installHint(name)
		}
	}
	return CheckResult{Missing: missing, Hints: hints}
}

// FormatError builds a user-facing message for missing dependencies.
func FormatError(r CheckResult) error {
	if len(r.Missing) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("missing playback dependencies: ")
	b.WriteString(strings.Join(r.Missing, ", "))
	b.WriteString("\n")
	for _, name := range r.Missing {
		if hint := r.Hints[name]; hint != "" {
			fmt.Fprintf(&b, "  %s: %s\n", name, hint)
		}
	}
	return fmt.Errorf("%s", strings.TrimSuffix(b.String(), "\n"))
}

// EnsurePlayback returns an error when required binaries are missing.
func EnsurePlayback() error {
	return FormatError(Check())
}
