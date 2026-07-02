#!/usr/bin/env bash
# Build macOS client tarballs (darwin/arm64 + darwin/amd64).
# Usage: ./packaging/build-macos.sh [version]
# Requires: go
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-${VERSION:-0.1.0}}"
DIST="${ROOT}/dist"
# -B gobuildid: Mach-O LC_UUID required on macOS 15+ / Tahoe (Go <1.24 omits it by default).
LDFLAGS="-s -w -B gobuildid -X main.version=${VERSION}"

rm -rf "${DIST}/macos-build"
mkdir -p "${DIST}"

build_arch() {
	local goarch=$1
	local label=$2
	local build_dir="${DIST}/macos-build/${label}"
	mkdir -p "${build_dir}"

	CGO_ENABLED=0 GOOS=darwin GOARCH="${goarch}" go build \
		-trimpath -ldflags="${LDFLAGS}" \
		-o "${build_dir}/music-room" ./cmd/music-room
	chmod +x "${build_dir}/music-room"
	(
		cd "${build_dir}"
		sha256sum music-room > SHA256SUMS
	)
	tar -C "${build_dir}" -czf "${DIST}/terminal-music-room_${VERSION}_darwin_${label}.tar.gz" music-room SHA256SUMS
}

build_arch arm64 arm64
build_arch amd64 amd64

echo "Built:"
ls -1 "${DIST}"/terminal-music-room_"${VERSION}"_darwin_*.tar.gz
