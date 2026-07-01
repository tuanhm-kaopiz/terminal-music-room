#!/usr/bin/env bash
# End-to-end smoke: music-roomd + two CLI clients, play URL, assert room snapshot.
# Requires: Go toolchain, yt-dlp, mpv (checked), curl, python3 (ephemeral port).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BIN_DIR="${BIN_DIR:-$ROOT/bin}"
DATA_DIR="$(mktemp -d)"
ROOM_SLUG="e2e-smoke-$$"
SERVER_HOST="127.0.0.1"
PORT="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')"
SERVER_URL="http://${SERVER_HOST}:${PORT}"
# Short public-domain style test clip (Me at the Zoo); override with E2E_VIDEO_URL.
VIDEO_URL="${E2E_VIDEO_URL:-https://www.youtube.com/watch?v=jNQXAC9IVRw}"
PLAY_TIMEOUT="${E2E_PLAY_TIMEOUT:-90s}"

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

wait_snapshot() {
  go run ./scripts/e2e/wait/main.go "$@"
}

log "checking dependencies"
need_cmd go
need_cmd curl
need_cmd python3
need_cmd yt-dlp
need_cmd mpv

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
"$BIN_DIR/music-room" --config "$HOST_CFG" login --name e2e-host --server "$SERVER_URL"
"$BIN_DIR/music-room" --config "$HOST_CFG" create "$ROOM_SLUG"

log "guest login + join"
"$BIN_DIR/music-room" --config "$GUEST_CFG" login --name e2e-guest --server "$SERVER_URL"
"$BIN_DIR/music-room" --config "$GUEST_CFG" join "$ROOM_SLUG" --repl=false

log "assert snapshot: 2 members"
wait_snapshot --config "$GUEST_CFG" --members 2 --timeout 15s

log "host play ${VIDEO_URL} (keep connection until playing)"
go run ./scripts/e2e/play/main.go --config "$HOST_CFG" --url "$VIDEO_URL" --timeout "$PLAY_TIMEOUT"

log "assert guest snapshot: playback playing"
wait_snapshot --config "$GUEST_CFG" --playback-status playing --timeout 15s

log "guest chat smoke"
"$BIN_DIR/music-room" --config "$GUEST_CFG" chat "e2e smoke ok"

log "e2e smoke passed"
