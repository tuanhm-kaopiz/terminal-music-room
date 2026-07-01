# Deploying music-roomd

Terminal Music Room v1 uses a **managed cloud** model: the operator runs `music-roomd`; Ubuntu clients connect with `music-room` (CLI/TUI + mpv). This document covers container and Fly.io deployment.

## Requirements

| Component | Notes |
|-----------|--------|
| `music-roomd` | WebSocket hub + yt-dlp resolver |
| `yt-dlp` | Bundled in `Dockerfile.music-roomd` image |
| Persistent volume | Chat logs at `MUSIC_ROOM_DATA_DIR` (default `/data/chat`) |
| TLS (production) | Caddy or Fly.io automatic HTTPS |

Clients are **not** containerized in v1 — install `music-room`, `mpv`, and `yt-dlp` on Ubuntu hosts.

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `MUSIC_ROOM_LISTEN` | `:8080` | HTTP + WebSocket listen address |
| `MUSIC_ROOM_DATA_DIR` | `./data/chat` | On-disk chat log directory |
| `MUSIC_ROOM_YTDLP` | `yt-dlp` | Path to yt-dlp binary (override if needed) |

Health check: `GET /healthz` → `{"status":"ok"}`  
WebSocket: `GET /v1/ws` (upgrade)

## Docker (local / VPS)

Build and run:

```bash
docker build -f Dockerfile.music-roomd -t music-roomd:local .
docker run --rm -p 8080:8080 \
  -e MUSIC_ROOM_LISTEN=:8080 \
  -e MUSIC_ROOM_DATA_DIR=/data/chat \
  -v music-room-data:/data/chat \
  music-roomd:local
```

Dev compose (server only):

```bash
docker compose up --build
```

Point clients at `http://localhost:8080`:

```bash
./bin/music-room login --name you --server http://localhost:8080
```

### Caddy TLS (example)

For a VPS with a real domain, use the Caddy overlay:

1. Edit `deploy/caddy/Caddyfile` — set your domain.
2. Ensure DNS points to the host.
3. Start stack:

```bash
docker compose -f docker-compose.yml -f deploy/caddy/docker-compose.caddy.yml up -d --build
```

Clients use `https://music.example.com` as `--server` (WebSocket upgrades to `wss://` automatically).

On bare metal without Docker, install [Caddy](https://caddyserver.com/) and proxy to `127.0.0.1:8080` using the same `reverse_proxy` stanza from `deploy/caddy/Caddyfile`.

## Fly.io

1. Install [flyctl](https://fly.io/docs/hands-on/install-flyctl/).
2. From the **repository root**:

```bash
fly apps create your-music-room
fly volumes create music_room_data --region sin --size 1
```

3. Set `app = "your-music-room"` in `deploy/fly.toml` (or pass `-a`).
4. Deploy:

```bash
fly deploy --config deploy/fly.toml
```

5. Client login:

```bash
music-room login --name you --server https://your-music-room.fly.dev
```

Fly terminates TLS and forwards HTTP to `:8080`. The mounted volume keeps chat logs across deploys.

Suggested machine size (v1): **shared-cpu-1x, 256MB** — see `[[vm]]` in `deploy/fly.toml`.

## Hetzner / generic VPS

1. Build or pull the image on the host (or use `docker compose`).
2. Run `music-roomd` bound to `127.0.0.1:8080` behind Caddy.
3. Open firewall ports 80/443 only (not 8080 publicly).
4. Back up the chat data volume periodically.

## Observability

- Logs: JSON on stdout (`slog`) — ship with your platform (Fly logs, Docker logging driver, journald).
- Liveness: `/healthz`
- No metrics endpoint in v1

## Security notes

- Rate limits apply per client IP at connect, room create, and chat (see server hub).
- Do not expose unauthenticated admin interfaces — v1 has no admin API.
- Keep yt-dlp updated in the image for source extraction fixes.

## Related

- [E2E.md](./E2E.md) — smoke tests and manual checklists
- [architecture.md](./vibe/001-terminal-music-room/architecture.md) — deployment diagram
