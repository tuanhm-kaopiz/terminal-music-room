#!/usr/bin/env bash
# Build .deb packages for music-room (client) and music-roomd (server).
# Usage: ./packaging/build-deb.sh [version]
# Requires: go, dpkg-deb
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-${VERSION:-0.1.0}}"
REV=1
DIST="${ROOT}/dist"
BUILD="${DIST}/deb-build"

rm -rf "${BUILD}"
mkdir -p "${DIST}"

build_deb() {
	local pkg=$1
	local cmd=$2
	local depends=$3
	local recommends=$4
	local desc=$5

	local root="${BUILD}/${pkg}_${VERSION}-${REV}_amd64"
	mkdir -p "${root}/DEBIAN" "${root}/usr/bin"

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-trimpath -ldflags="-s -w -X main.version=${VERSION}" \
		-o "${root}/usr/bin/${cmd}" "./cmd/${cmd}"

	cat >"${root}/DEBIAN/control" <<EOF
Package: ${pkg}
Version: ${VERSION}-${REV}
Section: sound
Priority: optional
Architecture: amd64
Depends: ${depends}
Recommends: ${recommends}
Maintainer: Terminal Music Room <noreply@example.com>
Homepage: https://github.com/tuanhm-kaopiz/terminal-music-room
Description: ${desc}
EOF

	dpkg-deb --build --root-owner-group "${root}" "${DIST}/${pkg}_${VERSION}-${REV}_amd64.deb"
}

build_deb music-room music-room "mpv, yt-dlp" ffmpeg \
	"Terminal Music Room CLI/TUI client"
build_deb music-roomd music-roomd "yt-dlp" ffmpeg \
	"Terminal Music Room sync server"

echo "Built:"
ls -1 "${DIST}"/*.deb
