class MusicRoom < Formula
  desc "Synchronized YouTube listening in the terminal"
  homepage "https://github.com/tuanhm-kaopiz/terminal-music-room"
  version "0.2.1"
  license "MIT"

  depends_on "mpv"
  depends_on "yt-dlp"
  depends_on "ffmpeg"

  on_arm do
    url "https://github.com/tuanhm-kaopiz/terminal-music-room/releases/download/v0.2.1/terminal-music-room_0.2.1_darwin_arm64.tar.gz"
    sha256 "ffdee4313ddd0c6455a8898f1f969a07b17f98fc6dd4392746388ed171ff5885"
  end

  on_intel do
    url "https://github.com/tuanhm-kaopiz/terminal-music-room/releases/download/v0.2.1/terminal-music-room_0.2.1_darwin_amd64.tar.gz"
    sha256 "096212f37a934636db6172dc7518397a85fa552bc89631492582c81807fa9fda"
  end

  def install
    bin.install "music-room"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/music-room --version")
  end
end
