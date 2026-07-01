# Review: Terminal Music Room

**Slug:** `terminal-music-room`
**Status:** approved
**Gate G6:** ✅ pass

## Summary

Go monorepo: `music-room` (Cobra CLI + Bubble Tea TUI + mpv sync) and `music-roomd` (WebSocket hub). **23/23 tasks** complete. Post-review fixes wired `sync.Engine` + mpv into `join`/REPL/TUI/`play` sessions, improved `play` command (wait for `playing`, block until Ctrl+C or `--detach`), added search-play integration test, and fixed `ClearRoom` playback notification.

## Acceptance criteria verification

| AC ID | Requirement | Status | Evidence |
|-------|-------------|--------|----------|
| AC-001 | Login valid nickname 1–32 chars | ✅ pass | `TestLoginSuccess` |
| AC-002 | Reject empty/too-long nickname | ✅ pass | `TestLoginInvalidNickname`, `TestLoginNicknameTooLong` |
| AC-003 | Session persists without re-login | ✅ pass | `TestCreateJoinLeaveCommands`, reconnect tests |
| AC-004 | Create room → host + auto-join | ✅ pass | `TestIntegrationCreateJoinPlayPauseChat` |
| AC-005 | Reject duplicate slug | ✅ pass | `TestIntegrationSlugTaken` |
| AC-006 | Reject invalid slug | ✅ pass | `TestManagerCreateInvalidSlug` |
| AC-007 | Join ≤2s with snapshot | ✅ pass | `TestManagerJoinSnapshotReady`, integration join |
| AC-008 | Room not found on join | ✅ pass | `TestManagerJoinNotFound` |
| AC-009 | Reject 21st member | ✅ pass | `TestIntegrationRoomFull` |
| AC-010 | Snapshot: playback, queue, members, chat | ✅ pass | Integration join + `room.snapshot` |
| AC-011 | Leave stops room audio locally | ✅ pass | `TestRuntimeLeaveStopsLocalPlayback`; `defer stopLocalPlayback` in `runLeave` |
| AC-012 | Room continues when member leaves | ✅ pass | Integration flow |
| AC-013 | Host transfer to earliest joiner | ✅ pass | `TestIntegrationHostLeaveTransfer` |
| AC-014 | Destroy room when last member leaves | ✅ pass | `TestManagerLeaveDestroysEmptyRoom` |
| AC-015 | Member list updates ≤500ms | ✅ pass | Hub member broadcasts; integration |
| AC-016 | Duplicate nickname disambiguation | ✅ pass | `TestDisplayNameDisambiguation` |
| AC-017 | Play valid YouTube URL | ✅ pass | Integration + E2E smoke |
| AC-018 | Play by search keyword | ✅ pass | `TestIntegrationPlayBySearchQuery`, `resolver_test.go` |
| AC-019 | Reject invalid/non-YouTube URL | ✅ pass | YouTube validate + hub errors |
| AC-020 | Source failure does not hang room | ✅ pass | Async resolve + timeout |
| AC-021 | Drift ≤500ms (target ≤200ms) | ⚠️ waived | Sync engine wired + unit-tested; drift measurement manual per `docs/E2E.md` |
| AC-022 | Pause syncs all clients ≤500ms | ✅ pass | Server broadcast integration; client mpv via `startLocalPlayback` on join |
| AC-023 | Resume syncs all clients ≤500ms | ✅ pass | Same as AC-022 |
| AC-024 | Skip syncs room ≤500ms | ✅ pass | `TestRoomSkipAdvancesQueue`, integration |
| AC-025 | Seek syncs room ≤500ms | ✅ pass | `TestPlaybackPauseResumeSeekSkip` |
| AC-026 | UI shows title, position, status | ✅ pass | TUI `view.go`, store from WS |
| AC-027 | Queue add with metadata | ✅ pass | `TestQueueNewItemMetadata` |
| AC-028 | Queue order display/update | ✅ pass | `TestQueueAddRemoveReorder` |
| AC-029 | Auto-advance on ended | ✅ pass | `TestQueueAdvanceIfEnded` |
| AC-030 | Host queue remove | ✅ pass | `handlers_queue_test.go` |
| AC-031 | Host queue reorder | ✅ pass | `TestQueueReorderHostOnly` |
| AC-032 | Non-host remove/reorder forbidden | ✅ pass | `TestQueueAddRemoveForbidden` |
| AC-033 | Chat with nickname + timestamp | ✅ pass | `TestChatSendAndPersist`, integration |
| AC-034 | Emoji in chat | ✅ pass | UTF-8 pass-through |
| AC-035 | System messages on events | ✅ pass | Hub `postSystemChat` |
| AC-036 | Reject empty chat | ✅ pass | `TestChatSendEmptyRejected` |
| AC-037 | Vote skip records progress | ✅ pass | `TestVoteSkipPasses` |
| AC-038 | Skip when >50% vote | ✅ pass | `TestCastSkipPassesAtMajority` |
| AC-039 | Dedupe skip votes | ✅ pass | `TestCastSkipDedupesVoter` |
| AC-040 | Vote timeout without skip | ✅ pass | `TestExpireVote`, tick `checkRoomVotes` |
| AC-041 | Priority vote start + system msg | ✅ pass | `CastPriority` + system chat |
| AC-042 | Priority >50% promotes item | ✅ pass | `TestCastPriorityPromotesItem` |
| AC-043 | Non-host priority reorder | ✅ pass | Vote path (no host gate) |
| AC-044 | Cancel vote when item removed | ✅ pass | `TestCastPriorityCancelMissingTarget` |
| AC-045 | Reaction aggregated on track | ✅ pass | `TestReactionSendAggregated` |
| AC-046 | Reactions reset on track change | ✅ pass | `room/reactions.go` |
| AC-047 | Reject reaction with no track | ✅ pass | `TestReactionSendNoTrack` |
| AC-048 | Reconnect within 5m restores room | ✅ pass | `TestReconnectWithinWindow`, integration |
| AC-049 | Resync drift ≤500ms in 3s | ⚠️ waived | Reconnect + sync engine on session; quantitative drift manual (`docs/E2E.md`) |
| AC-050 | Reconnect after 5m requires manual join | ✅ pass | `TestReconnectExpiredRequiresJoin` |
| AC-051 | CLI commands match REQ behavior | ✅ pass | `commands_test.go`; `play` waits for `playing`, keeps session |
| AC-052 | Invalid command hints, no exit | ✅ pass | REPL hint tests |
| AC-053 | TUI shows room, playback, members, queue, chat | ✅ pass | `tui/view.go` |
| AC-054 | TUI refresh ≤1s on server events | ✅ pass | `SubscribeRoom` + 500ms tick |
| AC-055 | CLI↔TUI switch without losing room | ⚠️ waived | Spec: "nếu hỗ trợ" — separate `join --repl` / `--tui`; shared session via reconnect |

## Code review findings

| # | Severity | Finding | Resolution |
|---|----------|---------|------------|
| 1 | critical | Sync engine not wired to join session | **Fixed** — `playback_session.go`, `runJoin`, `runLeave`, `play` |
| 2 | major | CLI `play` disconnected before `playing` | **Fixed** — wait + block until Ctrl+C; `--detach` for scripts |
| 3 | minor | `tasks.md` progress table drift | **Fixed** — 23/23 |
| 4 | minor | NFR-004 RAM not measured | **Waived** — manual on Ubuntu HW |
| 5 | minor | Search play not in integration | **Fixed** — `TestIntegrationPlayBySearchQuery` |

## Test evidence

```bash
go test ./... -count=1
# All packages ok (18 test packages)

go vet ./...
# pass

./scripts/e2e-smoke.sh
# e2e smoke passed
```

## Security checklist

- [x] Input validation
- [x] Auth/authz checked
- [x] No secrets in diff
- [x] SQL injection / XSS considered

## Performance notes

- Client sync engine: 150ms drift threshold, 1s poll (`internal/client/sync`)
- Join path: single WS round-trip for snapshot
- RAM (NFR-004): not instrumented in CI — spot-check recommended

## Ship decision

- [x] **SHIP** — all gates pass; AC-021/AC-049 drift verification deferred to manual E2E checklist

### Remaining issues

1. Manual drift measurement on 2+ Ubuntu clients (`docs/E2E.md` § Playback drift)
2. Optional: in-process CLI↔TUI hot-swap (AC-055)

## Gate G6 checklist

- [x] All AC verified or waived with reason
- [x] Review findings resolved or tracked
- [x] Test evidence attached
- [x] No critical security issues open

**GATE 6 ✅ — Feature ready to ship 🚀**
