# Terminal Music Room

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Synchronized YouTube listening for terminal-first teams. Join a room, share a queue, chat, vote to skip, and listen in sync — without leaving the shell.

**v1 scope:** Ubuntu clients (`music-room` CLI/TUI + mpv), managed cloud server (`music-roomd`), 2–20 members per room.

## Features

- Shared rooms with host transfer and reconnect (5-minute window)
- Server-authoritative playback sync over WebSocket
- YouTube play by URL or search query (server-side `yt-dlp`)
- Queue, chat, skip/priority votes, emoji reactions
- Interactive REPL or Bubble Tea TUI (`join --tui`)

## Requirements

### Client (Ubuntu 22.04 / 24.04)

| Package | Purpose |
|---------|---------|
| `mpv` | Local audio playback |
| `yt-dlp` | YouTube stream extraction on client (mpv `--ytdl`) |
| `ffmpeg` | Recommended — audio demux/decoding |

```bash
sudo apt update
sudo apt install -y mpv yt-dlp ffmpeg
```

### Server (operator)

| Package | Purpose |
|---------|---------|
| `yt-dlp` | Resolve/search YouTube metadata |
| `ffmpeg` | Recommended |

See [docs/DEPLOY.md](docs/DEPLOY.md) for Docker, Fly.io, and Caddy deployment.

## Install

### From `.deb` (release)

Download `music-room_*.deb` from [GitHub Releases](https://github.com/terminal-music-room/music-room/releases), then:

```bash
sudo apt install -y mpv yt-dlp ffmpeg
sudo dpkg -i music-room_0.1.0-1_amd64.deb
```

Server package (optional, self-host):

```bash
sudo dpkg -i music-roomd_0.1.0-1_amd64.deb
```

### From source

Requires Go 1.22+.

```bash
git clone https://github.com/terminal-music-room/music-room.git
cd music-room
make build
# binaries: bin/music-room, bin/music-roomd
```

## Quickstart

### 1. Start the server (dev)

```bash
make build
MUSIC_ROOM_LISTEN=:8080 ./bin/music-roomd
```

Health check: `curl -s http://localhost:8080/healthz`

### 2. Log in and create a room

```bash
./bin/music-room login --name alice --server http://localhost:8080
./bin/music-room create backend-team
./bin/music-room play --url 'https://www.youtube.com/watch?v=jNQXAC9IVRw'
```

Config is saved to `~/.config/music-room/config.yaml`. Override with `MUSIC_ROOM_CONFIG` or `--config`.

### 3. Join with the TUI (second terminal)

```bash
./bin/music-room login --name bob --server http://localhost:8080
./bin/music-room join backend-team --tui
```

Press `q` to quit the TUI and leave the room. Use `--repl=false` for a one-shot join without UI.

### 4. CLI without TUI

```bash
./bin/music-room join backend-team --repl=false
./bin/music-room chat hello from bob
./bin/music-room pause
./bin/music-room vote skip
./bin/music-room leave
```

## CLI reference

| Command | Description |
|---------|-------------|
| `login --name <nick> [--server URL]` | Authenticate and save session |
| `create <slug>` | Create and join a room (you become host) |
| `join <slug> [--tui] [--repl]` | Join a room (TUI or REPL) |
| `leave` | Leave the current room |
| `play --url URL` / `--query TEXT` | Play YouTube URL or search |
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
| `MUSIC_ROOM_LISTEN` | Server listen address (`music-roomd`) |
| `MUSIC_ROOM_DATA_DIR` | Server chat log directory |

## Architecture

Design docs and ADRs:

- [Architecture](docs/vibe/001-terminal-music-room/architecture.md) — components, WebSocket protocol, deployment
- [Specification](docs/vibe/001-terminal-music-room/spec.md) — requirements and acceptance criteria
- [E2E testing](docs/E2E.md) — smoke script and manual checklists
- [Deployment](docs/DEPLOY.md) — Docker, Fly.io, Caddy

High level: `music-roomd` holds authoritative room/playback state; each client runs mpv locally and syncs to the server clock.

## Development

```bash
make test          # go test ./...
make vet           # go vet ./...
make build
./scripts/e2e-smoke.sh
go test ./internal/server/... -run Integration
```

## Legal disclaimer

**YouTube and yt-dlp.** This project plays audio from YouTube using [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [mpv](https://mpv.io/). YouTube’s Terms of Service may restrict downloading or playing content outside official YouTube clients. **You are responsible** for ensuring your use complies with applicable terms, copyright law, and your organization’s policies.

Terminal Music Room is an independent open-source tool — **not affiliated with, endorsed by, or sponsored by Google or YouTube**. Use at your own risk. The authors provide no warranty that any particular video or stream will remain available or playable.

Operators of a managed `music-roomd` instance should publish their own acceptable-use policy and respond to abuse reports.

## License

MIT — see [LICENSE](LICENSE).
