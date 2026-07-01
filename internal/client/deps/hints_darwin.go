//go:build darwin

package deps

func installHint(binary string) string {
	switch binary {
	case "mpv", "yt-dlp":
		return "brew install mpv yt-dlp ffmpeg"
	default:
		return "brew install " + binary
	}
}
