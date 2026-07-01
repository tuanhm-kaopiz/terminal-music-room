#!/usr/bin/env bash
# TUI v2 smoke: music-roomd + headless sci-fi HUD join check (CI-friendly).
# Requires: Go toolchain, curl, python3. Set MUSIC_ROOM_NO_PLAYBACK=1 (no mpv/yt-dlp).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export MUSIC_ROOM_NO_PLAYBACK="${MUSIC_ROOM_NO_PLAYBACK:-1}"

BIN_DIR="${BIN_DIR:-$ROOT/bin}"
DATA_DIR="$(mktemp -d)"
ROOM_SLUG="tui-smoke-$$"
SERVER_HOST="127.0.0.1"
PORT="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')"
SERVER_URL="http://${SERVER_HOST}:${PORT}"
TUI_TIMEOUT="${TUI_SMOKE_TIMEOUT:-20s}"

HOST_CFG="${DATA_DIR}/host.yaml"
GUEST_CFG="${DATA_DIR}/guest.yaml"
DAEMON_PID=""

log() { printf '==> %s\n' "$*"; }
fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing dependency: $1"
}

cleanup() {
  if [[ -n "${DAEMON_PID}" ]] && kill -0 "${DAEMON_PID}" 2>/dev/null; then
    kill "${DAEMON_PID}" 2>/dev/null || true
    wait "${DAEMON_PID}" 2>/dev/null || true
  fi
  rm -rf "${DATA_DIR}"
}
trap cleanup EXIT

wait_health() {
  local url=$1 tries=${2:-40}
  for _ in $(seq 1 "$tries"); do
    if curl -sf "${url}/healthz" >/dev/null; then
      return 0
    fi
    sleep 0.25
  done
  return 1
}

log "checking dependencies"
need_cmd go
need_cmd curl
need_cmd python3

log "building binaries"
mkdir -p "$BIN_DIR"
go build -o "$BIN_DIR/music-room" ./cmd/music-room
go build -o "$BIN_DIR/music-roomd" ./cmd/music-roomd

log "starting music-roomd on ${SERVER_URL}"
mkdir -p "${DATA_DIR}/chat"
MUSIC_ROOM_LISTEN="${SERVER_HOST}:${PORT}" \
  MUSIC_ROOM_DATA_DIR="${DATA_DIR}/chat" \
  "$BIN_DIR/music-roomd" >/dev/null 2>&1 &
DAEMON_PID=$!

wait_health "$SERVER_URL" || fail "music-roomd did not become healthy"

log "host login + create room ${ROOM_SLUG}"
"$BIN_DIR/music-room" --config "$HOST_CFG" login --name tui-host --server "$SERVER_URL"
"$BIN_DIR/music-room" --config "$HOST_CFG" create "$ROOM_SLUG"

log "guest login"
"$BIN_DIR/music-room" --config "$GUEST_CFG" login --name tui-guest --server "$SERVER_URL"

log "headless TUI join smoke (MUSIC_ROOM_NO_PLAYBACK=${MUSIC_ROOM_NO_PLAYBACK})"
go run ./scripts/e2e/tui/main.go \
  --config "$GUEST_CFG" \
  --room "$ROOM_SLUG" \
  --timeout "$TUI_TIMEOUT"

log "assert room snapshot via waiter"
go run ./scripts/e2e/wait/main.go --config "$GUEST_CFG" --members 2 --timeout 10s

log "tui smoke passed"
