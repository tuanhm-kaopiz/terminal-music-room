# Tasks: macOS Cross-Platform Support (V0.2.1)

**Slug:** `macos-cross-platform`
**Status:** complete
**Gate G4:** ✅ pass

## Task list

> Execute in order. Mark `[x]` when done. Add `<!-- BLOCKED: reason -->` if stuck.

### Phase 1 — Foundation

- [x] **T-001:** Version injection via ldflags
  - Deliverable: `cmd/music-room/main.go` dùng `var version = "dev"` (không `const`); `packaging/build-deb.sh`, `packaging/build-macos.sh` (stub OK nếu T-004 chưa có), `.github/workflows/release.yml` truyền `-ldflags "-s -w -X main.version=${VERSION}"` cho mọi `go build` client; `music-room --version` in đúng tag khi build release
  - Maps to: AC-034, AC-035, FR-008, ADR-002
  - Validation: `VERSION=0.2.1-test go build -ldflags "-s -w -X main.version=0.2.1-test" -o /tmp/music-room ./cmd/music-room && /tmp/music-room --version | grep 0.2.1-test`
  - Files: `cmd/music-room/main.go`, `packaging/build-deb.sh`, `.github/workflows/release.yml`

- [x] **T-002:** `internal/client/deps` preflight package
  - Deliverable: `deps.Check()` dùng `exec.LookPath` cho `mpv`, `yt-dlp`; `CheckResult{Missing, Hints}`; `hints_linux.go` → `sudo apt install -y mpv yt-dlp ffmpeg`; `hints_darwin.go` → `brew install mpv yt-dlp ffmpeg`; `FormatError(result)` trả message user-facing; unit tests mock `lookPath`
  - Maps to: AC-004, AC-029, FR-003, ADR-008
  - Validation: `go test ./internal/client/deps/...`
  - Files: `internal/client/deps/deps.go`, `hints_linux.go`, `hints_darwin.go`, `deps_test.go`, `doc.go`

- [x] **T-003:** Wire dependency preflight vào playback paths
  - Deliverable: Gọi `deps.Check()` trong `startLocalPlayback()` (trước `player.New`) và lệnh `play` khi playback enabled; skip khi `MUSIC_ROOM_NO_PLAYBACK=1`; lỗi in `FormatError` ra stderr, không crash im lặng
  - Maps to: AC-004, AC-007, AC-008, AC-029, ADR-008
  - Validation: `go test ./internal/client/cli/... && MUSIC_ROOM_NO_PLAYBACK=1 go test ./internal/client/cli/...`
  - Files: `internal/client/cli/playback_session.go`, `internal/client/cli/playback.go`, `internal/client/cli/room.go` (nếu join start playback)

- [x] **T-004:** Verify darwin cross-compile + client tests
  - Deliverable: `GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/music-room` và `GOARCH=amd64` pass; `go test ./...` green trên dev; ghi chú fix nhỏ nếu phát hiện darwin-only issue (không thêm `//go:build linux` trên client)
  - Maps to: AC-001, AC-002, AC-008, AC-013, AC-014, FR-003, FR-004, ADR-001, ADR-002
  - Validation: `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/music-room-darwin-arm64 ./cmd/music-room && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o /tmp/music-room-darwin-amd64 ./cmd/music-room && go test ./...`
  - Files: `internal/client/*` (chỉ nếu cần fix)

### Phase 2 — Release & packaging

- [x] **T-005:** `packaging/build-macos.sh`
  - Deliverable: Script nhận `[version]`; build `music-room` cho `darwin/arm64` + `darwin/amd64` với version ldflags; tạo `terminal-music-room_{VERSION}_darwin_arm64.tar.gz` và `_darwin_amd64.tar.gz` (chỉ `music-room` + `SHA256SUMS`); in danh sách artifact
  - Maps to: AC-001, AC-002, AC-033, FR-001, FR-008, ADR-002
  - Validation: `./packaging/build-macos.sh 0.2.1-test && ls dist/terminal-music-room_0.2.1-test_darwin_*.tar.gz && tar -tzf dist/terminal-music-room_0.2.1-test_darwin_arm64.tar.gz | grep -E '^music-room$|^SHA256SUMS$'`
  - Files: `packaging/build-macos.sh`

- [x] **T-006:** Extend GitHub Release workflow
  - Deliverable: Job `release` gọi `build-macos.sh`; upload 4+ assets: linux tarball, 2 darwin tarballs, `.deb` x2, standalone binaries (giữ backward compat); tất cả client builds dùng `-X main.version=`; artifact names khớp architecture contract
  - Maps to: AC-033, AC-034, FR-001, FR-008, ADR-002
  - Validation: Review workflow YAML locally; dry-run: `act` optional hoặc `shellcheck packaging/build-macos.sh` + manual script run T-005
  - Files: `.github/workflows/release.yml`

- [x] **T-007:** Homebrew formula
  - Deliverable: `packaging/homebrew/Formula/music-room.rb` với `on_arm` / `on_intel` URL + sha256 placeholders; `depends_on "mpv", "yt-dlp", "ffmpeg"`; `bin.install "music-room"`; `packaging/homebrew/README.md` hướng dẫn `brew install --formula ./packaging/homebrew/Formula/music-room.rb` và tap workflow; script `packaging/homebrew/bump-formula.sh` (optional) cập nhật url/sha256 từ release
  - Maps to: AC-005, AC-006, AC-035, FR-002, ADR-004
  - Validation: `brew ruby -c packaging/homebrew/Formula/music-room.rb` (hoặc `ruby -c` syntax check); manual `brew install --formula` trên macOS khi có artifact
  - Files: `packaging/homebrew/Formula/music-room.rb`, `packaging/homebrew/README.md`

### Phase 3 — Docs & platform support

- [x] **T-008:** README — macOS install + Gatekeeper
  - Deliverable: Section **macOS (13+)** — brew deps, GitHub Release (chọn arm64 vs Intel), Homebrew install, `xattr -dr com.apple.quarantine` hoặc System Settings bypass; cập nhật header scope v1 → V0.2.1 multi-platform; version example `0.2.1`
  - Maps to: AC-030, AC-031, AC-032, FR-007, ADR-005, NFR-007
  - Validation: Manual read-through checklist AC-030–032; `go vet ./...`
  - Files: `README.md`

- [x] **T-009:** README + docs — Debian-based Linux support
  - Deliverable: Section **Supported Linux** — Ubuntu 22.04/24.04 + Debian 12+ / Mint / Pop!_OS; `.deb` install path; tarball path; explicit out-of-scope Fedora/Arch; `docs/PLATFORMS.md` bảng OS/arch/deps hoặc mở rộng `docs/E2E.md` §Platforms
  - Maps to: AC-027, AC-028, AC-029, FR-006, ADR-006
  - Validation: Docs review vs AC-028 bullet list
  - Files: `README.md`, `docs/PLATFORMS.md` (hoặc `docs/E2E.md`)

- [x] **T-010:** Cross-platform E2E documentation
  - Deliverable: `docs/E2E.md` §Cross-platform matrix (host macOS ↔ guest Linux, ngược lại); prerequisites macOS (`brew install mpv yt-dlp ffmpeg`); manual steps cho AC-018–026; note `MUSIC_ROOM_NO_PLAYBACK=1` cho headless; link `docs/PLATFORMS.md`
  - Maps to: AC-018–026, AC-014–017 (manual Terminal.app), NFR-001–003
  - Validation: Checklist trong doc khớp spec AC-018–026
  - Files: `docs/E2E.md`, `docs/PLATFORMS.md`

### Phase 4 — CI & verification

- [x] **T-011:** CI — `macos-latest` test job
  - Deliverable: Thêm job `test-macos` (hoặc matrix) trong `.github/workflows/release.yml` on `pull_request` + `push` main: `go test -short ./...`, `go vet ./...`, `MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh`; `CGO_ENABLED=0 go build ./cmd/music-room`
  - Maps to: AC-014, AC-016, FR-004, NFR-008, architecture cross-platform test matrix
  - Validation: Push branch / local `go test -short ./...`; `MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh`
  - Files: `.github/workflows/release.yml` (hoặc `.github/workflows/ci.yml` mới)

- [x] **T-012:** Linux release ldflags parity
  - Deliverable: `packaging/build-deb.sh` và linux tarball step trong release workflow dùng cùng `-X main.version=` như T-001; verify `.deb` binary `--version` sau build local
  - Maps to: AC-034, FR-008
  - Validation: `./packaging/build-deb.sh 0.2.1-test && dpkg-deb -x dist/music-room_0.2.1-test-1_amd64.deb /tmp/mr-deb && /tmp/mr-deb/usr/bin/music-room --version | grep 0.2.1-test`
  - Files: `packaging/build-deb.sh`, `.github/workflows/release.yml`

- [x] **T-013:** Full regression gate
  - Deliverable: `go test ./...`, `go vet ./...`, build linux + darwin arm64/amd64 client; không regression server packages; tasks.md progress updated
  - Maps to: AC-008–013 (parity via green tests), AC-044 inherited, NFR-004, NFR-009
  - Validation: `go test ./... && go vet ./... && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/music-room && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ./cmd/music-room`
  - Files: (whole repo)

- [x] **T-014:** Review artifact + AC evidence prep
  - Deliverable: `review.md` draft — map AC-001–035 → evidence (build output, brew install log, manual cross-platform session notes); tick tasks complete; `./bin/vibe validate --strict` pass
  - Maps to: Gate G6 prep, all AC
  - Validation: `./bin/vibe validate --strict`
  - Files: `docs/vibe/003-macos-cross-platform/review.md`

## Dependency graph (summary)

```
T-001 → T-005 → T-006 → T-007
T-001 → T-012 → T-006
T-002 → T-003 → T-013
T-004 → T-013
T-005 → T-006
T-006 → T-007 (formula URLs need release asset names)
T-008, T-009, T-010 (docs — parallel sau T-005)
T-011 (CI — sau T-003, T-004)
T-013 → T-014
```

## AC coverage map (by task)

| Task | Primary AC |
|------|------------|
| T-001, T-012 | AC-034, AC-035 |
| T-002, T-003 | AC-004, AC-029 |
| T-004 | AC-001, AC-002, AC-008, AC-013, AC-014 |
| T-005, T-006 | AC-001, AC-002, AC-033 |
| T-007 | AC-005, AC-006, AC-035 |
| T-008 | AC-030, AC-031, AC-032 |
| T-009 | AC-027, AC-028, AC-029 |
| T-010 | AC-018–026, AC-014–017 (manual) |
| T-011 | AC-014, AC-016 (automated smoke) |
| T-013–014 | AC-008–013 parity, all AC evidence |

### Parity AC verified via inheritance (no duplicate tasks)

| AC group | Verification |
|----------|--------------|
| AC-008–013 | Green `go test ./...` on linux + darwin build (T-004, T-013) + manual cross-platform E2E (T-010) |
| AC-014–017 | `tui-smoke.sh` on macOS CI (T-011) + manual Terminal.app (T-010) |
| AC-018–026 | Manual/scripted cross-platform matrix in `docs/E2E.md` (T-010) |

## Progress

| Phase | Total | Done | Blocked |
|-------|-------|------|---------|
| 1 — Foundation | 4 | 4 | 0 |
| 2 — Release & packaging | 3 | 3 | 0 |
| 3 — Docs & platform | 3 | 3 | 0 |
| 4 — CI & verify | 4 | 4 | 0 |
| **Total** | **14** | **14** | **0** |

## Gate G4 checklist

- [x] Tasks ordered by dependency
- [x] Each task has deliverable + validation
- [x] Each task maps to AC/requirement
- [x] No vague tasks
