# Terminal Music Room

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Synchronized YouTube listening for terminal-first teams. Join a room, share a queue, chat, vote to skip, and listen in sync — without leaving the shell.

**V0.2.1 scope:** Ubuntu and Debian-based Linux + **macOS 13+** clients (`music-room` CLI/TUI + mpv), managed cloud server (`music-roomd`), 2–20 members per room. See [docs/PLATFORMS.md](docs/PLATFORMS.md).

## Features

- Shared rooms with host transfer and reconnect (5-minute window)
- Server-authoritative playback sync over WebSocket
- YouTube play by URL or search query (server-side `yt-dlp`)
- Queue, chat, skip/priority votes, emoji reactions
- Interactive sci-fi Bubble Tea HUD (default on `join`) or REPL fallback (`join --repl`)

## Requirements

### Client — Linux (Ubuntu 22.04 / 24.04, Debian 12+)

| Package | Purpose |
|---------|---------|
| `mpv` | Local audio playback |
| `yt-dlp` | YouTube stream extraction on client (mpv `--ytdl`) |
| `ffmpeg` | Recommended — audio demux/decoding |

```bash
sudo apt update
sudo apt install -y mpv yt-dlp ffmpeg
```

Debian-based derivatives (Linux Mint, Pop!_OS, etc.) use the same packages via `apt`. **Not supported:** Fedora, Arch, and other non-Debian families in V0.2.1.

### Client — macOS (13 Ventura+)

| Package | Purpose |
|---------|---------|
| `mpv` | Local audio playback |
| `yt-dlp` | YouTube stream extraction (mpv `--ytdl`) |
| `ffmpeg` | Recommended — audio demux/decoding |

```bash
brew install mpv yt-dlp ffmpeg
```

Use **Terminal.app** (or iTerm2) at ≥ 80×24 for the sci-fi TUI. Release binaries are unsigned — see [Gatekeeper](#macos-gatekeeper-unsigned-binary) below.

### Server (operator)

| Package | Purpose |
|---------|---------|
| `yt-dlp` | Resolve/search YouTube metadata |
| `ffmpeg` | Recommended |

See [docs/DEPLOY.md](docs/DEPLOY.md) for Docker, Fly.io, and Caddy deployment.

## Install

Release assets: [GitHub Releases](https://github.com/tuanhm-kaopiz/terminal-music-room/releases) — pick **one** of the options below.

### From release tarball (linux/amd64)

> **Important:** On the [Releases](https://github.com/tuanhm-kaopiz/terminal-music-room/releases) page, GitHub also lists **“Source code (tar.gz)”** — that archive is **full source**, not ready-to-run binaries. After extract you get a folder like `terminal-music-room-0.1.0/` with `.go` files, **not** `music-room`.
>
> Download the **binary** asset instead:
> - `terminal-music-room_0.1.0_linux_amd64.tar.gz` — client + server + checksums  
> - or download `music-room` and `music-roomd` directly from the same release

The **binary** archive contains:

| File | Role |
|------|------|
| `music-room` | CLI/TUI client |
| `music-roomd` | Sync server (self-host) |
| `SHA256SUMS` | Checksums for the binaries |

**1. Install system dependencies (Ubuntu):**

```bash
sudo apt update
sudo apt install -y mpv yt-dlp ffmpeg curl
```

**2. Find the downloaded file, then extract:**

The archive name on GitHub may be `terminal-music-room_0.1.0_linux_amd64.tar.gz` or similar — **use the exact filename you downloaded** (hyphens vs underscores differ). `tar` fails with *Cannot open* when you are in the wrong directory or the name does not match.

```bash
# Find where the browser saved it (common: ~/Downloads)
ls ~/Downloads/*terminal-music-room*.tar.gz

# Set the path to YOUR **binary** tarball (not "Source code"):
TARBALL=~/Downloads/terminal-music-room_0.1.0_linux_amd64.tar.gz

mkdir -p ~/apps/terminal-music-room
tar -xzf "$TARBALL" -C ~/apps/terminal-music-room
cd ~/apps/terminal-music-room
ls
# expect: music-room  music-roomd  SHA256SUMS   (no nested source folder)
sha256sum -c SHA256SUMS
```

**Wrong archive?** If `ls` shows only `terminal-music-room-0.1.0/` (source tree), go back to Releases and download `terminal-music-room_*_linux_amd64.tar.gz` or the standalone `music-room` binary — see [From source](#from-source) only if you want to compile yourself.

**3. Put binaries on your `PATH` (pick one):**

```bash
# Option A — user-local install (recommended)
mkdir -p ~/.local/bin
cp music-room music-roomd ~/.local/bin/
# ensure ~/.local/bin is in PATH (add to ~/.bashrc or ~/.zshrc if needed):
export PATH="$HOME/.local/bin:$PATH"

# Option B — system-wide
sudo cp music-room music-roomd /usr/local/bin/
```

**4. Check:**

```bash
music-room --help
music-roomd --help   # optional, only if you self-host the server
```

Then follow [Quickstart](#quickstart) below — use `music-room` / `music-roomd` directly (no `./bin/` prefix).

### From `.deb` (release)

Download `music-room_*.deb` from [GitHub Releases](https://github.com/tuanhm-kaopiz/terminal-music-room/releases), then:

```bash
sudo apt install -y mpv yt-dlp ffmpeg
sudo dpkg -i music-room_0.1.0-1_amd64.deb
```

Server package (optional, self-host):

```bash
sudo dpkg -i music-roomd_0.1.0-1_amd64.deb
```

After `.deb` install, binaries are on `PATH` as `music-room` and `music-roomd`.

### macOS — GitHub Release (arm64 / Intel)

Download the tarball matching your Mac from [GitHub Releases](https://github.com/tuanhm-kaopiz/terminal-music-room/releases):

| Mac | Asset |
|-----|-------|
| Apple Silicon (M1/M2/M3…) | `terminal-music-room_0.2.1_darwin_arm64.tar.gz` |
| Intel | `terminal-music-room_0.2.1_darwin_amd64.tar.gz` |

**1. Install dependencies:**

```bash
brew install mpv yt-dlp ffmpeg
```

**2. Extract and install:**

```bash
TARBALL=~/Downloads/terminal-music-room_0.2.1_darwin_arm64.tar.gz   # adjust path/arch
mkdir -p ~/apps/terminal-music-room
tar -xzf "$TARBALL" -C ~/apps/terminal-music-room
cd ~/apps/terminal-music-room
sha256sum -c SHA256SUMS
mkdir -p ~/.local/bin
cp music-room ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
music-room --version
```

**3. Gatekeeper:** if macOS blocks the binary, see [macOS Gatekeeper](#macos-gatekeeper-unsigned-binary).

### macOS — Homebrew

Install dependencies first (`brew install mpv yt-dlp ffmpeg`), then either:

```bash
# After v0.2.1 release (formula url/sha256 pinned in repo):
brew install --formula ./packaging/homebrew/Formula/music-room.rb
```

See [packaging/homebrew/README.md](packaging/homebrew/README.md) for tap workflow and `bump-formula.sh`.

### macOS Gatekeeper (unsigned binary)

V0.2.1 binaries are **not** notarized. On first run you may see *"cannot be opened because the developer cannot be verified"*.

**Option A — System Settings**

1. Try running `music-room` once.
2. Open **System Settings → Privacy & Security**.
3. Click **Open Anyway** for `music-room`.

**Option B — remove quarantine attribute**

```bash
xattr -dr com.apple.quarantine /path/to/music-room
```

### From source

Requires Go 1.24+ (1.24 adds Mach-O LC_UUID required on macOS 15+/Tahoe).

```bash
git clone https://github.com/tuanhm-kaopiz/terminal-music-room.git
cd music-room
make build
# binaries: bin/music-room, bin/music-roomd
```

### Upgrade from v0.1.x

Your saved session in `~/.config/music-room/config.yaml` is kept — **no need to `login` again**. Only replace the binaries.

**One-liner** (from GitHub; needs `curl`):

```bash
# Tarball → ~/.local/bin or same dir as your current music-room
curl -fsSL https://raw.githubusercontent.com/tuanhm-kaopiz/terminal-music-room/main/scripts/upgrade.sh | bash -s -- tarball

# .deb (Ubuntu; prompts for sudo)
curl -fsSL https://raw.githubusercontent.com/tuanhm-kaopiz/terminal-music-room/main/scripts/upgrade.sh | bash -s -- deb
```

From a cloned repo:

```bash
chmod +x scripts/upgrade.sh
./scripts/upgrade.sh tarball          # client only
./scripts/upgrade.sh deb                # client .deb
INSTALL_SERVER=1 ./scripts/upgrade.sh tarball   # client + music-roomd
```

Pin a version: `./scripts/upgrade.sh tarball 0.2.0` or `VERSION=0.2.0 ./scripts/upgrade.sh deb`.

**v0.2.0 changes:** `join` opens the sci-fi TUI by default (use `join --repl` for REPL). `q` exits the TUI only; use `l` or `music-room leave` to leave the room. See [Breaking changes (TUI v2)](#breaking-changes-tui-v2).

## Quickstart

> **Paths:** If you installed from a **release** (tarball or `.deb`), use `music-room` and `music-roomd`. If you **built from source**, use `./bin/music-room` and `./bin/music-roomd` instead.

### 1. Start the server (dev / self-host)

Terminal A:

```bash
MUSIC_ROOM_LISTEN=:8080 music-roomd
```

Health check: `curl -s http://localhost:8080/healthz`

### 2. Log in and create a room

Terminal B:

```bash
music-room login --name alice --server http://localhost:8080
music-room create backend-team
music-room play --url 'https://www.youtube.com/watch?v=jNQXAC9IVRw'
# plays until Ctrl+C — audio via mpv; keep this terminal open or run join in another terminal
```

Config is saved to `~/.config/music-room/config.yaml`. Override with `MUSIC_ROOM_CONFIG` or `--config`.

### 3. Join with the sci-fi TUI (second user)

Terminal C:

```bash
music-room login --name bob --server http://localhost:8080
music-room join backend-team          # opens sci-fi HUD by default (v2)
# or: music-room tui backend-team     # join + HUD in one command
```

**Quit vs leave:** `q` exits the TUI but **stays in the room** (playback/sync continues in the background if another client session is active). `l` opens a confirm dialog and **leaves the room** via WebSocket.

Power-user fallbacks:

```bash
music-room join backend-team --repl              # interactive REPL instead of TUI
music-room join backend-team --tui=false --repl=false   # one-shot join, no UI
```

### 4. CLI without TUI

```bash
music-room join backend-team --tui=false --repl=false
music-room chat hello from bob
music-room pause
music-room vote skip
music-room leave
```

### Typical flow (2 people, one server)

```
Terminal 1:  music-roomd                          # server
Terminal 2:  music-room login → create → play     # host (Alice)
Terminal 3:  music-room login → join              # guest (Bob) — sci-fi TUI
```

Both Alice and Bob need **mpv** + **yt-dlp** installed locally to hear audio in sync.

## Sci-fi TUI (v2)

Cyberpunk-style HUD for host and member — unified layout at 80×24 (degraded mode below that size).

```
┌─ ◈ ROOM: backend-team  CREW: 2  ● connected ─────────────────────────────┐
│ NOW PLAYING          │ CREW              │ SIGNALS (vote / reactions)    │
│ Neon Nights          │ host#1 (host)     │ skip vote 1/2                 │
│ ████████░░░░ 3:00    │ guest#2           │ 🔥×1                          │
├──────────────────────┴───────────────────┴──────────────────────────────────┤
│ QUEUE                                                                       │
│ › Track Two — guest#2                                                       │
│   Track Three — host#1                                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│ COMMS                                                                       │
│ guest#2: hello                                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ > Type message… (? help · q exit TUI)                                       │
└─────────────────────────────────────────────────────────────────────────────┘
  Space pause · s skip · a add · Tab focus · ? help · q exit · l leave
```

Focused panel border uses **magenta**; unfocused panels use **cyan**. Press `?` in the TUI for the full overlay.

### TUI keymap

| Key | Action |
|-----|--------|
| `Space` | Pause / resume |
| `s` | Skip track |
| `S` | Seek (milliseconds modal) |
| `a` | Add URL or search (`Tab` toggles play vs queue) |
| `Enter` | Send chat |
| `v` | Vote skip |
| `V` | Vote priority (selected queue item) |
| `1`–`4` | Quick react 🔥 ❤️ 😂 👍 |
| `Tab` / `Shift+Tab` | Cycle focus: queue → chat → crew |
| `↑` / `↓` | Scroll focused panel (queue selection / chat / members) |
| `d` | Remove selected queue item **(host)** |
| `Ctrl+↑` / `Ctrl+↓` | Reorder queue item **(host)** |
| `q` | Exit TUI (stay in room) |
| `l` | Leave room (confirm) |
| `?` / `Esc` | Help overlay |

## Breaking changes (TUI v2)

| Before (v1) | After (v2) |
|-------------|------------|
| `join` opened REPL by default | `join` opens **sci-fi TUI** by default |
| `--tui` required for Bubble Tea | `--tui` still exists (default `true`); use `--repl` for REPL |
| Simple wireframe TUI panels | Unified cyberpunk HUD for host and member |
| `q` behavior varied | `q` **only** exits TUI; use `l` or `music-room leave` to leave the room |

Scripting and automation: use `join --tui=false --repl=false` or one-shot CLI commands (`chat`, `play`, …). `create` remains one-shot (no auto-TUI).

## CLI reference

| Command | Description |
|---------|-------------|
| `login --name <nick> [--server URL]` | Authenticate and save session |
| `create <slug>` | Create and join a room (you become host) |
| `join <slug> [--tui] [--repl]` | Join a room (TUI by default; `--repl` for REPL) |
| `tui [slug]` | Open sci-fi HUD (join `slug` first if given) |
| `leave` | Leave the current room |
| `play --url URL` / `--query TEXT` | Play YouTube URL or search (listens until Ctrl+C; use `--detach` to return after start) |
| `pause` / `resume` / `skip` | Playback control |
| `seek <ms>` | Seek to position in milliseconds |
| `queue add --url URL` / `--query TEXT` | Add to queue |
| `queue remove <id>` / `reorder <id> --after <id>` | Host queue management |
| `chat <message...>` | Send chat |
| `vote skip` / `vote priority <id>` | Room votes |
| `react <emoji>` | Reaction on current track |

Environment:

| Variable | Description |
|----------|-------------|
| `MUSIC_ROOM_SERVER_URL` | Default server URL for login |
| `MUSIC_ROOM_CONFIG` | Config file path |
| `MUSIC_ROOM_NO_PLAYBACK` | Set to `1` to disable local mpv (tests) |
| `MUSIC_ROOM_LISTEN` | Server listen address (`music-roomd`) |
| `MUSIC_ROOM_DATA_DIR` | Server chat log directory |

## Architecture

Design docs and ADRs:

- [Architecture](docs/vibe/001-terminal-music-room/architecture.md) — components, WebSocket protocol, deployment
- [Specification](docs/vibe/001-terminal-music-room/spec.md) — requirements and acceptance criteria
- [E2E testing](docs/E2E.md) — smoke scripts (`e2e-smoke.sh`, `tui-smoke.sh`) and manual checklists
- [Supported platforms](docs/PLATFORMS.md) — Linux, macOS, release artifacts
- [Deployment](docs/DEPLOY.md) — Docker, Fly.io, Caddy

High level: `music-roomd` holds authoritative room/playback state; each client runs mpv locally and syncs to the server clock.

## Development

```bash
make test          # go test ./...
make vet           # go vet ./...
make build
./scripts/e2e-smoke.sh
MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh
go test ./internal/server/... -run Integration
```

## Legal disclaimer

**YouTube and yt-dlp.** This project plays audio from YouTube using [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [mpv](https://mpv.io/). YouTube’s Terms of Service may restrict downloading or playing content outside official YouTube clients. **You are responsible** for ensuring your use complies with applicable terms, copyright law, and your organization’s policies.

Terminal Music Room is an independent open-source tool — **not affiliated with, endorsed by, or sponsored by Google or YouTube**. Use at your own risk. The authors provide no warranty that any particular video or stream will remain available or playable.

Operators of a managed `music-roomd` instance should publish their own acceptable-use policy and respond to abuse reports.

## Releasing (maintainers)

GitHub Releases are **not** created when you push `main` only. The workflow in `.github/workflows/release.yml` runs when you push a **version tag** `v*`.

```bash
# After tests pass locally:
go test ./...
git tag -a v0.2.1 -m "v0.2.1 — macOS cross-platform"
git push origin v0.2.1
```

GitHub Actions will build Linux + macOS client tarballs, `.deb` packages, and publish assets to:

https://github.com/tuanhm-kaopiz/terminal-music-room/releases

(Use `/releases`, not `/release`.)

## License

MIT — see [LICENSE](LICENSE).
