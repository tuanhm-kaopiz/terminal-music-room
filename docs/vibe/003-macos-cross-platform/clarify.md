# Clarify: macOS Cross-Platform Support (V0.2.1)

**Slug:** `macos-cross-platform`
**Status:** complete
**Gate G1:** ✅ pass

## Resolved questions

| # | Question | Answer | Decided by |
|---|----------|--------|------------|
| 1 | Kiến trúc macOS cần hỗ trợ trong V0.2.1? | **Cả hai** — Apple Silicon (arm64) và Intel (amd64) | user |
| 2 | Cách phân phối bản macOS cho end user? | **Cả GitHub Release binary và Homebrew** (formula/tap) | user |
| 3 | Mức feature parity macOS so với Ubuntu hiện tại? | **Full parity** — CLI + sci-fi TUI + toàn bộ tính năng (queue, chat, vote, reactions, sync…) | user |
| 4 | Yêu cầu code signing / notarization cho macOS V0.2.1? | **Không cần** — user chấp nhận Gatekeeper warning hoặc `xattr` workaround | user |
| 5 | Phía Linux trong phòng cross-platform — distro nào in-scope? | **Ubuntu + Debian-based** distros (cùng mức hỗ trợ Linux hiện có, mở rộng Debian family) | user |
| 6 | Host/guest cross-platform — chiều nào bắt buộc pass? | **Hai chiều** — host macOS + guest Linux và host Linux + guest macOS | idea.md (confirmed) |
| 7 | Windows trong V0.2.1? | **Out of scope** — deferred | idea.md |
| 8 | Deployment model thay đổi? | **Không** — vẫn managed cloud SaaS, server-authoritative sync như v1 | inherited from 001 |

## Open questions (blocking)

| # | Question | Owner | Blocking? |
|---|----------|-------|-----------|

## Scope

### In scope

- macOS client (CLI + sci-fi TUI) với **full feature parity** so với Ubuntu client hiện tại
- Hai kiến trúc release: **darwin/arm64** và **darwin/amd64**
- Phân phối: **GitHub Release** artifacts + **Homebrew** formula/tap
- Hướng dẫn cài đặt macOS trong README (bao gồm xử lý Gatekeeper khi không notarize)
- Cross-platform rooms:
  - Host macOS ↔ guest Ubuntu/Debian-based Linux
  - Host Ubuntu/Debian-based Linux ↔ guest macOS
- Luồng end-to-end: create/join room, playback sync, queue, chat, vote, reactions, reconnect
- Mở rộng hỗ trợ Linux sang **Debian-based** distros (ngoài Ubuntu) ở mức tương đương bản Linux hiện tại
- CI/build pipeline cho macOS artifacts trong release V0.2.1
- Performance targets kế thừa từ v1 (join < 2s, broadcast < 500ms, sync drift ≤ 500ms) áp dụng cho macOS client

### Out of scope

- Windows client
- GUI native macOS (không phải terminal/TUI)
- Mac App Store distribution
- Code signing và notarization (V0.2.1)
- Web/mobile client hoặc thay đổi giao thức sync
- Self-hosted / P2P deployment model
- Audio backend optimization vượt mức cần cho parity cross-platform
- Hỗ trợ Linux distros ngoài Ubuntu + Debian-based (Fedora, Arch, v.v.)
- Tính năng mới ngoài scope v1 (chỉ port + cross-platform compatibility)

## Actors / users

| Actor | Role | Key actions |
|-------|------|-------------|
| macOS room member | User chạy `music-room` trên Mac (Terminal) | Host hoặc join phòng, dùng CLI/TUI, playback/queue/chat/vote/reactions như Linux user |
| Linux room member | User trên Ubuntu hoặc Debian-based distro | Host hoặc join phòng; tương tác với macOS peers trong cùng room |
| Room host | Member tạo phòng (macOS hoặc Linux) | Queue admin (remove/reorder) + mọi member actions |
| SaaS operator | Vận hành managed cloud sync server | Không đổi so với v1 — infra phục vụ mọi client OS |
| Anonymous visitor | User chưa trong phòng | Login nickname, create/join bằng room slug |

## Assumptions

1. **Giao thức và server** không đổi — macOS là client mới; cloud sync server không cần breaking change.
2. macOS user cài **dependencies ngoài** (yt-dlp, ffmpeg, audio player tương đương mpv) qua Homebrew hoặc manual — pattern tương tự Ubuntu; chi tiết deps ở architecture phase.
3. **Minimum macOS version** sẽ được xác định ở architecture (mặc định giả định: 2 phiên bản macOS gần nhất tại thời điểm release, trừ khi spec/architecture ghi khác).
4. Không notarize → README phải hướng dẫn rõ **Gatekeeper bypass** (`xattr -cr` hoặc System Settings) — đây là trade-off chấp nhận được cho OSS dev tool.
5. **Debian-based** = distros dùng apt và tương thích dependency stack hiện có (Debian, Linux Mint, Pop!_OS, v.v.); không đảm bảo mọi derivative.
6. Cross-platform testing dùng **ít nhất một cặp thật**: macOS host + Linux guest và ngược lại, trên cùng network có thể reach SaaS.
7. Feature set và permission model (democratic playback, host queue admin) **giữ nguyên** từ `001-terminal-music-room` clarify — không redesign.
8. `002-room-host-sci-fi-tui` TUI là baseline UI trên macOS — full parity nghĩa là TUI sci-fi chạy được trên macOS Terminal.

## Risks & constraints

| Risk | Impact | Mitigation |
|------|--------|------------|
| macOS audio stack khác Linux (CoreAudio vs Pulse/PipeWire) | Playback drift, install friction | Architecture chọn player path rõ; document brew deps; test sync trên macOS host/guest |
| Gatekeeper chặn unsigned binary | User không chạy được app | README + release notes hướng dẫn xattr; chấp nhận friction (no notarize) |
| Hai arch (arm64/amd64) tăng CI/release complexity | Build fail, artifact nhầm arch | Matrix build trong CI; đặt tên artifact rõ `darwin-arm64` / `darwin-amd64` |
| Homebrew formula maintenance | Stale formula, broken install | Pin version trong tap; test `brew install` trong review |
| Cross-platform charset/terminal rendering (TUI) | Layout vỡ trên Terminal.app vs iTerm | Test trên Terminal.app mặc định; document terminal khuyến nghị nếu cần |
| yt-dlp/ffmpeg trên macOS vs Linux version skew | Extraction fail một phía | Document minimum versions; align với Linux docs |
| Debian family variance (audio deps) | Install fail trên một số derivative | Target Debian + Ubuntu explicitly; graceful error khi thiếu dep |

## Gate G1 checklist

- [x] No blocking open questions
- [x] Scope bounded (in/out explicit)
- [x] Actors identified
- [x] Assumptions listed
