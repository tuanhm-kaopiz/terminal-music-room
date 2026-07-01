# Tasks: Terminal Music Room

**Slug:** `terminal-music-room`
**Status:** pending
**Gate G4:** ✅ pass

## Task list

> Execute in order. Mark `[x]` when done. Add `<!-- BLOCKED: reason -->` if stuck.

### Phase 1 — Foundation

- [x] **T-001:** Go monorepo scaffold + toolchain config
  - Deliverable: `go.mod`, `cmd/music-room/main.go`, `cmd/music-roomd/main.go`, `Makefile` (`test`, `lint`, `build`), `LICENSE` (MIT), `.gitignore`; `vibe.config.yaml` commands → `go test ./...`, `go vet ./...`
  - Maps to: NFR-009, ADR-001, ADR-006
  - Validation: `go build ./... && go vet ./...`
  - Files: `go.mod`, `cmd/`, `Makefile`, `LICENSE`, `vibe.config.yaml`

- [x] **T-002:** WebSocket protocol package
  - Deliverable: `internal/protocol` — envelope `{type,id,payload}`, all message types (architecture §API), error codes, JSON marshal/unmarshal helpers
  - Maps to: FR-007, ADR-003; enables AC-007–010, AC-051
  - Validation: `go test ./internal/protocol/...`
  - Files: `internal/protocol/*.go`, `internal/protocol/*_test.go`

- [x] **T-003:** Domain models + playback clock (unit-tested)
  - Deliverable: `Room`, `Member`, `PlaybackState`, `QueueItem`, `ChatMessage`, `Vote`, `Track`; `EffectivePositionMs()`; slug/nickname validators
  - Maps to: AC-006, AC-016, AC-021, AC-026, FR-002, FR-007
  - Validation: `go test ./internal/server/playback/... ./internal/server/room/...`
  - Files: `internal/server/room/`, `internal/server/playback/`

- [x] **T-004:** Server bootstrap (`music-roomd`)
  - Deliverable: HTTP server, `GET /healthz`, `GET /v1/ws` WebSocket upgrade (`coder/websocket`), graceful shutdown, `slog` logging, env config (`MUSIC_ROOM_LISTEN`)
  - Maps to: NFR-001 (transport ready), deployment §architecture
  - Validation: `go test ./internal/server/hub/... -run TestHealth` && `go build -o bin/music-roomd ./cmd/music-roomd`
  - Files: `cmd/music-roomd/main.go`, `internal/server/hub/server.go`

### Phase 2 — Server core

- [x] **T-005:** Session hub + rate limiting
  - Deliverable: `session.hello` → `session.ack` with UUID; headers `X-Session-Id`/`X-Nickname`; IP token-bucket (connect, create room, chat); max 3 sessions/IP
  - Maps to: AC-001–003, AC-002, NFR-007, FR-001
  - Validation: `go test ./internal/server/hub/... -run Session`
  - Files: `internal/server/hub/session.go`, `internal/server/hub/ratelimit.go`

- [x] **T-006:** Room lifecycle
  - Deliverable: `room.create`, `room.join`, `room.leave`; slug registry unique; cap 20; `room.snapshot` on join; host transfer (AC-013); destroy + free slug (AC-014); `display_name` disambiguation (AC-016); member events <500ms broadcast
  - Maps to: AC-004–015, FR-002–005, NFR-001, NFR-005
  - Validation: `go test ./internal/server/room/...`
  - Files: `internal/server/room/manager.go`, `internal/server/hub/handlers_room.go`

- [x] **T-007:** Authoritative playback + broadcast
  - Deliverable: Handlers `playback.play|pause|resume|skip|seek`; state machine (playing/paused/buffering/ended); `playback.tick` every 1s + immediate on command; democratic member commands
  - Maps to: AC-017–026, AC-022–025, FR-006, FR-007, NFR-002, NFR-003
  - Validation: `go test ./internal/server/playback/...`
  - Files: `internal/server/playback/clock.go`, `internal/server/hub/handlers_playback.go`

- [x] **T-008:** YouTube resolver (server-side yt-dlp)
  - Deliverable: `internal/server/youtube` — URL validate, keyword search (`ytsearch5:`), metadata extract, worker pool (non-blocking WS), 10s timeout, in-memory search cache 5m; errors `INVALID_SOURCE`, `SOURCE_UNAVAILABLE`
  - Maps to: AC-017–020, ADR-007, FR-006
  - Validation: `go test ./internal/server/youtube/...` (mock exec); manual: `go test -tags=integration ./internal/server/youtube/...` if yt-dlp present
  - Files: `internal/server/youtube/resolver.go`, `internal/server/youtube/resolver_test.go`

- [x] **T-009:** Queue service
  - Deliverable: `queue.add` (any member), `queue.remove`/`queue.reorder` (host only → `FORBIDDEN`), `queue.updated` broadcast, auto-advance on ended (AC-029)
  - Maps to: AC-027–032, FR-008, FR-009
  - Validation: `go test ./internal/server/room/... -run Queue`
  - Files: `internal/server/room/queue.go`, `internal/server/hub/handlers_queue.go`

- [x] **T-010:** Chat service
  - Deliverable: `chat.send` with validation (non-empty AC-036), emoji pass-through, ring buffer ~100, system messages on join/leave/song/vote (AC-035), `chat.message` broadcast
  - Maps to: AC-033–036, FR-010
  - Validation: `go test ./internal/server/chat/...`
  - Files: `internal/server/chat/buffer.go`, `internal/server/hub/handlers_chat.go`

- [x] **T-011:** Vote engine (skip + priority)
  - Deliverable: `vote.skip`, `vote.priority`; >50% threshold snapshot at vote start; dedupe votes (AC-039); progress in `vote.updated`; cancel on missing queue item (AC-044); timeout handling (AC-040)
  - Maps to: AC-037–044, FR-011, FR-012
  - Validation: `go test ./internal/server/vote/...`
  - Files: `internal/server/vote/engine.go`, `internal/server/hub/handlers_vote.go`

- [x] **T-012:** Reactions + reconnect window
  - Deliverable: `reaction.send` aggregated per track; reset on track change (AC-046); reject when no track (AC-047); session reconnect within 5m restores room membership + snapshot (AC-048–050)
  - Maps to: AC-045–050, FR-013, FR-014, NFR-006
  - Validation: `go test ./internal/server/hub/... -run Reconnect`
  - Files: `internal/server/room/reactions.go`, `internal/server/hub/reconnect.go`

### Phase 3 — Client core

- [x] **T-013:** Client config + login command
  - Deliverable: `~/.config/music-room/config.yaml` (nickname, server_url, session_id); Cobra `music-room login --name`; validation 1–32 chars (AC-001, AC-002)
  - Maps to: AC-001–003, FR-001, FR-015
  - Validation: `go test ./internal/client/config/... && go build -o bin/music-room ./cmd/music-room`
  - Files: `internal/client/config/config.go`, `internal/client/cli/login.go`, `cmd/music-room/main.go`

- [x] **T-014:** WebSocket client + shared state store
  - Deliverable: Connect to `wss://…/v1/ws`, dispatch server events to state store, exponential backoff reconnect 1s→30s max 5m, correlation IDs
  - Maps to: AC-048–050, FR-014, NFR-006
  - Validation: `go test ./internal/client/ws/...`
  - Files: `internal/client/ws/client.go`, `internal/client/state/store.go`

- [x] **T-015:** mpv player driver
  - Deliverable: Spawn mpv with `--input-ipc-server`, `--ytdl`, `--ytdl-format=bestaudio`; play/pause/seek/get position; kill on leave (AC-011); YouTube via video_id URL
  - Maps to: AC-011, AC-017, ADR-002, NFR-004, NFR-008
  - Validation: `go test ./internal/client/player/...` (mock IPC); manual smoke with mpv installed
  - Files: `internal/client/player/mpv.go`, `internal/client/player/mpv_test.go`

- [x] **T-016:** Playback sync engine
  - Deliverable: Subscribe `playback.state`/`playback.tick`; drift correction seek if |drift|>150ms; pause/resume sync; load new track on change
  - Maps to: AC-021–026, AC-049, FR-007, NFR-003
  - Validation: `go test ./internal/client/sync/...`
  - Files: `internal/client/sync/engine.go`, `internal/client/sync/engine_test.go`

- [x] **T-017:** Cobra CLI commands + interactive REPL
  - Deliverable: Subcommands `create`, `join`, `leave`, `play`, `pause`, `resume`, `skip`, `seek`, `queue`, `chat`, `vote`, `react`; REPL mode `/play`, `/queue`, `/chat` when joined; invalid command hints (AC-052)
  - Maps to: AC-051–052, FR-015, REQ-002–013 (CLI path)
  - Validation: `go test ./internal/client/cli/...`
  - Files: `internal/client/cli/*.go`

- [x] **T-018:** Bubble Tea TUI
  - Deliverable: `music-room join <slug> --tui` — panels: room, now playing, members, queue, chat; refresh on WS events <1s (AC-054); shared WS with CLI mode (AC-055); follow `tui-design`/`bubbletea` skills
  - Maps to: AC-053–055, FR-016, NFR-010
  - Validation: `go build ./cmd/music-room` + manual TUI smoke in 80×24 terminal
  - Files: `internal/client/tui/model.go`, `internal/client/tui/view.go`, `internal/client/tui/update.go`

### Phase 4 — Integration, packaging & release

- [x] **T-019:** Server WebSocket integration tests
  - Deliverable: `httptest` + WS fake clients: create→join→play→pause→chat flow; room full; slug taken; host leave transfer
  - Maps to: AC-004–010, AC-013, AC-032, AC-036
  - Validation: `go test ./internal/server/... -run Integration`
  - Files: `internal/server/hub/integration_test.go`

- [x] **T-020:** End-to-end smoke script + docs
  - Deliverable: `scripts/e2e-smoke.sh` — start `music-roomd`, 2 CLI clients join same room, play URL, assert snapshot; `docs/E2E.md` manual checklist (drift, vote, reconnect)
  - Maps to: AC-021, AC-038, AC-048 (manual verification)
  - Validation: `./scripts/e2e-smoke.sh` (requires mpv, yt-dlp, Ubuntu)
  - Files: `scripts/e2e-smoke.sh`, `docs/E2E.md`

- [x] **T-021:** Server deployment artifacts
  - Deliverable: `Dockerfile.music-roomd`, `docker-compose.yml` (dev), Fly.io `fly.toml` or deploy README section; Caddy TLS example
  - Maps to: clarify SaaS model, architecture deployment
  - Validation: `docker build -f Dockerfile.music-roomd -t music-roomd:local .`
  - Files: `Dockerfile.music-roomd`, `deploy/`, `docs/DEPLOY.md`

- [x] **T-022:** Ubuntu packaging + CI release
  - Deliverable: `packaging/debian/control` (Depends: mpv, yt-dlp); GitHub Actions: `go test`, build `linux/amd64` binaries, attach release assets
  - Maps to: NFR-008, NFR-009, AC-001 (install path)
  - Validation: `go test ./...` in CI; local `dpkg-deb` build optional
  - Files: `packaging/debian/`, `.github/workflows/release.yml`

- [x] **T-023:** Project README + legal disclaimer
  - Deliverable: `README.md` — install (apt/deb), deps, quickstart (`login`→`create`→`join --tui`), architecture link, YouTube/yt-dlp ToS disclaimer, MIT license badge
  - Maps to: NFR-009, ADR-002, ADR-006, risks §clarify
  - Validation: README commands copy-paste accurate; `go test ./...` green
  - Files: `README.md`

## Dependency graph (summary)

```
T-001 → T-002 → T-003 → T-004
T-004 → T-005 → T-006 → T-007
T-007 → T-008 → T-009
T-006 → T-010 → T-011 → T-012
T-002 → T-013 → T-014 → T-015 → T-016 → T-017
T-012 + T-017 → T-019 → T-020
T-019 → T-021 → T-022 → T-023
```

## AC coverage map (by task)

| Task | Primary AC |
|------|------------|
| T-005 | AC-001–003 |
| T-006 | AC-004–016 |
| T-007 | AC-017–026 |
| T-008 | AC-017–020 |
| T-009 | AC-027–032 |
| T-010 | AC-033–036 |
| T-011 | AC-037–044 |
| T-012 | AC-045–050 |
| T-013–017 | AC-051–055, AC-011 |
| T-019–020 | cross-cutting integration |

## Progress

| Phase | Total | Done | Blocked |
|-------|-------|------|---------|
| 1 — Foundation | 4 | 4 | 0 |
| 2 — Server core | 8 | 8 | 0 |
| 3 — Client core | 6 | 5 | 0 |
| 4 — Integration & release | 5 | 0 | 0 |
| **Total** | **23** | **17** | **0** |

## Gate G4 checklist

- [x] Tasks ordered by dependency
- [x] Each task has deliverable + validation
- [x] Each task maps to AC/requirement
- [x] No vague tasks
