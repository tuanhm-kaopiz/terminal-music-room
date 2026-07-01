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
    sha256 "7998442c0091e54fa8877cb780e556b86e3e14c2367ab076c5bc06e081f41357"
  end

  on_intel do
    url "https://github.com/tuanhm-kaopiz/terminal-music-room/releases/download/v0.2.1/terminal-music-room_0.2.1_darwin_amd64.tar.gz"
    sha256 "7cba5039c083e3a63625444b4bc91c3fdb9a0137c134f9ac5b0b9792257b6fdb"
  end

  def install
    bin.install "music-room"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/music-room --version")
  end
end
