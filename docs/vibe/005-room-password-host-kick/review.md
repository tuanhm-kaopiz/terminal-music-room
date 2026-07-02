# Review: Room Password & Host Kick

**Slug:** `room-password-host-kick`
**Status:** approved
**Gate G6:** ✅ pass

## Summary

Implemented optional **room password** (bcrypt server-side, `password_protected` on snapshot) and **host kick** (`room.kick` / `room.kicked`) across hub, protocol, CLI, and TUI. Open rooms remain backward compatible. Primary touchpoints:

| Layer | Files |
|-------|-------|
| Server password | `internal/server/room/password.go`, `room.go`, `manager.go` |
| Hub handlers | `internal/server/hub/handlers_room.go` |
| Protocol | `internal/protocol/messages.go`, `errors.go`, `types.go` |
| Client | `internal/client/cli/room.go`, `cli/tui.go`, `actions/social.go`, `state/store.go` |
| TUI | `modals/password.go`, `panels/members.go`, `update.go`, `keys/keymap.go` |

**Note:** Working tree may contain unrelated diffs from other in-progress features; this review scopes verification to password/kick behavior above.

## Acceptance criteria verification

| AC ID | Requirement | Status | Evidence |
|-------|-------------|--------|----------|
| AC-001 | Open room create (no password) | ✅ pass | `TestManagerCreateAndDuplicateSlug`; `IsEmptyPassword` → no hash; `handlers_room_test` join open room |
| AC-002 | Protected room create; host admitted without re-entering password | ✅ pass | `TestManagerCreateWithPassword`; create flow admits host before join gate |
| AC-003 | CLI `--password` on create | ✅ pass | `room.go` flags + `RoomCreatePayload.Password`; integration create with password |
| AC-004 | TUI masked prompt on **create** | ⚠️ waived | `PasswordCreate` intent exists in `modals/password.go` but **not wired** — `music-room create` is CLI-only and does not launch TUI. Join path covered (AC-009). Defer create-via-TUI to backlog. |
| AC-005 | Reject invalid password (>32, whitespace-only) | ✅ pass | `TestValidatePassword`; `ValidationError` on `SetPassword` |
| AC-006 | Join protected room with correct password | ✅ pass | `TestManagerJoinCorrectPassword`; `TestIntegrationPasswordJoinAndKick` |
| AC-007 | Join open room without password | ✅ pass | `TestManagerJoinSnapshotReady`; existing hub join tests unchanged |
| AC-008 | CLI `--password` on join | ✅ pass | `room.go` / `tui.go` pass password in `RoomJoinPayload` |
| AC-009 | TUI masked join prompt | ✅ pass | `modals/password.go` `EchoPassword`; `music-room tui <slug>` → `PendingJoinSlug` + `ModeModalPassword`; `TestPasswordMaskedInput` |
| AC-010 | Reject join without password on protected room | ✅ pass | `TestManagerJoinPasswordRequired` → `ErrAuthRequired` / `AUTH_REQUIRED` |
| AC-011 | Wrong password rejected, no room content | ✅ pass | `TestManagerJoinWrongPassword`; integration wrong-password → `AUTH_FAILED`, no snapshot |
| AC-012 | Error message does not leak correct password | ✅ pass | Generic `authentication failed` / `AUTH_FAILED`; no password in `RoomSnapshot` |
| AC-013 | No lockout on repeated wrong password | ✅ pass | No lockout code path; per-attempt reject only (architecture ADR-007) |
| AC-014 | Host kick disconnects member ≤3s | ✅ pass | `TestIntegrationPasswordJoinAndKick` kick leg passes in ~0.19s; TUI `K`/`Del` → `actions.Kick` |
| AC-015 | Kicked member removed from others' member list | ✅ pass | Integration `room.member_left` broadcast; `applyMemberLeft` |
| AC-016 | Kick without confirm dialog | ✅ pass | `handleRoomKick` immediate; no modal |
| AC-017 | No kick target when only host | ✅ pass | `handleMemberKick` no-op on empty/invalid selection; host row blocked |
| AC-018 | Member cannot kick (no effective action) | ✅ pass | `handleMemberKick` → `ErrHostOnly`; help kick section host-only |
| AC-019 | Server rejects non-host kick | ✅ pass | `TestManagerKickMember` non-host → `ErrForbidden`; integration guest kick → `FORBIDDEN` |
| AC-020 | Host cannot kick self | ✅ pass | `TestManagerKickMember` self kick → `ErrForbidden`; client blocks host row |
| AC-021 | Kicked member sees clear message | ✅ pass | `room.kicked` payload; `TestApplyRoomKicked`; TUI sets `KickedMessage` |
| AC-022 | Kicked client leaves TUI / room session | ✅ pass | `applyRoomKicked` clears `InRoom`; `model.refresh` → `tea.Quit` |
| AC-023 | Re-join after kick allowed | ✅ pass | Integration re-join after kick with password |
| AC-024 | Re-join = fresh session state (no stale vote/reaction) | ⚠️ waived | Same `session_id` re-joins; room-level vote/reaction maps are **not** purged per-member on kick/leave today. Acceptable v1 — vote tallies may still include prior session votes until vote reset. Track for follow-up if strict per-session isolation required. |
| AC-025 | Password not shown in TUI panels | ✅ pass | Only `password_protected` bool on snapshot; grep confirms no password field in panels/header/help |
| AC-026 | TUI password input masked | ✅ pass | `textinput.EchoPassword` in `modals/password.go`; `TestPasswordMaskedInput` |
| AC-027 | CLI help warns shell history | ✅ pass | `create`/`join` `Long` + flag help text mention shell history |

## Code review findings

| # | Severity | Finding | Resolution |
|---|----------|---------|------------|
| 1 | minor | **AC-004 gap:** TUI create-with-password not wired (`PasswordCreate` unused) | **tracked** — CLI create + TUI join modal sufficient for v1 per clarify ADR-006 |
| 2 | minor | **AC-024:** Vote/reaction state not cleared per kicked session on re-join | **tracked** — existing room-level vote model; no ban/kick registry by design |
| 3 | minor | Password sent plaintext over WebSocket (no TLS in v1) | **waived** — documented in `architecture.md` ADR-001; LAN/self-hosted assumption |
| 4 | minor | `music-room join` with `--tui` still joins before TUI without password modal if slug known | **tracked** — use `music-room tui <slug>` for masked prompt; or pass `--password` |
| 5 | info | No dedicated `update_test.go` kick key test | Acceptable — integration + `handleMemberKick` logic reviewed; host gating in code |

No **critical** or **major** findings. Authz enforced server-side (`KickMember`, `IsHost`). Password stored as bcrypt hash only.

## Test evidence

```bash
# Full test suite
go test ./...
# Result: all packages ok (hub integration ~4.5s)

# Feature-focused integration
go test ./internal/server/hub/... -run TestIntegrationPasswordJoinAndKick -v -count=1
# === RUN   TestIntegrationPasswordJoinAndKick
# --- PASS: TestIntegrationPasswordJoinAndKick (0.19s)
# PASS

# Password / room domain
go test ./internal/server/room/... -run 'Password|Manager|Kick|Room' -count=1
# ok

# Client state kick
go test ./internal/client/state/... -run Kicked -count=1
# ok

# TUI password modal
go test ./internal/client/tui/modals/... -run Password -count=1
# ok

# Lint / build
go vet ./...
go build ./...
# Result: pass (exit 0)
```

## Security checklist

- [x] Input validation — `ValidatePassword` 1–32 chars, reject whitespace-only
- [x] Auth/authz — join password check; kick host-only server + client UX gating
- [x] No secrets in diff — no hardcoded passwords; hash only on server
- [x] Injection — N/A (no SQL); WS JSON payloads validated

## Performance notes

- bcrypt on join adds ~50–100ms per join (acceptable per architecture ADR-001).
- Kick path: integration test completes kick + re-join in <200ms on localhost.

## Ship decision

- [x] **SHIP** — all gates pass
- [ ] **HOLD** — issues remain (list below)

### Remaining issues (non-blocking)

1. Wire TUI password modal for `create` flow (AC-004 backlog).
2. Optional: purge vote entries for kicked `session_id` on re-join (AC-024 strict mode).
3. Commit hygiene: isolate feature 005 diff from unrelated working-tree changes before merge.

## Gate G6 checklist

- [x] All AC verified or waived with reason
- [x] Review findings resolved or tracked
- [x] Test evidence attached
- [x] No critical security issues open
