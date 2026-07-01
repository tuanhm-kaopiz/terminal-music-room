#!/usr/bin/env bash
# Upgrade music-room client (and optional music-roomd) from GitHub Releases.
# Usage:
#   ./scripts/upgrade.sh tarball [VERSION]   # default VERSION=0.2.0
#   ./scripts/upgrade.sh deb [VERSION]
#   INSTALL_SERVER=1 ./scripts/upgrade.sh tarball   # also upgrade music-roomd
set -euo pipefail

MODE="${1:-tarball}"
VERSION="${2:-${VERSION:-0.2.0}}"
REPO="${GITHUB_REPO:-tuanhm-kaopiz/terminal-music-room}"
BASE="https://github.com/${REPO}/releases/download/v${VERSION}"
INSTALL_SERVER="${INSTALL_SERVER:-0}"

log() { printf '==> %s\n' "$*"; }
fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || fail "missing dependency: $1"
}

install_dir() {
	if [[ -n "${INSTALL_DIR:-}" ]]; then
		echo "$INSTALL_DIR"
		return
	fi
	if command -v music-room >/dev/null 2>&1; then
		dirname "$(command -v music-room)"
		return
	fi
	echo "${HOME}/.local/bin"
}

upgrade_tarball() {
	need_cmd curl
	need_cmd tar
	need_cmd sha256sum

	local dest shell_wrap
	dest="$(install_dir)"
	mkdir -p "$dest"

	local tmpdir archive
	tmpdir="$(mktemp -d)"
	trap 'rm -rf "${tmpdir}"' EXIT

	archive="terminal-music-room_${VERSION}_linux_amd64.tar.gz"
	log "downloading v${VERSION} tarball"
	curl -fsSL -o "${tmpdir}/${archive}" "${BASE}/${archive}"

	log "verifying SHA256SUMS"
	tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"
	(cd "$tmpdir" && sha256sum -c SHA256SUMS)

	log "installing music-room → ${dest}/"
	cp "${tmpdir}/music-room" "${dest}/music-room"
	chmod +x "${dest}/music-room"

	if [[ "$INSTALL_SERVER" == "1" ]]; then
		log "installing music-roomd → ${dest}/"
		cp "${tmpdir}/music-roomd" "${dest}/music-roomd"
		chmod +x "${dest}/music-roomd"
	fi

	if ! command -v music-room >/dev/null 2>&1; then
		shell_wrap=0
		case ":${PATH}:" in
			*":${dest}:"*) shell_wrap=1 ;;
		esac
		if [[ "$shell_wrap" -eq 0 ]]; then
			log "add to PATH: export PATH=\"${dest}:\$PATH\""
		fi
	fi
}

upgrade_deb() {
	need_cmd curl
	need_cmd sudo

	local tmpdir deb_client deb_server
	tmpdir="$(mktemp -d)"
	trap 'rm -rf "${tmpdir}"' EXIT

	deb_client="music-room_${VERSION}-1_amd64.deb"
	log "downloading ${deb_client}"
	curl -fsSL -o "${tmpdir}/${deb_client}" "${BASE}/${deb_client}"

	log "installing client package (sudo)"
	sudo dpkg -i "${tmpdir}/${deb_client}"

	if [[ "$INSTALL_SERVER" == "1" ]]; then
		deb_server="music-roomd_${VERSION}-1_amd64.deb"
		log "downloading ${deb_server}"
		curl -fsSL -o "${tmpdir}/${deb_server}" "${BASE}/${deb_server}"
		log "installing server package (sudo)"
		sudo dpkg -i "${tmpdir}/${deb_server}"
	fi
}

case "$MODE" in
	tarball|tar|tgz) upgrade_tarball ;;
	deb|dpkg) upgrade_deb ;;
	-h|--help|help)
		cat <<EOF
Usage: $0 [tarball|deb] [VERSION]

  tarball   download release tarball, verify checksums, copy to INSTALL_DIR
  deb       download .deb and sudo dpkg -i

Environment:
  VERSION=${VERSION}     release tag without leading v
  INSTALL_DIR=           target bin dir (default: dirname of music-room, else ~/.local/bin)
  INSTALL_SERVER=1       also upgrade music-roomd
  GITHUB_REPO=           owner/repo (default: ${REPO})

Examples:
  $0 tarball 0.2.0
  INSTALL_SERVER=1 $0 deb
EOF
		exit 0
		;;
	*) fail "unknown mode: ${MODE} (use tarball or deb)" ;;
esac

if command -v music-room >/dev/null 2>&1; then
	log "done — $(music-room --version 2>/dev/null || music-room version 2>/dev/null || echo 'music-room installed')"
else
	log "done — run: music-room --version"
fi

log "config unchanged: ~/.config/music-room/config.yaml (no re-login needed)"
