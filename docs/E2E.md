# End-to-end testing — Terminal Music Room

Automated smoke plus manual checklists for playback sync, voting, and reconnect.

## Prerequisites

| Tool | Purpose |
|------|---------|
| Go 1.22+ | Build `music-room` and `music-roomd` |
| `yt-dlp` | Resolve YouTube URLs on the server |
| `mpv` | Local audio on clients (required by packaging; checked by smoke script) |
| `curl` | Health check in smoke script |

Ubuntu:

```bash
sudo apt update
sudo apt install -y mpv yt-dlp curl
```

## Automated smoke

From the repo root:

```bash
chmod +x scripts/e2e-smoke.sh
./scripts/e2e-smoke.sh
```

What it does:

1. Builds `bin/music-room` and `bin/music-roomd`
2. Starts `music-roomd` on a random local port
3. Host: `login` → `create <slug>`
4. Guest: `login` → `join <slug> --repl=false`
5. Asserts **room snapshot** with 2 members via WebSocket reconnect
6. Host: `play --url <YouTube URL>`
7. Asserts **playback playing** in snapshot (waits for yt-dlp, default 90s)
8. Guest: `chat` smoke message

Environment overrides:

| Variable | Default | Meaning |
|----------|---------|---------|
| `E2E_VIDEO_URL` | Me at the Zoo (`jNQXAC9IVRw`) | YouTube URL for play step |
| `E2E_PLAY_TIMEOUT` | `90s` | Max wait for `playing` after `play` |
| `BIN_DIR` | `./bin` | Output directory for binaries |

The snapshot waiter lives at `scripts/e2e/wait/main.go` and is invoked by the smoke script.

Play uses `scripts/e2e/play/main.go` in the smoke script (host keeps WS open until `playing`). The CLI also supports `music-room play --url …` (waits for `playing`, then listens until Ctrl+C) or `--detach` for fire-and-forget after playback starts.

Local audio requires an active session with sync enabled: `join` (REPL/TUI) or `play` in a room. Set `MUSIC_ROOM_NO_PLAYBACK=1` to disable mpv (tests/CI).

## TUI v2 smoke (automated)

Headless sci-fi HUD check — no mpv or yt-dlp required:

```bash
chmod +x scripts/tui-smoke.sh
MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh
```

What it does:

1. Builds `bin/music-room` and `bin/music-roomd`
2. Starts `music-roomd` on a random local port
3. Host: `login` → `create <slug>`
4. Guest: `login` → headless join + Bubble Tea HUD render (`scripts/e2e/tui/main.go`)
5. Asserts HUD panels: `ROOM`, `NOW PLAYING`, `QUEUE`, `COMMS`
6. Asserts **room snapshot** with 2 members via WebSocket

Environment overrides:

| Variable | Default | Meaning |
|----------|---------|---------|
| `MUSIC_ROOM_NO_PLAYBACK` | `1` in script | Disable local mpv (recommended for CI) |
| `TUI_SMOKE_TIMEOUT` | `20s` | Max wait for connect/join/render |
| `BIN_DIR` | `./bin` | Output directory for binaries |

The headless renderer lives at `scripts/e2e/tui/main.go`. It mirrors `music-room join --tui` (default since v2): connect, join room, start TUI, quit with `q`, verify HUD composition.

## Manual checklist

Run these after the automated smoke passes (or on a staging server). Two terminals plus optional third for TUI.

### Setup

Terminal A — server:

```bash
make build
MUSIC_ROOM_LISTEN=:8080 ./bin/music-roomd
```

Terminal B — host:

```bash
./bin/music-room login --name host --server http://localhost:8080
./bin/music-room create my-room
./bin/music-room play --url 'https://www.youtube.com/watch?v=jNQXAC9IVRw'
```

Terminal C — guest:

```bash
./bin/music-room login --name guest --server http://localhost:8080
./bin/music-room join my-room --repl=false
```

### Playback drift (AC-021)

Requires **mpv** on each client (`join` or `play` starts the sync engine automatically).

- [ ] With 2+ members and a track playing, audio starts on both clients within ~1s of server `playing`
- [ ] After 2 minutes on stable LAN, perceived drift ≤ 500ms (target ≤ 200ms)
- [ ] Pause on one client (`music-room pause` or REPL `/pause`) pauses for all within 1s
- [ ] Resume syncs position within 500ms

Notes: measure drift by clapping against a shared metronome video or comparing `playback.tick` position in logs vs mpv OSD.

### Skip vote (AC-038)

With ≥2 members and a track playing:

- [ ] Guest runs `music-room vote skip` (or REPL `/vote skip`)
- [ ] Host runs `music-room vote skip`
- [ ] When votes exceed 50% of online members, track skips and queue advances (if any)
- [ ] System chat announces vote start and result

Example with 2 members (need both votes):

```bash
./bin/music-room vote skip   # guest
./bin/music-room vote skip   # host
```

### Reconnect (AC-048)

With guest joined and a track playing:

- [ ] Stop Wi‑Fi or kill server TCP briefly (< 5 minutes)
- [ ] Restart network; run any CLI command (`music-room chat hi`) or `join --repl` again
- [ ] Client reconnects with saved `session_id` in config
- [ ] Room snapshot restores: same slug, members, queue, playback position
- [ ] Playback resyncs within ~3s (AC-049), drift ≤ 500ms after stable

Config path: `~/.config/music-room/config.yaml` (`MUSIC_ROOM_CONFIG` to override).

### TUI session (AC-053–055)

```bash
./bin/music-room join my-room --tui
```

- [ ] Panels show room, now playing, members, queue, chat
- [ ] Chat and playback updates appear within 1s
- [ ] `q` exits TUI only (does not leave room); `l` opens leave confirm

### TUI v2 sci-fi HUD — manual (AC-006, AC-023)

Run after `MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh` passes (or full manual setup below).

**Setup** — Terminal A (server), B (host), C (guest):

```bash
# A
make build && MUSIC_ROOM_LISTEN=:8080 ./bin/music-roomd

# B
./bin/music-room login --name host --server http://localhost:8080
./bin/music-room create sci-fi-room
./bin/music-room play --url 'https://www.youtube.com/watch?v=jNQXAC9IVRw' --detach

# C
./bin/music-room login --name guest --server http://localhost:8080
./bin/music-room join sci-fi-room          # opens sci-fi TUI by default
# or: ./bin/music-room tui sci-fi-room
```

#### AC-006 — aesthetic (cyberpunk / sci-fi)

With ≥3 internal testers in the v2 HUD (80×24 or larger):

- [ ] Each tester describes the UI using at least one of: *cyberpunk*, *sci-fi*, *futuristic*, *neon*
- [ ] Focused panel shows magenta border; unfocused panels use cyan
- [ ] HUD includes: header (`ROOM` / `CREW`), `NOW PLAYING`, `QUEUE`, `COMMS`, `SIGNALS` (vote/reactions), status bar

#### AC-023 — host workflow entirely in TUI

As host in the sci-fi TUI (guest playing a track):

- [ ] `a` → add/play URL or search query → track plays or queues
- [ ] `Tab` / `Shift+Tab` → cycle focus (queue → chat → crew)
- [ ] `↑` / `↓` in queue panel → scroll / select item
- [ ] `s` → skip; `v` / `V` → vote skip / priority
- [ ] `d` or `ctrl+↑` / `ctrl+↓` → host queue admin (remove / reorder)
- [ ] `?` → help overlay lists shortcuts
- [ ] Subjective: flow feels ≤ CLI step count or faster (≥2/3 host testers agree)

Notes: compare against REPL fallback (`join --repl`) on the same room if needed.

### Host transfer (AC-013)

- [ ] Guest joins host room
- [ ] Host runs `music-room leave`
- [ ] Guest receives host change; guest can manage queue as new host

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Smoke hangs on play | `yt-dlp --version`; try `E2E_VIDEO_URL` with a short video |
| TUI smoke fails HUD assert | Terminal ≥80×24; re-run with `TUI_SMOKE_TIMEOUT=30s` |
| `ROOM_FULL` / rate limit | Fresh slug; wait 1 minute between smoke runs |
| Login fails | Server health: `curl -s http://127.0.0.1:PORT/healthz` |
| No audio | mpv installed; sync/TUI playback path (client-side) |

## Related tests

- Hub integration: `go test ./internal/server/... -run Integration`
- Full unit suite: `go test ./...`
