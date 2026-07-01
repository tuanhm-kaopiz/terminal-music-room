# Spec: Room Host Sci-Fi TUI (v2)

**Slug:** `room-host-sci-fi-tui`
**Status:** draft
**Gate G2:** ✅ pass

## Overview

Thay **TUI đơn giản v1** bằng **một TUI sci-fi thống nhất** cho mọi người trong room (host và member). Giao diện theo hướng **cyberpunk / neon tối** — nền tối, accent neon, layout gợi HUD điều khiển — vẫn hoạt động trên terminal Ubuntu tiêu chuẩn với **tối thiểu 16 màu** và kích thước **80×24**.

Hành vi nghiệp vụ (sync playback, queue, chat, vote, reactions, phân quyền host/member) **giữ nguyên v1**; feature này chỉ định nghĩa **trải nghiệm TUI v2** và **parity điều khiển** so với CLI v1. CLI v1 vẫn tồn tại song song.

**Tham chiếu hành vi v1:** `docs/vibe/001-terminal-music-room/spec.md` — khi không mâu thuẫn, clarify.md của feature này được ưu tiên cho phạm vi UI.

## User stories / requirements

### REQ-001: Vào room qua TUI sci-fi thống nhất

**As a** thành viên đã join room  
**I want** mở TUI sci-fi khi vào room (join / tui)  
**So that** tôi có trải nghiệm terminal hiện đại thay layout đơn giản v1

**Acceptance criteria:**

- [ ] AC-001: Given người dùng đã đăng nhập nickname và join room thành công, When mở TUI (join kèm TUI hoặc lệnh tui tương đương), Then hiển thị **TUI sci-fi mới** thay TUI đơn giản v1
- [ ] AC-002: Given host và member trong cùng room, When cả hai mở TUI, Then dùng **cùng shell/layout sci-fi**; khác biệt chỉ ở quyền thao tác (REQ-010)
- [ ] AC-003: Given người dùng chưa join room, When cố mở TUI room, Then không hiển thị dashboard room; hướng dẫn join hoặc báo lỗi rõ ràng
- [ ] AC-004: Given người dùng thoát TUI (quit/escape theo UX định nghĩa), When thoát, Then kết nối room và playback theo quy tắc v1 (không tự leave room trừ khi user chủ động leave)

### REQ-002: Thẩm mỹ cyberpunk trên terminal 16 màu

**As a** người dùng terminal  
**I want** giao diện có cảm giác cyberpunk / sci-fi trending  
**So that** công cụ có nhận diện thị giác mạnh mà vẫn đọc được

**Acceptance criteria:**

- [ ] AC-005: Given terminal hỗ trợ ít nhất 16 màu trên nền tối, When xem TUI, Then palette dùng **nền tối** và **ít nhất hai màu accent neon** (vd. cyan, magenta, yellow) phân biệt được vùng HUD, tiêu đề, và nội dung
- [ ] AC-006: Given layout mặc định, When quan sát không cần hướng dẫn, Then ít nhất 3 người dùng thử nội bộ mô tả giao diện bằng ít nhất một trong các từ: *cyberpunk*, *sci-fi*, *futuristic*, *neon* — không cần giải thích thêm
- [ ] AC-007: Given terminal chỉ hỗ trợ 16 màu (không truecolor), When mở TUI, Then giao diện vẫn usable: chữ đọc được, panel phân tách rõ, không phụ thuộc gradient 24-bit
- [ ] AC-008: Given thao tác người dùng (chọn mục, gửi lệnh, focus panel), When có phản hồi, Then chỉ **micro-feedback** (highlight, đổi màu accent, trạng thái focus) — không có animation toàn màn hình hoặc hiệu ứng lặp liên tục gây lag cảm nhận được

### REQ-003: Dashboard mặc định — thông tin nhìn thấy ngay

**As a** host hoặc member  
**I want** thấy trạng thái room quan trọng ngay màn hình đầu  
**So that** tôi không phải đào sâu menu/tab để biết phòng đang như thế nào

**Acceptance criteria:**

- [ ] AC-009: Given người dùng trong room và terminal ≥ 80×24, When mở TUI lần đầu, Then **không cần tab ẩn** để thấy đồng thời: tên room, **bài đang phát** (hoặc trạng thái không có bài), **tiến độ phát** (vị trí / tổng thời lượng hoặc trạng thái tương đương), **ít nhất 3 mục queue kế tiếp** (hoặc toàn bộ queue nếu ≤3), **số người online**
- [ ] AC-010: Given terminal rộng/h cao hơn 80×24, When xem TUI, Then layout tận dụng không gian thêm mà không ẩn các thông tin AC-009
- [ ] AC-011: Given terminal nhỏ hơn 80×24, When mở TUI, Then hiển thị cảnh báo kích thước không đủ hoặc layout degraded có nhãn rõ; ưu tiên giữ readable: tên room + bài đang phát + số người online

### REQ-004: Điều khiển playback trong TUI

**As a** thành viên room  
**I want** play, pause, skip, seek từ TUI  
**So that** tôi điều phối nhạc không cần CLI

**Acceptance criteria:**

- [ ] AC-012: Given bài đang phát, When thành viên (host hoặc member) thực hiện pause từ TUI, Then trạng thái paused phản ánh trên TUI và đồng bộ room theo quy tắc v1 (≤ 500ms broadcast)
- [ ] AC-013: Given bài đang pause, When thành viên resume/play từ TUI, Then playback resume đồng bộ room theo quy tắc v1
- [ ] AC-014: Given bài đang phát, When thành viên skip từ TUI, Then chuyển bài theo queue/rules v1; TUI cập nhật bài và tiến độ mới
- [ ] AC-015: Given bài đang phát, When thành viên seek đến vị trí hợp lệ từ TUI, Then vị trí phát cập nhật đồng bộ room; thanh/tiến độ trên TUI phản ánh sau cập nhật server
- [ ] AC-016: Given không có bài đang phát, When thành viên thử skip/seek/pause, Then TUI hiển thị thông báo phù hợp; không crash hoặc treo TUI

### REQ-005: Thêm bài vào queue / phát từ TUI

**As a** thành viên room  
**I want** thêm bài bằng URL hoặc tìm kiếm từ TUI  
**So that** host và member cùng dựng playlist trong giao diện sci-fi

**Acceptance criteria:**

- [ ] AC-017: Given URL YouTube hợp lệ, When thành viên thêm/phát từ TUI, Then hành vi khớp v1: bài load và queue/now-playing cập nhật; TUI hiển thị metadata (tiêu đề tối thiểu)
- [ ] AC-018: Given từ khóa tìm kiếm, When thành viên search từ TUI, Then hiển thị danh sách kết quả (≥1 mục khi có kết quả) trong TUI; user chọn mục để thêm/phát
- [ ] AC-019: Given URL không hợp lệ hoặc search không có kết quả, When thao tác từ TUI, Then hiển thị lỗi trong TUI; không đổi bài đang phát (trừ khi user chọn kết quả hợp lệ)

### REQ-006: Quản trị queue (host-only) trong TUI

**As a** host  
**I want** xóa và sắp xếp lại queue từ TUI  
**So that** tôi đạt full parity với CLI host mà không thoát TUI

**Acceptance criteria:**

- [ ] AC-020: Given người dùng là host và queue có ≥1 mục, When xóa mục từ TUI, Then mục biến khỏi queue toàn room; danh sách queue trên TUI cập nhật
- [ ] AC-021: Given người dùng là host và queue có ≥2 mục, When reorder (đổi thứ tự) từ TUI, Then thứ tự mới áp dụng toàn room; TUI phản ánh thứ tự mới
- [ ] AC-022: Given người dùng **không** phải host, When cố xóa hoặc reorder queue từ TUI, Then thao tác bị từ chối với thông báo không đủ quyền; controls host-only **không hiển thị** hoặc **disabled** rõ ràng
- [ ] AC-023: Given host thực hiện luồng “thêm bài → xem queue → skip → xem vote” hoàn toàn trong TUI, When so sánh với CLI v1, Then số bước ≤ CLI hoặc cảm nhận nhanh hơn (đánh giá chủ quan ≥2/3 người thử nội bộ host)

### REQ-007: Danh sách thành viên trong TUI

**As a** thành viên room  
**I want** xem ai đang online và ai là host  
**So that** tôi biết bối cảnh phòng

**Acceptance criteria:**

- [ ] AC-024: Given có thay đổi join/leave, When xem panel members trên TUI, Then danh sách cập nhật trong vòng 1 giây với nickname và **đánh dấu host** rõ ràng
- [ ] AC-025: Given hai nickname trùng nhau trong room, When hiển thị trên TUI, Then phân biệt được (hậu tố hoặc ký hiệu) theo quy tắc v1

### REQ-008: Chat trong TUI

**As a** thành viên room  
**I want** đọc và gửi chat (text + emoji) trong TUI  
**So that** tôi tương tác team không cần CLI

**Acceptance criteria:**

- [ ] AC-026: Given thành viên trong room, When gửi tin nhắn text hợp lệ từ TUI, Then tin xuất hiện trong luồng chat TUI với nickname và timestamp; broadcast theo v1
- [ ] AC-027: Given tin nhắn rỗng hoặc chỉ khoảng trắng, When gửi từ TUI, Then từ chối; không broadcast
- [ ] AC-028: Given system message (join, leave, đổi bài, vote), When sự kiện xảy ra, Then hiển thị trong luồng chat TUI trong vòng 1 giây
- [ ] AC-029: Given chat dài hơn vùng hiển thị, When có tin mới, Then TUI cho phép xem lịch sử gần đây (scroll hoặc panel chat riêng); tin mới nhất luôn reachable

### REQ-009: Vote skip và vote priority trong TUI

**As a** thành viên room  
**I want** vote skip và vote priority từ TUI  
**So that** tôi tham gia quyết định cộng đồng trong cùng giao diện

**Acceptance criteria:**

- [ ] AC-030: Given bài đang phát, When thành viên vote skip từ TUI, Then hiển thị tiến độ vote (số vote / ngưỡng) trên TUI; đạt >50% online thì skip theo v1
- [ ] AC-031: Given queue có ≥2 mục chưa phát, When thành viên bắt đầu vote priority cho một mục từ TUI, Then TUI hiển thị mục được đề cử và tiến độ vote; đạt ngưỡng thì reorder theo v1
- [ ] AC-032: Given vote đang diễn ra, When trạng thái vote thay đổi, Then TUI cập nhật tiến độ trong vòng 1 giây
- [ ] AC-033: Given host xem phản hồi vote sau khi skip, When vote kết thúc, Then kết quả (đạt/không đạt) hiển thị trên TUI hoặc qua system message chat

### REQ-010: Phản ứng emoji trong TUI

**As a** thành viên room  
**I want** gửi reaction lên bài đang phát từ TUI  
**So that** tôi phản hồi nhanh trong session sci-fi

**Acceptance criteria:**

- [ ] AC-034: Given bài đang phát, When thành viên gửi reaction hợp lệ từ TUI, Then reaction hiển thị trên TUI (tổng hợp hoặc danh sách gần đây) gắn với bài hiện tại
- [ ] AC-035: Given bài chuyển sang bài mới, When skip hoặc queue advance, Then vùng reaction trên TUI reset cho bài mới
- [ ] AC-036: Given không có bài đang phát, When gửi reaction từ TUI, Then từ chối với thông báo rõ ràng

### REQ-011: Phân quyền host vs member trong cùng TUI

**As a** product owner  
**I want** một TUI sci-fi với quyền theo role  
**So that** member không thấy/không dùng được thao tác host-only

**Acceptance criteria:**

- [ ] AC-037: Given người dùng là **member** (không phải host), When duyệt TUI, Then có đủ: playback, thêm queue, chat, vote, reaction, xem members/queue — **không** có remove/reorder queue hoạt động được
- [ ] AC-038: Given người dùng là **host**, When duyệt TUI, Then có **toàn bộ** thao tác member **cộng** remove/reorder queue — tương đương parity CLI host v1
- [ ] AC-039: Given host rời room và quyền host chuyển theo v1, When TUI đang mở, Then UI cập nhật quyền trong vòng 1 giây (controls host-only ẩn/disable nếu không còn host)

### REQ-012: Cập nhật real-time và mất kết nối trong TUI

**As a** thành viên room  
**I want** TUI phản ánh trạng thái server và xử lý mất mạng  
**So that** tôi tin tưởng dashboard khi sync hoặc reconnect

**Acceptance criteria:**

- [ ] AC-040: Given trạng thái room thay đổi từ server (playback, queue, members, chat, vote), When client nhận cập nhật, Then các vùng TUI tương ứng refresh trong vòng **1 giây**
- [ ] AC-041: Given mất kết nối tạm thời, When TUI đang mở, Then hiển thị trạng thái disconnected/reconnecting rõ ràng (không im lặng)
- [ ] AC-042: Given reconnect thành công trong vòng 5 phút theo v1, When TUI sync lại, Then dashboard khôi phục đúng trạng thái room; drift playback theo NFR v1 sau resync
- [ ] AC-043: Given mất kết nối > 5 phút, When reconnect, Then TUI yêu cầu join lại thủ công với thông báo rõ ràng

### REQ-013: CLI v1 vẫn hoạt động song song

**As a** người dùng quen scripting  
**I want** tiếp tục dùng lệnh CLI v1  
**So that** TUI không khóa tôi vào một chế độ duy nhất

**Acceptance criteria:**

- [ ] AC-044: Given người dùng dùng lệnh CLI v1 cho room/playback/queue/chat/vote/reaction, When thực hiện thao tác, Then hành vi khớp spec v1; không bị regression do TUI v2
- [ ] AC-045: Given người dùng chuyển giữa CLI và TUI trong cùng phiên (nếu luồng v1 hỗ trợ), When chuyển chế độ, Then không mất kết nối room và trạng thái phát

### REQ-014: Rời room và thoát từ TUI

**As a** thành viên room  
**I want** leave room từ TUI  
**So that** tôi kết thúc session gọn trong giao diện sci-fi

**Acceptance criteria:**

- [ ] AC-046: Given người dùng trong room, When leave room từ TUI, Then rời room theo v1; TUI đóng hoặc chuyển về trạng thái ngoài room
- [ ] AC-047: Given host leave khi còn thành viên, When leave từ TUI, Then chuyển host theo v1; TUI của các client còn lại cập nhật đánh dấu host mới

## Functional requirements

| ID | Requirement | Priority | Trace |
|----|-------------|----------|-------|
| FR-001 | TUI sci-fi thống nhất thay TUI v1 khi join/tui | Must | REQ-001 |
| FR-002 | Palette cyberpunk, tối ưu 16 màu, micro-feedback only | Must | REQ-002 |
| FR-003 | Dashboard mặc định: now playing, tiến độ, queue, số online | Must | REQ-003 |
| FR-004 | Playback play/pause/skip/seek trong TUI | Must | REQ-004 |
| FR-005 | Thêm bài / search YouTube trong TUI | Must | REQ-005 |
| FR-006 | Host remove/reorder queue trong TUI | Must | REQ-006 |
| FR-007 | Panel members với đánh dấu host | Must | REQ-007 |
| FR-008 | Chat text, emoji, system messages trong TUI | Must | REQ-008 |
| FR-009 | Vote skip và vote priority trong TUI | Must | REQ-009 |
| FR-010 | Emoji reactions trong TUI | Must | REQ-010 |
| FR-011 | Role-based UI host vs member | Must | REQ-011 |
| FR-012 | Real-time refresh và reconnect UX trong TUI | Must | REQ-012 |
| FR-013 | CLI v1 không regression | Must | REQ-013 |
| FR-014 | Leave room từ TUI | Must | REQ-014 |

## Edge cases & error scenarios

| Scenario | Expected behavior |
|----------|-------------------|
| Terminal < 80×24 | Cảnh báo hoặc layout degraded; ưu tiên room name + now playing + online count (AC-011) |
| Terminal chỉ 16 màu | Giao diện usable, không yêu cầu truecolor (AC-007) |
| Terminal thiếu Unicode box-drawing | Layout fallback vẫn đọc được; borders không vỡ nội dung (chi tiết layout → architecture) |
| Member cố queue admin | Từ chối + controls disabled/ẩn (AC-022, AC-037) |
| Host mất quyền khi host chuyển | UI cập nhật quyền ≤1s (AC-039) |
| Không có bài đang phát | Skip/seek/pause/reaction báo lỗi phù hợp (AC-016, AC-036) |
| URL/search không hợp lệ | Lỗi trong TUI; không đổi now playing (AC-019) |
| Chat rỗng | Không gửi (AC-027) |
| Vote chưa đủ ngưỡng | Hiển thị tiến độ; không skip/reorder (AC-030, AC-031) |
| Mất mạng tạm | Indicator reconnecting; auto-resync ≤5 phút (AC-041, AC-042) |
| Mất mạng >5 phút | Yêu cầu join lại (AC-043) |
| Server rate limit / lỗi | Thông báo trong TUI; không crash hoặc treo vĩnh viễn |
| Resize terminal khi TUI mở | Layout thích ứng hoặc degraded có nhãn; không corrupt nội dung |
| Nhiều thành viên (đến 20) | Members list scroll được; performance theo NFR |

## Non-functional requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-001 | Usability | Terminal tối thiểu 80×24; 16 màu trên nền tối (clarify.md) |
| NFR-002 | Usability | ≥3 người thử mô tả được cyberpunk/sci-fi/futuristic/neon (AC-006) |
| NFR-003 | Performance | Cập nhật vùng TUI sau sự kiện server ≤ 1 giây (AC-040) |
| NFR-004 | Performance | Không animation nặng; micro-feedback only (AC-008) |
| NFR-005 | Performance | Giữ ngưỡng v1: broadcast playback ≤ 500ms; drift sync theo spec v1 |
| NFR-006 | Performance | Client RAM idle không tăng >20% so với TUI v1 trên cùng hardware reference |
| NFR-007 | Compatibility | Ubuntu terminal tiêu chuẩn (GNOME Terminal hoặc tương đương) |
| NFR-008 | Maintainability | Một visual direction cyberpunk; không multi-theme v2 |
| NFR-009 | Scope | Không thay đổi permission model, voting rules, hoặc sync protocol backend |

## Dependencies

- **Hành vi v1:** `docs/vibe/001-terminal-music-room/spec.md` — playback sync, queue rules, chat, vote, reactions, reconnect, host transfer
- **Clarify v2:** `docs/vibe/002-room-host-sci-fi-tui/clarify.md` — unified TUI, cyberpunk, 16-color, full host parity
- External APIs: không thêm (vẫn YouTube qua client v1)
- Data migrations: none

## Traceability matrix (AC → REQ)

| AC | REQ | AC | REQ | AC | REQ |
|----|-----|----|-----|----|-----|
| AC-001–004 | REQ-001 | AC-017–019 | REQ-005 | AC-030–033 | REQ-009 |
| AC-005–008 | REQ-002 | AC-020–023 | REQ-006 | AC-034–036 | REQ-010 |
| AC-009–011 | REQ-003 | AC-024–025 | REQ-007 | AC-037–039 | REQ-011 |
| AC-012–016 | REQ-004 | AC-026–029 | REQ-008 | AC-040–043 | REQ-012 |
| | | | | AC-044–045 | REQ-013 |
| | | | | AC-046–047 | REQ-014 |

## Gate G2 checklist

- [x] All requirements have testable AC
- [x] Edge cases documented
- [x] No implementation/tech details (defer to architecture)
- [x] Traceability: AC → REQ mapping complete
