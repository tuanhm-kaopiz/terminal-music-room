#!/usr/bin/env bash
# Update Homebrew formula url/sha256 from dist macOS tarballs.
# Usage: ./packaging/homebrew/bump-formula.sh <version>
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VERSION="${1:?usage: bump-formula.sh <version>}"
FORMULA="${ROOT}/packaging/homebrew/Formula/music-room.rb"
REPO="tuanhm-kaopiz/terminal-music-room"

arm_tar="${ROOT}/dist/terminal-music-room_${VERSION}_darwin_arm64.tar.gz"
amd_tar="${ROOT}/dist/terminal-music-room_${VERSION}_darwin_amd64.tar.gz"

for f in "$arm_tar" "$amd_tar"; do
	if [[ ! -f "$f" ]]; then
		echo "missing: $f (run ./packaging/build-macos.sh ${VERSION} first)" >&2
		exit 1
	fi
done

arm_sha=$(sha256sum "$arm_tar" | awk '{print $1}')
amd_sha=$(sha256sum "$amd_tar" | awk '{print $1}')

cat >"$FORMULA" <<EOF
class MusicRoom < Formula
  desc "Synchronized YouTube listening in the terminal"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  depends_on "mpv"
  depends_on "yt-dlp"
  depends_on "ffmpeg"

  on_arm do
    url "https://github.com/${REPO}/releases/download/v${VERSION}/terminal-music-room_${VERSION}_darwin_arm64.tar.gz"
    sha256 "${arm_sha}"
  end

  on_intel do
    url "https://github.com/${REPO}/releases/download/v${VERSION}/terminal-music-room_${VERSION}_darwin_amd64.tar.gz"
    sha256 "${amd_sha}"
  end

  def install
    bin.install "music-room"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/music-room --version")
  end
end
EOF

echo "Updated ${FORMULA}"
