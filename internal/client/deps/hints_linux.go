//go:build linux

package deps

func installHint(binary string) string {
	switch binary {
	case "mpv", "yt-dlp":
		return "sudo apt install -y mpv yt-dlp ffmpeg"
	default:
		return "sudo apt install -y " + binary
	}
}
