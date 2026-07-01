# Tasks: Room Host Sci-Fi TUI (v2)

**Slug:** `room-host-sci-fi-tui`
**Status:** complete
**Gate G4:** ✅ pass

## Task list

> Execute in order. Mark `[x]` when done. Add `<!-- BLOCKED: reason -->` if stuck.

### Phase 1 — Foundation

- [x] **T-001:** Extract `internal/client/actions` package
  - Deliverable: `actions.Room` với `Play`, `Pause`, `Resume`, `Skip`, `Seek`, `QueueAdd`, `QueueRemove`, `QueueReorder`, `Chat`, `VoteSkip`, `VotePriority`, `React`, `Leave`; `parseSourceArgs` + validation (empty chat, invalid seek) moved từ CLI; unit tests table-driven
  - Maps to: AC-038, AC-044, FR-013, ADR-003
  - Validation: `go test ./internal/client/actions/...`
  - Files: `internal/client/actions/doc.go`, `playback.go`, `queue.go`, `social.go`, `source.go`, `*_test.go`

- [x] **T-002:** Refactor CLI/REPL to delegate `actions.Room`
  - Deliverable: `repl.go`, `repl_actions.go`, `queue.go`, `playback.go` gọi `actions.*`; `Runtime` helper `Actions()`; **không đổi** hành vi CLI — existing `commands_test.go` pass
  - Maps to: AC-044, AC-045, FR-013, ADR-003
  - Validation: `go test ./internal/client/cli/...`
  - Files: `internal/client/cli/repl.go`, `repl_actions.go`, `queue.go`, `playback.go`, `runtime.go`

- [x] **T-003:** Cyberpunk 16-color theme
  - Deliverable: `internal/client/tui/theme/cyberpunk.go` — semantic roles → ANSI indices; `Panel(focused)`, `Header()`, `Title()`, `Muted()`, `Error()`, `ProgressBar(filled,total)`; snapshot test cho style strings
  - Maps to: AC-005, AC-007, AC-008, FR-002, NFR-001, NFR-008, ADR-004
  - Validation: `go test ./internal/client/tui/theme/...`
  - Files: `internal/client/tui/theme/cyberpunk.go`, `cyberpunk_test.go`

- [x] **T-004:** HUD layout engine + degraded mode
  - Deliverable: `internal/client/tui/layout/hud.go` — `Compute(width,height) → Regions`; baseline 80×24; `Degraded` khi &lt;80×24 (AC-011); ASCII border fallback flag; tests matrix 80×24, 120×40, 60×20
  - Maps to: AC-009, AC-010, AC-011, FR-003, NFR-001
  - Validation: `go test ./internal/client/tui/layout/...`
  - Files: `internal/client/tui/layout/hud.go`, `hud_test.go`

### Phase 2 — Sci-Fi HUD (read + chat)

- [x] **T-005:** Panel renderers
  - Deliverable: `panels/` — `header`, `now_playing` (ASCII bar), `members` (host `*`), `queue` (scroll offset), `chat`, `vote`, `reactions`, `statusbar`; mỗi panel nhận `theme` + `state.View` + dimensions; golden render tests với fixture View
  - Maps to: AC-009, AC-024, AC-025, AC-028, AC-030–032, AC-034, AC-041, FR-003, FR-007, FR-008, FR-009, FR-010
  - Validation: `go test ./internal/client/tui/panels/...`
  - Files: `internal/client/tui/panels/*.go`, `panels/*_test.go`

- [x] **T-006:** Integrate HUD view (replace v1 wireframe layout)
  - Deliverable: `view.go` compose theme+layout+panels; xóa lipgloss styles cũ (205/241/RoundedBorder monolith); `View()` render full dashboard; `view_test.go` golden 80×24 và 120×40
  - Maps to: AC-001, AC-002, AC-005, AC-009, AC-040, FR-001, FR-002, FR-003, ADR-001, ADR-002
  - Validation: `go test ./internal/client/tui/... -run View`
  - Files: `internal/client/tui/view.go`, `view_test.go`

- [x] **T-007:** Model state machine + `actions` wiring
  - Deliverable: `model.go` — `Mode` (dashboard/modal/help), `FocusPanel`, `SelectedQueueIdx`, `QueueScroll`, `IsHost(view)`; `Config.Actions *actions.Room`; giữ store subscribe + 500ms tick; `q` quit **không** leave (AC-004)
  - Maps to: AC-004, AC-037, AC-039, ADR-002
  - Validation: `go test ./internal/client/tui/... -run Model`
  - Files: `internal/client/tui/model.go`, `run.go`

- [x] **T-008:** Chat input + send via actions
  - Deliverable: `update.go` — `Enter` gửi chat qua `actions.Chat`; reject empty (AC-027); error toast từ `LastErr`; chat scroll khi tin mới (AC-029)
  - Maps to: AC-026, AC-027, AC-028, AC-029, FR-008
  - Validation: `go test ./internal/client/tui/... -run Chat`
  - Files: `internal/client/tui/update.go`, `update_test.go`

### Phase 3 — Full parity controls

- [x] **T-009:** Keymap + playback shortcuts
  - Deliverable: `keys/keymap.go` + `update.go` handlers — `Space` pause/resume, `s` skip, guard khi no track (AC-016); gọi `actions.Pause/Resume/Skip`
  - Maps to: AC-012–016, FR-004, ADR-006
  - Validation: `go test ./internal/client/tui/... -run Playback`
  - Files: `internal/client/tui/keys/keymap.go`, `internal/client/tui/update.go`

- [x] **T-010:** Modals — add/play source + queue add
  - Deliverable: `modals/add_source.go` — `a` mở overlay textinput URL/query; submit → `actions.Play` hoặc `actions.QueueAdd`; invalid source toast (AC-019); `Esc` đóng
  - Maps to: AC-017, AC-019, FR-005, ADR-002
  - Validation: `go test ./internal/client/tui/modals/...`
  - Files: `internal/client/tui/modals/add_source.go`, `modals/*_test.go`

- [x] **T-011:** Modal seek + help overlay
  - Deliverable: `modals/seek.go` — `S` nhập position ms → `actions.Seek`; `?` help overlay liệt kê shortcuts (ADR-006); `Esc`/`?` toggle
  - Maps to: AC-015, FR-004, ADR-006
  - Validation: `go test ./internal/client/tui/modals/... -run Seek`
  - Files: `internal/client/tui/modals/seek.go`, `internal/client/tui/modals/help.go`

- [x] **T-012:** Vote + reaction shortcuts
  - Deliverable: `v` → `actions.VoteSkip`; `V` → `actions.VotePriority(selected item id)`; `1`–`4` quick react 🔥❤️😂👍; no-track guard (AC-036); vote progress từ `Room.Vote` (AC-030–033)
  - Maps to: AC-030–036, AC-033, FR-009, FR-010
  - Validation: `go test ./internal/client/tui/... -run Vote`
  - Files: `internal/client/tui/update.go`, `internal/client/tui/panels/vote.go`, `reactions.go`

- [x] **T-013:** Host queue admin + role gating
  - Deliverable: `d` remove selected → `actions.QueueRemove`; `Ctrl+↑/↓` reorder → `actions.QueueReorder`; member: keys disabled + toast (AC-022); host-only hint trên statusbar; cập nhật khi `host_changed` (AC-039)
  - Maps to: AC-020–022, AC-037–039, FR-006, FR-011, ADR-006
  - Validation: `go test ./internal/client/tui/... -run Queue`
  - Files: `internal/client/tui/update.go`, `internal/client/tui/panels/queue.go`

- [x] **T-014:** Leave confirm + connection UX
  - Deliverable: `l` → confirm modal → `actions.Leave`; header conn badge `connected`/`reconnecting`/`disconnected` từ `view.Status` (AC-041–043); host leave updates crew panel (AC-047)
  - Maps to: AC-041–043, AC-046, AC-047, FR-012, FR-014
  - Validation: `go test ./internal/client/tui/... -run Leave`
  - Files: `internal/client/tui/modals/confirm_leave.go`, `internal/client/tui/panels/header.go`, `update.go`

- [x] **T-015:** Focus navigation + queue scroll
  - Deliverable: `Tab`/`Shift+Tab` cycle focus; `↑/↓` scroll queue/members/chat khi focused; micro-feedback focus border magenta (AC-008)
  - Maps to: AC-008, AC-010, NFR-004
  - Validation: `go test ./internal/client/tui/... -run Focus`
  - Files: `internal/client/tui/update.go`, `keys/keymap.go`

### Phase 4 — Entry, docs & verification

- [x] **T-016:** CLI entry — join defaults + `music-room tui`
  - Deliverable: `join --tui` default `true`, `--repl` default `false`; `music-room tui [slug]` command; wire `tui.Config{Actions: rt.Actions()}`; `create` unchanged; test join flag defaults
  - Maps to: AC-001, AC-003, FR-001, ADR-005
  - Validation: `go test ./internal/client/cli/... -run Join`
  - Files: `internal/client/cli/room.go`, `internal/client/cli/tui.go`, `commands_test.go`

- [x] **T-017:** End-to-end TUI smoke script
  - Deliverable: `scripts/tui-smoke.sh` — start `music-roomd`, join `--tui` headless-friendly check (hoặc `MUSIC_ROOM_NO_PLAYBACK=1` + timeout); document manual steps cho AC-006/AC-023 trong `docs/E2E.md` §TUI v2
  - Maps to: AC-006, AC-023, NFR-002
  - Validation: `MUSIC_ROOM_NO_PLAYBACK=1 ./scripts/tui-smoke.sh` (hoặc documented manual checklist)
  - Files: `scripts/tui-smoke.sh`, `docs/E2E.md`

- [x] **T-018:** README + breaking change note
  - Deliverable: `README.md` — quickstart `join` mở TUI sci-fi; keymap table; `--repl` fallback; screenshot ASCII hoặc mô tả HUD; note `q` vs `l`
  - Maps to: AC-001, AC-004, ADR-005, NFR-007
  - Validation: `go test ./... && go vet ./...`
  - Files: `README.md`

- [x] **T-019:** Full regression + lint gate
  - Deliverable: Tất cả packages green; không regression server tests; `go build ./cmd/music-room`
  - Maps to: AC-044, AC-045, NFR-005, NFR-006, NFR-009
  - Validation: `go test ./... && go vet ./... && go build -o bin/music-room ./cmd/music-room`
  - Files: (whole client/tui tree)

- [x] **T-020:** Review artifact + AC trace prep
  - Deliverable: `review.md` draft — checklist map AC-001–047 → evidence (test output, manual AC-006 session); tick tasks.md complete
  - Maps to: Gate G6 prep, all AC
  - Validation: `./bin/vibe validate --strict`
  - Files: `docs/vibe/002-room-host-sci-fi-tui/review.md`

## Dependency graph (summary)

```
T-001 → T-002 → T-007 → T-008
T-003 → T-005 → T-006
T-004 → T-005 → T-006
T-006 + T-007 → T-009 → T-010 → T-011 → T-012 → T-013 → T-014 → T-015
T-002 + T-007 → T-016
T-009..T-016 → T-017 → T-018 → T-019 → T-020
```

## AC coverage map (by task)

| Task | Primary AC |
|------|------------|
| T-001–002 | AC-044, AC-045 |
| T-003–004 | AC-005–011 |
| T-005–006 | AC-009, AC-024–025, AC-028, AC-030–034, AC-041 |
| T-007–008 | AC-004, AC-026–029, AC-037, AC-039 |
| T-009 | AC-012–016 |
| T-010 | AC-017, AC-019 |
| T-011 | AC-015 |
| T-012 | AC-030–036, AC-033 |
| T-013 | AC-020–022, AC-037–039 |
| T-014 | AC-041–043, AC-046–047 |
| T-015 | AC-008, AC-010 |
| T-016 | AC-001, AC-003 |
| T-017–018 | AC-006, AC-023 |
| T-019–020 | AC-044–045, all AC evidence |

## Progress

| Phase | Total | Done | Blocked |
|-------|-------|------|---------|
| 1 — Foundation | 4 | 4 | 0 |
| 2 — Sci-Fi HUD | 4 | 4 | 0 |
| 3 — Full parity | 7 | 7 | 0 |
| 4 — Entry & verify | 5 | 5 | 0 |
| **Total** | **20** | **20** | **0** |

## Gate G4 checklist

- [x] Tasks ordered by dependency
- [x] Each task has deliverable + validation
- [x] Each task maps to AC/requirement
- [x] No vague tasks
