# Tasks: Room Password & Host Kick

**Slug:** `room-password-host-kick`
**Status:** complete
**Gate G4:** ✅ pass

## Task list

> Execute in order. Mark `[x]` when done. Add `<!-- BLOCKED: reason -->` if stuck.

### Phase 1 — Foundation

- [x] **T-001:** Password validation and bcrypt helpers
  - Deliverable: `internal/server/room/password.go` — `ValidatePassword` (trim, 1–32 chars, reject whitespace-only), `HashPassword`, `CheckPassword`; `password_test.go` table-driven (valid, empty open-room path, too long, whitespace-only); add `golang.org/x/crypto/bcrypt` to `go.mod`
  - Maps to: AC-002, AC-005, FR-002, FR-005, NFR-002, ADR-001
  - Validation: `go test ./internal/server/room/... -run Password`
  - Files: `internal/server/room/password.go`, `internal/server/room/password_test.go`, `go.mod`, `go.sum`

- [x] **T-002:** Protocol types for password and kick
  - Deliverable: `RoomCreatePayload` / `RoomJoinPayload` + optional `Password` field; `MsgRoomKick`, `MsgRoomKicked`, `RoomKickPayload`, `RoomKickedPayload`; `RoomSnapshot.PasswordProtected bool`; `ErrAuthFailed`, `ErrAuthRequired` in `errors.go` + `KnownErrorCodes`; encode/decode tests in `envelope_test.go` if needed
  - Maps to: AC-003, AC-008, AC-014, AC-021, FR-004, FR-008, FR-013, ADR-002, ADR-004
  - Validation: `go test ./internal/protocol/...`
  - Files: `internal/protocol/messages.go`, `internal/protocol/errors.go`, `internal/protocol/types.go`, `internal/protocol/envelope_test.go`

- [x] **T-003:** Room aggregate password storage
  - Deliverable: unexported `passwordHash []byte` on `Room`; `PasswordProtected()`, `SetPassword(plain)`, `CheckPassword(plain)`; `Snapshot()` sets `password_protected`; `ErrAuthFailed`, `ErrAuthRequired` in `room/errors.go`; unit tests — open room, protected room, wrong password check
  - Maps to: AC-001, AC-002, AC-007, AC-025, FR-001, FR-002, FR-021, ADR-001, ADR-002
  - Validation: `go test ./internal/server/room/... -run Room`
  - Files: `internal/server/room/room.go`, `internal/server/room/errors.go`, `internal/server/room/room_test.go` (or extend existing)

### Phase 2 — Core implementation (server)

- [x] **T-004:** Manager create/join with password gate
  - Deliverable: `Manager.Create` accepts optional password (empty → open room); `Manager.Join` checks password before `AddMember` — missing password on protected room → `ErrAuthRequired`; wrong → `ErrAuthFailed`; open room join unchanged
  - Maps to: AC-001, AC-006, AC-007, AC-010, AC-011, AC-013, FR-001, FR-006–FR-012, NFR-005, ADR-003, ADR-007
  - Validation: `go test ./internal/server/room/... -run Manager`
  - Files: `internal/server/room/manager.go`, `internal/server/room/manager_test.go`

- [x] **T-005:** Manager KickMember domain logic
  - Deliverable: `KickMember(slug, hostSessionID, targetSessionID)` — verify host, target exists, target not host; `RemoveMember`; return kick result for hub broadcast; tests for forbidden cases (non-host, kick self, kick host row)
  - Maps to: AC-014, AC-017, AC-019, AC-020, FR-013, FR-016, FR-017, ADR-004, ADR-007
  - Validation: `go test ./internal/server/room/... -run Kick`
  - Files: `internal/server/room/manager.go`, `internal/server/room/manager_test.go`

- [x] **T-006:** Hub handlers — password create/join + kick
  - Deliverable: `handleRoomCreate` / `handleRoomJoin` read password from payload, call `SetPassword` / join gate; `handleRoomKick` + `forceLeaveMember` — send `room.kicked` to target via `clientRegistry`, `room.member_left` to others, clear target `sess.RoomSlug`; register `MsgRoomKick` in `dispatchMessage`; `roomErrorCode` maps auth errors; **no password in logs**
  - Maps to: AC-003, AC-006, AC-011, AC-012, AC-014, AC-015, AC-021, AC-022, FR-003, FR-010, FR-014, FR-018, NFR-001, NFR-004, ADR-003, ADR-004
  - Validation: `go test ./internal/server/hub/... -run Room`
  - Files: `internal/server/hub/handlers_room.go`, `internal/server/hub/handlers_room_test.go`

### Phase 3 — Client core

- [x] **T-007:** Client store and actions
  - Deliverable: `Store.Apply` case `room.kicked` → `InRoom=false`, clear `Room`, surface kick message for TUI; `actions.Room` methods `Create(slug, password)`, `Join(slug, password)`, `Kick(targetSessionID)` sending correct payloads; handle `AUTH_REQUIRED` / `AUTH_FAILED` via `LastErr`
  - Maps to: AC-021, AC-022, AC-023, FR-018, FR-019, ADR-004
  - Validation: `go test ./internal/client/state/... ./internal/client/actions/...`
  - Files: `internal/client/state/store.go`, `internal/client/state/store_test.go`, `internal/client/actions/room.go`, `internal/client/actions/room_test.go`

- [x] **T-008:** CLI `--password` on create and join
  - Deliverable: Cobra `StringVar` `--password` on `create` and `join`; pass into `RoomCreatePayload` / `RoomJoinPayload`; help text warns shell history (AC-027); existing create/join without flag unchanged
  - Maps to: AC-003, AC-008, AC-027, FR-004, FR-008, FR-022, ADR-006
  - Validation: `go test ./internal/client/cli/... -run -E 'Create|Join|Room'`
  - Files: `internal/client/cli/room.go`, `internal/client/cli/commands_test.go`

### Phase 4 — TUI integration

- [x] **T-009:** TUI masked password modal
  - Deliverable: `modals/password.go` — `textinput` with `EchoPassword`, prompt label for create/join; wire into TUI flow when slug join without CLI password (retry on `AUTH_REQUIRED`); modal tests
  - Maps to: AC-004, AC-009, AC-010, AC-026, FR-004, FR-008, FR-022, ADR-006
  - Validation: `go test ./internal/client/tui/modals/... -run Password`
  - Files: `internal/client/tui/modals/password.go`, `internal/client/tui/modals/password_test.go`, `internal/client/tui/model.go`, `internal/client/tui/update.go`, `internal/client/cli/tui.go` (if join entry needs password)

- [x] **T-010:** TUI member selection in CREW panel
  - Deliverable: `selectedMemberIdx` on `Model`; Up/Down when `FocusMembers` moves selection (mirror queue); `panels.Members` highlights selected row (`>`); `RenderOpts.MembersSelectedIdx`; clamp on member list changes; focus tests
  - Maps to: AC-014, AC-017, ADR-005
  - Validation: `go test ./internal/client/tui/... -run -E 'Focus|Members'`
  - Files: `internal/client/tui/model.go`, `internal/client/tui/update.go`, `internal/client/tui/panels/members.go`, `internal/client/tui/panels/options.go`, `internal/client/tui/focus_test.go`, `internal/client/tui/panels/panels_test.go`

- [x] **T-011:** TUI host kick keys and kicked exit
  - Deliverable: `K` and `Del` when `FocusMembers && IsHost` → `actions.Kick(selectedMember.SessionID)`; skip/no-op if selected is host; member sees no kick hint in help; on `room.kicked` show message and exit TUI (`tea.Quit` or leave screen); update `modals/help.go` and host statusbar hint
  - Maps to: AC-014, AC-016, AC-018, AC-021, AC-022, FR-013, FR-015, FR-016, FR-018, ADR-005
  - Validation: `go test ./internal/client/tui/... -run -E 'Kick|Kicked'`
  - Files: `internal/client/tui/update.go`, `internal/client/tui/update_test.go`, `internal/client/tui/keys/keymap.go`, `internal/client/tui/modals/help.go`

### Phase 5 — Integration & verification

- [x] **T-012:** Hub integration tests — password and kick E2E
  - Deliverable: integration tests — create protected room, join wrong password (no member count increase), join correct password, host kick member (target disconnected, others see `member_left`), kicked user re-joins successfully; non-host kick → `FORBIDDEN`
  - Maps to: AC-006, AC-011, AC-014, AC-015, AC-019, AC-023, AC-024, FR-006, FR-010, FR-014, FR-019, FR-020
  - Validation: `go test ./internal/server/hub/... -run Integration -count=1`
  - Files: `internal/server/hub/integration_test.go`, `internal/server/hub/handlers_room_test.go`

- [x] **T-013:** Full regression and spec trace checklist
  - Deliverable: `go test ./...` and `go vet ./...` pass; manually tick AC checklist in `spec.md` or note in `review.md` prep; confirm open rooms backward compatible (AC-001, AC-007, NFR-005)
  - Maps to: all AC-001–AC-027, GATE G5 prep
  - Validation: `go test ./...` && `go vet ./...` && `go build ./...`
  - Files: (verification only); update checkboxes in `docs/vibe/005-room-password-host-kick/tasks.md`

## Progress

| Phase | Total | Done | Blocked |
|-------|-------|------|---------|
| 1 — Foundation | 3 | 3 | 0 |
| 2 — Core (server) | 3 | 3 | 0 |
| 3 — Client core | 2 | 2 | 0 |
| 4 — TUI | 3 | 3 | 0 |
| 5 — Integration | 2 | 2 | 0 |
| **Total** | **13** | **13** | **0** |

## AC coverage matrix

| AC | Task(s) |
|----|---------|
| AC-001 | T-003, T-004, T-013 |
| AC-002 | T-001, T-003 |
| AC-003 | T-002, T-006, T-008 |
| AC-004 | T-009 |
| AC-005 | T-001 |
| AC-006 | T-004, T-012 |
| AC-007 | T-003, T-004, T-013 |
| AC-008 | T-008 |
| AC-009 | T-009 |
| AC-010 | T-004, T-009 |
| AC-011 | T-004, T-006, T-012 |
| AC-012 | T-006 |
| AC-013 | T-004 |
| AC-014 | T-005, T-006, T-010, T-011, T-012 |
| AC-015 | T-006, T-012 |
| AC-016 | T-011 |
| AC-017 | T-005, T-010 |
| AC-018 | T-011 |
| AC-019 | T-005, T-012 |
| AC-020 | T-005 |
| AC-021 | T-002, T-006, T-007, T-011 |
| AC-022 | T-006, T-007, T-011 |
| AC-023 | T-007, T-012 |
| AC-024 | T-012 |
| AC-025 | T-003 |
| AC-026 | T-009 |
| AC-027 | T-008 |

## Gate G4 checklist

- [x] Tasks ordered by dependency
- [x] Each task has deliverable + validation
- [x] Each task maps to AC/requirement
- [x] No vague tasks
