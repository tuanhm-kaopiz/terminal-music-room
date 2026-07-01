# Clarify: Terminal Music Room

**Slug:** `terminal-music-room`
**Status:** complete
**Gate G1:** ✅ pass

## Resolved questions

| # | Question | Answer | Decided by |
|---|----------|--------|------------|
| 1 | v1 deployment model — how do rooms and sync work? | **Managed cloud (SaaS)** — operator hosts the sync server; clients connect to cloud infra | user |
| 2 | Source of truth for playback state? | **Server-authoritative** — central server owns playback clock, position, and play/pause/skip state | user |
| 3 | YouTube audio technical approach? | **Stream extraction** (e.g. yt-dlp/ffmpeg) — accept YouTube ToS/legal risk for OSS dev tool | user |
| 4 | Open-source for v1? | **Yes** — open-source release (license decided in architecture phase) | user |
| 5 | How do users join a room (no passwords/invite links)? | **Global unique room slug** — e.g. `music-room join backend-team` | user |
| 6 | Song reactions (PRD US-011) in full v1? | **Yes** — emoji reactions on current song included in v1 | user |
| 7 | Who can remove or reorder queue items? | **Host only** | user |
| 8 | Who can play/pause/skip/seek and add tracks? | **Any room member** — democratic playback commands; host retains queue admin (remove/reorder) | idea.md (confirmed, not contradicted) |
| 9 | Authentication model for v1? | **Anonymous nickname only** — no OAuth/GitHub/Google | idea.md |
| 10 | Vote skip threshold? | **>50% of online members** (e.g. 5 online → 3 votes) | idea.md / PRD |
| 11 | Client platform for v1? | **Ubuntu Linux only** — macOS, Windows, Docker deferred | idea.md |
| 12 | UI modes required? | **CLI commands + simple TUI** — both supported; visual polish not a release blocker | idea.md |

## Open questions (blocking)

| # | Question | Owner | Blocking? |
|---|----------|-------|-----------|

## Scope

### In scope

- Managed cloud sync server (SaaS) with server-authoritative playback state
- Ubuntu CLI/TUI client — lightweight, terminal-native
- Anonymous login via nickname
- Room lifecycle: create, join (global slug), leave, list online members
- YouTube audio: search by keyword and play by URL via stream extraction
- Synchronized playback: play, pause, resume, skip, seek — any member may issue commands
- Queue: any member adds; host removes/reorders; all members view upcoming tracks
- Text chat with emoji and system messages (join, song change, vote started)
- Voting: skip current song and next-song priority (>50% online members)
- Song reactions (emoji on current track)
- Reconnect after disconnect — auto-resync playback state
- Open-source distribution
- Performance targets: join < 2s, command broadcast < 500ms, sync drift ≤ 500ms (ideal ≤ 200ms), 2–20 users/room, client RAM < 300MB idle

### Out of scope

- Video streaming, voice chat, screen sharing
- Mobile app, browser dashboard
- Music sources other than YouTube (Spotify, SoundCloud, local files, radio)
- OAuth / GitHub / Google login
- macOS, Windows, Docker runtime, self-hosted or P2P deployment models
- Private/password rooms, invite links
- AI DJ, smart playlists, personalized recommendations
- DJ mixing / music production features
- Rich UI polish (themes, animations, advanced a11y beyond terminal baseline)

## Actors / users

| Actor | Role | Key actions |
|-------|------|-------------|
| Room member | Any user who joined a room | Set nickname, join/leave room, play/pause/skip/seek, add to queue, chat, vote skip/priority, react to song |
| Room host | Member who created the room | All member actions + remove/reorder queue items |
| SaaS operator | Runs managed cloud sync server | Host infra, enforce rate limits/abuse protection (operational; not an in-app user role for v1) |
| Anonymous visitor | User not yet in a room | Login with nickname, create or join room by slug |

## Assumptions

1. v1 ships **full feature set** (not a cut-down MVP) with **simple** TUI/CLI — wireframe-level UI is acceptable for release.
2. Room slugs are **globally unique** across the SaaS instance; first creator owns the slug for the room session lifetime.
3. **Server-authoritative sync** means all clients follow server playback clock; clients correct drift locally as needed.
4. Stream extraction dependencies (yt-dlp, ffmpeg, OS audio stack) are acceptable on Ubuntu; breakage from YouTube changes is an operational risk.
5. **Host** is the room creator; host role does not transfer in v1 unless creator leaves (behavior for host leave → spec edge case).
6. Voting counts **online members only** at vote start; abstentions do not count toward the threshold.
7. PRD `docs/spec.md` is the functional reference; where it conflicts with resolved answers above, **clarify.md wins** (e.g. voting and reactions are v1, not Phase 2).
8. Open-source license (MIT, Apache-2.0, etc.) deferred to architecture phase.

## Risks & constraints

| Risk | Impact | Mitigation |
|------|--------|------------|
| YouTube ToS / legal exposure from stream extraction | Takedown, liability, blocked streams | Document risk in architecture; OSS disclaimer; fallback error UX when source unavailable |
| Playback sync complexity (latency, jitter, drift) | Poor UX, core value failure | Server-authoritative clock; reconnect resync; target metrics in spec AC |
| YouTube API/format changes breaking extraction | Playback failures | Pin tool versions; monitor upstream; graceful error + skip/retry |
| SaaS hosting cost and abuse | Runaway infra bill, spam rooms | Rate limiting, abuse protection (PRD §13); room/user limits for v1 |
| Ubuntu-only audio stack variance | Install failures across distros | Target Ubuntu specifically; document deps in architecture |
| Host leaves room | Orphaned queue admin / playback | Define behavior in spec (edge case) |

## Gate G1 checklist

- [x] No blocking open questions
- [x] Scope bounded (in/out explicit)
- [x] Actors identified
- [x] Assumptions listed
