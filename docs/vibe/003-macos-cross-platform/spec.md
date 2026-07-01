# Spec: macOS Cross-Platform Support (V0.2.1)

**Slug:** `macos-cross-platform`
**Status:** approved
**Gate G2:** ✅ pass

## Overview

Mở rộng Terminal Music Room sang **macOS** (release V0.2.1) với **full feature parity** so với client Linux hiện tại: CLI, TUI sci-fi, playback sync, queue, chat, vote, reactions. Người dùng macOS có thể **host hoặc join** phòng cùng người dùng **Ubuntu / Debian-based Linux** — hai chiều cross-platform đều phải hoạt động.

Phân phối qua **GitHub Release** (hai kiến trúc) và **Homebrew**. Không yêu cầu code signing/notarization; tài liệu hướng dẫn xử lý Gatekeeper là bắt buộc.

**Tham chiếu hành vi nghiệp vụ (parity):**

- `docs/vibe/001-terminal-music-room/spec.md` — toàn bộ REQ-001–REQ-016 và AC tương ứng áp dụng cho client macOS trừ khi spec này ghi đè.
- `docs/vibe/002-room-host-sci-fi-tui/spec.md` — TUI sci-fi áp dụng cho macOS Terminal thay TUI đơn giản v1.

**Tham chiếu scope:** `docs/vibe/003-macos-cross-platform/clarify.md`

## User stories / requirements

### REQ-001: Cài đặt client macOS từ GitHub Release

**As a** người dùng macOS  
**I want** tải và chạy client từ release chính thức  
**So that** tôi dùng Terminal Music Room mà không cần build từ source

**Acceptance criteria:**

- [ ] AC-001: Given release V0.2.1 đã publish, When người dùng tải artifact macOS cho **Apple Silicon**, Then có file/binary chạy được trên máy arm64 và lệnh `music-room --version` (hoặc tương đương) in ra phiên bản V0.2.1
- [ ] AC-002: Given release V0.2.1 đã publish, When người dùng tải artifact macOS cho **Intel**, Then có file/binary chạy được trên máy amd64 và lệnh version in ra V0.2.1
- [ ] AC-003: Given người dùng chạy binary trên kiến trúc **không khớp** artifact (vd. arm64 binary trên Intel), When thử chạy, Then không crash im lặng — hiển thị lỗi rõ ràng hoặc từ chối chạy với hướng dẫn chọn đúng artifact
- [ ] AC-004: Given dependencies phát nhạc bắt buộc chưa được cài, When người dùng chạy lệnh cần playback, Then hiển thị thông báo thiếu dependency và hướng dẫn cài (không crash không giải thích)

### REQ-002: Cài đặt client macOS qua Homebrew

**As a** người dùng macOS quen Homebrew  
**I want** cài client bằng một lệnh brew  
**So that** cài đặt và cập nhật thuận tiện như các CLI khác

**Acceptance criteria:**

- [ ] AC-005: Given Homebrew đã cài trên macOS arm64, When chạy lệnh cài từ tap/formula chính thức của project, Then `music-room` có trong PATH và `--version` trả về V0.2.1
- [ ] AC-006: Given Homebrew trên macOS Intel (amd64), When cài từ cùng tap/formula, Then `music-room` chạy được và version khớp V0.2.1
- [ ] AC-007: Given cài qua Homebrew thành công, When người dùng thực hiện login → create/join room, Then hành vi khớp AC của REQ-003 (parity v1)

### REQ-003: Feature parity — client macOS đầy đủ chức năng v1

**As a** người dùng macOS  
**I want** mọi chức năng room/playback/chat/vote/reaction như trên Linux  
**So that** tôi không bị thiếu tính năng khi chuyển sang Mac

**Acceptance criteria:**

- [ ] AC-008: Given client macOS đã cài đủ dependencies, When thực hiện lần lượt các luồng: đăng nhập nickname, tạo room, join room, leave room, Then mỗi luồng pass các AC tương ứng REQ-001–REQ-004 trong spec `001-terminal-music-room`
- [ ] AC-009: Given ≥2 thành viên trong room (cùng macOS hoặc mixed — xem REQ-005/006), When thực hiện phát YouTube, play/pause/skip/seek, Then pass AC REQ-006 và REQ-007 spec `001` (sync ≤ 500ms, drift ≤ 500ms)
- [ ] AC-010: Given thành viên và host trong room, When thêm/xem queue, host xóa/reorder, non-host bị từ chối quyền admin, Then pass AC REQ-008 và REQ-009 spec `001`
- [ ] AC-011: Given thành viên trong room, When chat, vote skip, vote priority, emoji reaction, Then pass AC REQ-010, REQ-011, REQ-012, REQ-013 spec `001`
- [ ] AC-012: Given client macOS mất mạng tạm thời, When reconnect trong 5 phút, Then pass AC REQ-014 spec `001` (resync drift ≤ 500ms trong 3 giây)
- [ ] AC-013: Given người dùng dùng chế độ CLI trên macOS, When thực hiện lệnh room/playback/queue/chat/vote/reaction, Then pass AC REQ-015 spec `001`

### REQ-004: TUI sci-fi trên macOS Terminal

**As a** người dùng macOS  
**I want** TUI sci-fi giống trải nghiệm Linux hiện tại  
**So that** giao diện terminal nhất quán cross-platform

**Acceptance criteria:**

- [ ] AC-014: Given người dùng macOS đã join room và mở TUI trên **Terminal.app** mặc định (≥ 80×24, 16 màu), When xem màn hình, Then pass AC REQ-001–REQ-003 spec `002-room-host-sci-fi-tui` (TUI sci-fi, cyberpunk palette, dashboard mặc định)
- [ ] AC-015: Given host và member trên macOS trong cùng room, When cả hai điều khiển playback, queue, chat, vote, reaction từ TUI, Then pass AC REQ-004–REQ-012 spec `002` và phân quyền host/member pass AC REQ-011 spec `002`
- [ ] AC-016: Given cập nhật trạng thái từ server, When xem TUI trên macOS, Then các vùng dashboard refresh trong vòng 1 giây (tương đương AC-054 spec `001` / AC REQ-012 spec `002`)
- [ ] AC-017: Given người dùng thoát TUI trên macOS, When quit/escape theo UX định nghĩa spec `002`, Then không tự leave room; kết nối room giữ theo quy tắc v1

### REQ-005: Cross-platform — host macOS, guest Linux

**As a** team mixed-OS  
**I want** host room từ Mac và bạn Linux join được  
**So that** Mac user có thể làm host cho nhóm

**Acceptance criteria:**

- [ ] AC-018: Given host trên macOS đã tạo room slug `cross-test-mac-host`, When guest trên **Ubuntu** join cùng slug, Then guest vào room trong ≤ 2 giây và nhận trạng thái ban đầu (now-playing, queue, online list) khớp host
- [ ] AC-019: Given host macOS + guest Ubuntu trong room, When host phát bài YouTube hợp lệ, Then guest nghe audio đồng bộ; drift trung bình ≤ 500ms sau ổn định 10 giây
- [ ] AC-020: Given host macOS + guest Ubuntu, When guest thực hiện pause, skip, thêm queue, chat, vote skip, reaction, Then mọi thành viên (kể cả host) thấy cập nhật đúng trong thời gian broadcast ≤ 500ms cho lệnh playback
- [ ] AC-021: Given host macOS + guest trên **Debian** (hoặc derivative Debian-based đã liệt kê trong docs), When join và phát nhạc, Then pass AC-018–AC-020 tương đương
- [ ] AC-022: Given host macOS là queue admin, When host xóa/reorder queue, Then guest Linux thấy queue cập nhật; khi guest không phải host cố xóa queue, Then bị từ chối quyền

### REQ-006: Cross-platform — host Linux, guest macOS

**As a** team mixed-OS  
**I want** host room từ Linux và bạn Mac join được  
**So that** Linux user vẫn host được khi có thành viên Mac

**Acceptance criteria:**

- [ ] AC-023: Given host trên **Ubuntu** đã tạo room, When guest macOS join, Then guest vào room trong ≤ 2 giây và trạng thái khớp host
- [ ] AC-024: Given host Ubuntu + guest macOS, When host phát bài và guest pause/skip/thêm queue/chat/vote, Then đồng bộ hai chiều pass tương tự AC-019–AC-020
- [ ] AC-025: Given host trên **Debian** + guest macOS, When join và phát nhạc, Then pass AC-023–AC-024 tương đương
- [ ] AC-026: Given guest macOS mở TUI sci-fi trong room host Linux, When có thay đổi playback từ host, Then TUI macOS cập nhật now-playing và tiến độ trong vòng 1 giây

### REQ-007: Mở rộng hỗ trợ Linux Debian-based

**As a** người dùng Debian / derivative  
**I want** client Linux hoạt động ổn định như trên Ubuntu  
**So that** tôi tham gia phòng cross-platform mà không bắt buộc Ubuntu

**Acceptance criteria:**

- [ ] AC-027: Given máy **Debian stable** (phiên bản ghi trong docs release), When cài client theo hướng dẫn Linux và đủ dependencies, Then login → join room → playback pass AC REQ-003 và REQ-007 spec `001`
- [ ] AC-028: Given docs cài đặt Linux, When đọc phần distro support, Then liệt kê rõ **Ubuntu** và **Debian-based** in-scope; ghi chú distro ngoài scope (Fedora, Arch, …) là không hỗ trợ V0.2.1
- [ ] AC-029: Given dependency bắt buộc thiếu trên Debian derivative, When chạy lệnh playback, Then thông báo lỗi rõ ràng tương tự AC-004 (không crash im lặng)

### REQ-008: Tài liệu cài đặt macOS và Gatekeeper

**As a** người dùng macOS lần đầu  
**I want** hướng dẫn cài và mở app khi macOS chặn unsigned binary  
**So that** tôi không bỏ cuộc vì Gatekeeper

**Acceptance criteria:**

- [ ] AC-030: Given README hoặc docs cài đặt V0.2.1, When người dùng macOS đọc, Then có section riêng **macOS** với: cách cài GitHub Release (chọn arm64 vs Intel), cách cài Homebrew, và danh sách dependencies cần cài trước khi phát nhạc
- [ ] AC-031: Given binary không notarize, When macOS Gatekeeper chặn lần chạy đầu, Then docs mô tả **ít nhất một** cách xử lý được (vd. mở từ System Settings, hoặc lệnh xóa quarantine attribute) — bước có thể reproduce theo docs
- [ ] AC-032: Given release notes V0.2.1, When đọc, Then ghi rõ: hỗ trợ macOS arm64 + amd64, không signed/notarized, và link tới section macOS trong README

### REQ-009: Release V0.2.1 — artifact đầy đủ

**As a** người dùng cuối  
**I want** release V0.2.1 có đủ bản Linux và macOS  
**So that** mọi nền tảng in-scope cập nhật cùng phiên bản

**Acceptance criteria:**

- [ ] AC-033: Given tag/release V0.2.1 trên kênh phân phối chính thức, When xem assets, Then có **ít nhất bốn** artifact client: Linux (kiến trúc hiện có), macOS arm64, macOS amd64 — tên file phân biệt rõ nền tảng và kiến trúc
- [ ] AC-034: Given cùng tag V0.2.1, When so sánh `--version` trên Linux và macOS, Then cùng số phiên bản semver V0.2.1
- [ ] AC-035: Given Homebrew formula/tap cho V0.2.1, When cài trên macOS, Then version khớp AC-034

## Functional requirements

| ID | Requirement | Priority | Trace |
|----|-------------|----------|-------|
| FR-001 | GitHub Release macOS arm64 + amd64 | Must | REQ-001 |
| FR-002 | Homebrew install macOS arm64 + amd64 | Must | REQ-002 |
| FR-003 | Full v1 feature parity trên macOS (CLI) | Must | REQ-003 |
| FR-004 | TUI sci-fi parity trên macOS Terminal | Must | REQ-004 |
| FR-005 | Cross-platform host macOS ↔ guest Linux | Must | REQ-005, REQ-006 |
| FR-006 | Debian-based Linux support mở rộng | Must | REQ-007 |
| FR-007 | Docs macOS + Gatekeeper workaround | Must | REQ-008 |
| FR-008 | Release V0.2.1 multi-platform artifacts | Must | REQ-009 |
| FR-009 | Kế thừa permission model v1 (democratic playback, host queue admin) | Must | REQ-005, REQ-006 |
| FR-010 | Không code signing / notarization V0.2.1 | Must | REQ-008 |

## Edge cases & error scenarios

| Scenario | Expected behavior |
|----------|-------------------|
| macOS user tải sai kiến trúc binary | Lỗi rõ ràng hoặc từ chối chạy; hướng dẫn chọn đúng artifact (AC-003) |
| Gatekeeper chặn unsigned binary | Docs hướng dẫn bypass; không yêu cầu notarize (AC-031) |
| Thiếu dependency phát nhạc trên macOS | Thông báo + hướng dẫn cài; không crash im lặng (AC-004) |
| Thiếu dependency trên Debian derivative | Thông báo rõ (AC-029) |
| Host macOS + guest Linux mất mạng một phía | Reconnect/resync theo REQ-014 spec `001`; room không treo cho phía còn lại |
| Cross-platform drift cao do latency | Client tự hiệu chỉnh; drift ≤ 500ms sau ổn định (AC-019, NFR-003) |
| TUI trên terminal macOS < 80×24 | Cảnh báo hoặc layout degraded theo spec `002` AC-011 |
| TUI emoji/Unicode trên Terminal.app | Chat emoji hiển thị đúng hoặc degraded có nhãn — không crash TUI |
| Guest macOS join room Linux host đầy (20 người) | Từ chối join theo AC-009 spec `001` |
| Host leave cross-platform room | Chuyển host theo AC-013 spec `001`; guest macOS/Linux thấy host mới |
| YouTube không phát được trên một OS | Lỗi hiển thị cho room; phía OS kia vẫn ổn định (AC-020 spec `001`) |
| Cố cài Homebrew formula trên Linux | Ngoài scope — formula chỉ dành macOS |
| Windows user cố tải macOS/Linux binary | Ngoài scope V0.2.1 — không artifact Windows |

## Non-functional requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-001 | Performance | Join room ≤ 2 giây trên macOS client (AC-018, AC-023) |
| NFR-002 | Performance | Broadcast playback commands ≤ 500ms cross-platform (AC-020, AC-024) |
| NFR-003 | Performance | Playback drift ≤ 500ms cross-platform sau ổn định (AC-019) |
| NFR-004 | Performance | Client macOS idle RAM < 300MB (cùng mức mục tiêu v1) |
| NFR-005 | Platform | macOS: arm64 + amd64; Linux: Ubuntu + Debian-based (AC-027, AC-028) |
| NFR-006 | Distribution | GitHub Release + Homebrew cho macOS V0.2.1 |
| NFR-007 | Security | Không notarize — user chấp nhận Gatekeeper friction; docs bắt buộc |
| NFR-008 | Compatibility | TUI usable trên Terminal.app mặc định, 16 màu, tối thiểu 80×24 |
| NFR-009 | Reliability | Reconnect macOS client ≤ 5 phút; resync theo spec `001` REQ-014 |
| NFR-010 | Out of scope | Windows, App Store, notarization, Fedora/Arch, tính năng mới ngoài port |

## Dependencies

- External APIs: YouTube (nguồn nhạc — kế thừa v1, không đổi)
- Managed cloud sync service: kế thừa v1 (server-authoritative, không đổi giao thức)
- Internal modules: _(defer to architecture — macOS client build, Homebrew tap, cross-platform audio, release pipeline)_
- Data migrations: none
- Upstream specs: `001-terminal-music-room`, `002-room-host-sci-fi-tui`

## Traceability matrix (AC → REQ)

| AC | REQ | AC | REQ | AC | REQ |
|----|-----|----|-----|----|-----|
| AC-001–004 | REQ-001 | AC-013 | REQ-003 | AC-023–026 | REQ-006 |
| AC-005–007 | REQ-002 | AC-014–017 | REQ-004 | AC-027–029 | REQ-007 |
| AC-008–012 | REQ-003 | AC-018–022 | REQ-005 | AC-030–032 | REQ-008 |
| | | | | AC-033–035 | REQ-009 |

### Parity inheritance (macOS → upstream specs)

| macOS REQ | Upstream coverage |
|-----------|-------------------|
| REQ-003 | `001` REQ-001–REQ-016 (AC-008–AC-013 verify pass) |
| REQ-004 | `002` REQ-001–REQ-014 (AC-014–AC-017 verify pass) |
| REQ-005, REQ-006 | `001` REQ-006–REQ-014 behaviors trong context cross-platform |

## Gate G2 checklist

- [x] All requirements have testable AC
- [x] Edge cases documented
- [x] No implementation/tech details (defer to architecture)
- [x] Traceability: AC → REQ mapping complete
