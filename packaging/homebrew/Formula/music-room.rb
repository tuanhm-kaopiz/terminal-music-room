class MusicRoom < Formula
  desc "Synchronized YouTube listening in the terminal"
  homepage "https://github.com/tuanhm-kaopiz/terminal-music-room"
  version "0.3.2"
  license "MIT"

  depends_on "mpv"
  depends_on "yt-dlp"
  depends_on "ffmpeg"

  on_arm do
    url "https://github.com/tuanhm-kaopiz/terminal-music-room/releases/download/v0.3.2/terminal-music-room_0.3.2_darwin_arm64.tar.gz"
    sha256 "9af08fdd6b7684c9fe69e696035fd227b745c0817f1b03413aacea9efd4e1bab"
  end

  on_intel do
    url "https://github.com/tuanhm-kaopiz/terminal-music-room/releases/download/v0.3.2/terminal-music-room_0.3.2_darwin_amd64.tar.gz"
    sha256 "75e35004ee7f4774106617001cd0ed602007eba013b0be5843ea0f373e56eca6"
  end

  def install
    bin.install "music-room"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/music-room --version")
  end
end
