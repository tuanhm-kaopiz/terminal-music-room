# Spec: Terminal Music Room

**Slug:** `terminal-music-room`
**Status:** approved
**Gate G2:** ✅ pass

## Overview

Terminal Music Room là công cụ terminal (CLI + TUI đơn giản) cho phép 2–20 người dùng Ubuntu tham gia cùng một room trên dịch vụ cloud được quản lý, nghe nhạc YouTube đồng bộ theo thời gian thực, quản lý queue, chat text, vote skip/priority và phản ứng emoji — mà không cần rời terminal.

Trạng thái phát nhạc do **server là nguồn sự thật** (server-authoritative): mọi client trong room tuân theo cùng bài, vị trí phát và trạng thái play/pause do server điều phối.

## User stories / requirements

### REQ-001: Đăng nhập bằng nickname ẩn danh

**As a** người dùng mới  
**I want** đặt nickname trước khi vào room  
**So that** team nhận diện tôi trong chat và danh sách thành viên mà không cần tài khoản OAuth

**Acceptance criteria:**

- [ ] AC-001: Given người dùng chưa có nickname phiên hiện tại, When chạy lệnh đăng nhập với nickname hợp lệ (1–32 ký tự, không rỗng), Then nickname được lưu cho phiên và hiển thị trong client
- [ ] AC-002: Given nickname rỗng hoặc vượt quá 32 ký tự, When người dùng đăng nhập, Then hiển thị thông báo lỗi rõ ràng và không tạo phiên
- [ ] AC-003: Given người dùng đã đăng nhập, When thực hiện thao tác room/chat/playback, Then không bắt buộc đăng nhập lại trong cùng phiên client

### REQ-002: Tạo room

**As a** người dùng đã đăng nhập  
**I want** tạo room bằng slug toàn cục  
**So that** đồng nghiệp có thể join bằng tên room dễ nhớ

**Acceptance criteria:**

- [ ] AC-004: Given người dùng đã đăng nhập và slug chưa tồn tại, When tạo room với slug hợp lệ, Then room được tạo, người dùng trở thành host và tự động tham gia room
- [ ] AC-005: Given slug đã được room khác sử dụng, When người dùng tạo room cùng slug, Then từ chối với thông báo slug không khả dụng
- [ ] AC-006: Given slug không hợp lệ (rỗng, ký tự không cho phép), When tạo room, Then từ chối với thông báo lỗi validation

### REQ-003: Tham gia room

**As a** người dùng đã đăng nhập  
**I want** join room bằng slug  
**So that** tôi nghe nhạc cùng team

**Acceptance criteria:**

- [ ] AC-007: Given room tồn tại và có ít hơn 20 thành viên online, When join bằng slug đúng, Then người dùng vào room trong vòng 2 giây (đo từ lệnh join đến nhận trạng thái room ban đầu)
- [ ] AC-008: Given room không tồn tại, When join bằng slug, Then hiển thị lỗi room không tìm thấy
- [ ] AC-009: Given room đã đủ 20 thành viên online, When người thứ 21 join, Then từ chối với thông báo room đầy
- [ ] AC-010: Given người dùng join thành công, When vào room, Then nhận trạng thái hiện tại: bài đang phát (nếu có), vị trí playback, queue, danh sách online, tin chat gần đây

### REQ-004: Rời room

**As a** thành viên room  
**I want** leave room bất cứ lúc nào  
**So that** tôi ngừng nhận sync và chat khi không còn tham gia

**Acceptance criteria:**

- [ ] AC-011: Given người dùng đang trong room, When leave room, Then người dùng bị gỡ khỏi danh sách online và ngừng phát audio của room
- [ ] AC-012: Given thành viên leave, When còn thành viên khác online, Then room và playback tiếp tục cho những người còn lại
- [ ] AC-013: Given host leave và còn ít nhất một thành viên, When host rời room, Then quyền host chuyển cho thành viên join sớm nhất còn online (sau host)
- [ ] AC-014: Given host là thành viên cuối cùng, When host leave, Then room kết thúc và slug được giải phóng để tạo mới

### REQ-005: Danh sách thành viên online

**As a** thành viên room  
**I want** xem ai đang online  
**So that** tôi biết team nào đang nghe cùng

**Acceptance criteria:**

- [ ] AC-015: Given có thay đổi join/leave, When xem danh sách thành viên, Then danh sách phản ánh nickname và đánh dấu host trong vòng 500ms sau sự kiện
- [ ] AC-016: Given hai người dùng chọn cùng nickname ở hai phiên khác nhau trong cùng room, When cả hai online, Then server cho phép nhưng hiển thị phân biệt (ví dụ hậu tố) để tránh nhầm lẫn trong chat và danh sách

### REQ-006: Phát nhạc YouTube

**As a** thành viên room  
**I want** phát nhạc từ YouTube bằng tìm kiếm hoặc URL  
**So that** cả room nghe cùng nguồn nhạc

**Acceptance criteria:**

- [ ] AC-017: Given URL YouTube hợp lệ, When thành viên ra lệnh phát, Then bài được load và trở thành bài hiện tại cho toàn room
- [ ] AC-018: Given từ khóa tìm kiếm, When thành viên ra lệnh phát với keyword, Then hiển thị kết quả tìm kiếm (tối thiểu 1 mục) và phát mục được chọn (mặc định: kết quả đầu tiên nếu user không chọn)
- [ ] AC-019: Given URL không hợp lệ hoặc không phải YouTube, When ra lệnh phát, Then hiển thị lỗi nguồn không hợp lệ và không thay đổi bài đang phát
- [ ] AC-020: Given nguồn YouTube không phát được (bị chặn, không tồn tại, lỗi tải), When hệ thống thử phát, Then hiển thị lỗi rõ ràng cho room và giữ trạng thái phát ổn định (không treo room)

### REQ-007: Đồng bộ playback

**As a** thành viên room  
**I want** mọi người nghe cùng bài, cùng vị trí, cùng play/pause  
**So that** trải nghiệm “nghe chung” thật sự

**Acceptance criteria:**

- [ ] AC-021: Given ≥2 thành viên online và bài đang phát, When đo drift giữa client và server trong điều kiện mạng ổn định, Then drift trung bình ≤ 500ms (mục tiêu ≤ 200ms)
- [ ] AC-022: Given bài đang phát, When một thành viên pause, Then tất cả client chuyển sang paused trong vòng 500ms
- [ ] AC-023: Given bài đang pause, When một thành viên resume/play, Then tất cả client resume trong vòng 500ms
- [ ] AC-024: Given bài đang phát, When một thành viên skip, Then chuyển sang bài kế trong queue (hoặc kết thúc nếu queue rỗng) đồng bộ cho cả room trong vòng 500ms
- [ ] AC-025: Given bài đang phát, When một thành viên seek đến vị trí hợp lệ, Then tất cả client cập nhật vị trí tương ứng trong vòng 500ms
- [ ] AC-026: Given client hiển thị trạng thái phát, When có thay đổi từ server, Then hiển thị: tên bài, vị trí hiện tại / tổng thời lượng, trạng thái (playing, paused, buffering, ended)

### REQ-008: Queue — thêm và xem

**As a** thành viên room  
**I want** thêm bài vào queue và xem danh sách sắp phát  
**So that** team dựng playlist làm việc chung

**Acceptance criteria:**

- [ ] AC-027: Given URL hoặc keyword YouTube hợp lệ, When thành viên thêm vào queue, Then mục xuất hiện cuối queue với metadata: tiêu đề, thời lượng (nếu có), người thêm, thời điểm thêm
- [ ] AC-028: Given queue có nhiều mục, When xem queue, Then thứ tự phát hiển thị đúng và cập nhật khi có thêm/xóa/skip
- [ ] AC-029: Given bài hiện tại kết thúc và queue còn mục, When playback ended, Then tự động phát mục kế tiếp đồng bộ cho cả room

### REQ-009: Queue — quản trị host

**As a** host  
**I want** xóa và sắp xếp lại queue  
**So that** tôi kiểm soát playlist của room

**Acceptance criteria:**

- [ ] AC-030: Given người dùng là host, When xóa mục khỏi queue, Then mục biến mất khỏi queue của toàn room
- [ ] AC-031: Given người dùng là host, When đổi thứ tự queue, Then thứ tự mới áp dụng cho toàn room
- [ ] AC-032: Given người dùng không phải host, When cố xóa hoặc reorder queue, Then từ chối với thông báo không đủ quyền

### REQ-010: Chat text trong room

**As a** thành viên room  
**I want** gửi tin nhắn text và emoji trong room  
**So that** tôi tương tác với team mà không rời terminal

**Acceptance criteria:**

- [ ] AC-033: Given thành viên trong room, When gửi tin nhắn text, Then tin hiển thị cho tất cả thành viên với nickname người gửi và timestamp
- [ ] AC-034: Given tin nhắn chứa emoji Unicode, When gửi chat, Then emoji hiển thị đúng trên client hỗ trợ terminal
- [ ] AC-035: Given sự kiện hệ thống (join, leave, đổi bài, bắt đầu vote), When sự kiện xảy ra, Then tin system message xuất hiện trong luồng chat
- [ ] AC-036: Given tin nhắn rỗng hoặc chỉ khoảng trắng, When gửi chat, Then từ chối và không broadcast

### REQ-011: Vote skip bài hiện tại

**As a** thành viên room  
**I want** vote skip bài đang phát  
**So that** đa số có thể chuyển bài khi không ai muốn nghe

**Acceptance criteria:**

- [ ] AC-037: Given bài đang phát, When thành viên vote skip, Then vote được ghi nhận và hiển thị tiến độ (số vote / ngưỡng)
- [ ] AC-038: Given N thành viên online tại thời điểm bắt đầu vote, When số vote skip > 50% của N (ví dụ N=5 cần ≥3), Then skip bài hiện tại đồng bộ như REQ-007 skip
- [ ] AC-039: Given một thành viên đã vote skip, When vote lại cùng phiên vote, Then không tính trùng vote
- [ ] AC-040: Given vote chưa đạt ngưỡng, When hết thời gian vote (nếu có timeout) hoặc bài tự chuyển, Then vote kết thúc mà không skip

### REQ-012: Vote ưu tiên bài kế trong queue

**As a** thành viên room  
**I want** vote đưa một bài trong queue lên ưu tiên  
**So that** đa số chọn bài kế tiếp muốn nghe

**Acceptance criteria:**

- [ ] AC-041: Given queue có ≥2 mục chưa phát, When thành viên bắt đầu vote priority cho một mục cụ thể, Then thông báo vote bắt đầu (system message) và hiển thị mục được đề cử
- [ ] AC-042: Given N thành viên online tại thời điểm bắt đầu vote, When số vote priority > 50% của N, Then mục được vote chuyển lên vị trí kế tiếp sẽ phát (ngay sau bài hiện tại)
- [ ] AC-043: Given người dùng không phải host, When vote priority thành công, Then reorder có hiệu lực mà không cần host xác nhận thêm
- [ ] AC-044: Given mục không còn trong queue, When vote đang diễn ra, Then vote bị hủy với thông báo phù hợp

### REQ-013: Phản ứng emoji lên bài đang phát

**As a** thành viên room  
**I want** gửi emoji reaction lên bài hiện tại  
**So that** tôi phản hồi nhanh không cần chat dài

**Acceptance criteria:**

- [ ] AC-045: Given bài đang phát, When thành viên gửi reaction emoji hợp lệ, Then reaction hiển thị cho room gắn với bài hiện tại (tổng hợp theo loại emoji hoặc danh sách gần đây)
- [ ] AC-046: Given bài chuyển sang bài mới, When skip hoặc queue advance, Then reaction reset cho bài mới
- [ ] AC-047: Given không có bài đang phát, When gửi reaction, Then từ chối với thông báo không có bài để react

### REQ-014: Tự phục hồi khi mất kết nối

**As a** thành viên room  
**I want** client tự reconnect và đồng bộ lại  
**So that** mất mạng tạm thời không phá vỡ session

**Acceptance criteria:**

- [ ] AC-048: Given client đang trong room và mất kết nối mạng tạm thời, When kết nối trở lại trong vòng 5 phút, Then client tự reconnect, khôi phục trạng thái room và đồng bộ lại vị trí playback
- [ ] AC-049: Given reconnect thành công, When sync lại, Then drift sau resync ≤ 500ms trong vòng 3 giây
- [ ] AC-050: Given mất kết nối quá 5 phút, When client reconnect, Then yêu cầu join lại room thủ công (không giả định còn trong room)

### REQ-015: Chế độ CLI

**As a** người dùng quen terminal  
**I want** thao tác bằng lệnh text  
**So that** tôi không bắt buộc dùng giao diện đồ họa terminal

**Acceptance criteria:**

- [ ] AC-051: Given người dùng ở chế độ CLI, When thực hiện các lệnh room/playback/queue/chat/vote/reaction tương đương, Then kết quả hành vi khớp với AC của từng REQ tương ứng
- [ ] AC-052: Given lệnh không hợp lệ hoặc sai cú pháp, When người dùng nhập lệnh, Then hiển thị gợi ý lỗi ngắn gọn mà không thoát ứng dụng

### REQ-016: Chế độ TUI

**As a** người dùng muốn xem tổng quan  
**I want** giao diện TUI đơn giản hiển thị room, bài, queue, thành viên, chat  
**So that** tôi theo dõi trạng thái mà không gõ lệnh liên tục

**Acceptance criteria:**

- [ ] AC-053: Given người dùng mở TUI trong room, When xem màn hình chính, Then hiển thị đồng thời: tên room, bài & thời gian phát, danh sách online, queue, khung chat gần đây
- [ ] AC-054: Given trạng thái room thay đổi, When cập nhật từ server, Then TUI refresh các vùng tương ứng trong vòng 1 giây
- [ ] AC-055: Given người dùng chuyển giữa CLI và TUI (nếu hỗ trợ trong cùng phiên), When chuyển chế độ, Then không mất kết nối room và trạng thái phát

## Functional requirements

| ID | Requirement | Priority | Trace |
|----|-------------|----------|-------|
| FR-001 | Đăng nhập nickname ẩn danh, validation độ dài | Must | REQ-001 |
| FR-002 | Tạo room với slug toàn cục unique | Must | REQ-002 |
| FR-003 | Join room bằng slug, giới hạn 20 thành viên | Must | REQ-003 |
| FR-004 | Leave room; chuyển host khi host rời | Must | REQ-004 |
| FR-005 | Danh sách thành viên online real-time | Must | REQ-005 |
| FR-006 | Phát YouTube qua URL và tìm kiếm keyword | Must | REQ-006 |
| FR-007 | Server-authoritative sync: play/pause/skip/seek | Must | REQ-007 |
| FR-008 | Queue: thêm, xem, auto-advance | Must | REQ-008 |
| FR-009 | Queue: host xóa và reorder | Must | REQ-009 |
| FR-010 | Chat text, emoji, system messages | Must | REQ-010 |
| FR-011 | Vote skip >50% online members | Must | REQ-011 |
| FR-012 | Vote priority >50% online members | Must | REQ-012 |
| FR-013 | Emoji reactions trên bài hiện tại | Must | REQ-013 |
| FR-014 | Auto-reconnect và resync playback | Must | REQ-014 |
| FR-015 | Chế độ CLI đầy đủ chức năng | Must | REQ-015 |
| FR-016 | Chế độ TUI layout đơn giản | Must | REQ-016 |

## Edge cases & error scenarios

| Scenario | Expected behavior |
|----------|-------------------|
| Nickname rỗng / quá dài | Từ chối đăng nhập, thông báo validation (AC-002) |
| Slug room trùng | Từ chối tạo room (AC-005) |
| Join room không tồn tại | Lỗi not found (AC-008) |
| Room đầy (21 người) | Từ chối join (AC-009) |
| Trùng nickname trong room | Cho phép nhưng phân biệt hiển thị (AC-016) |
| URL YouTube sai / không phải YouTube | Lỗi nguồn không hợp lệ (AC-019) |
| YouTube không phát được | Lỗi cho room, không treo playback (AC-020) |
| Tìm kiếm không có kết quả | Thông báo không tìm thấy, không đổi bài hiện tại |
| Non-host xóa/reorder queue | Từ chối quyền (AC-032) |
| Chat rỗng | Không gửi (AC-036) |
| Vote skip chưa đủ >50% | Không skip; có thể timeout vote |
| Vote priority mục đã mất | Hủy vote (AC-044) |
| Reaction khi không có bài | Từ chối (AC-047) |
| Host leave, còn thành viên | Chuyển host cho member join sớm nhất (AC-013) |
| Host leave, room trống | Giải phóng slug (AC-014) |
| Mất mạng tạm thời | Auto-reconnect ≤5 phút, resync (AC-048, AC-049) |
| Mất mạng >5 phút | Yêu cầu join lại thủ công (AC-050) |
| Lệnh playback khi đang buffering | Server xếp hàng hoặc từ chối tạm thời với thông báo; không desync |
| Rate limit / abuse | Từ chối thao tác với thông báo thử lại sau (không crash client) |

## Non-functional requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-001 | Performance | Join room hoàn tất ≤ 2 giây (AC-007) |
| NFR-002 | Performance | Broadcast play/pause/skip/seek ≤ 500ms tới mọi client (AC-022–025) |
| NFR-003 | Performance | Playback drift ≤ 500ms acceptable, mục tiêu ≤ 200ms (AC-021) |
| NFR-004 | Performance | Client idle RAM < 300MB trên Ubuntu reference hardware |
| NFR-005 | Scalability | Hỗ trợ 2–20 thành viên đồng thời mỗi room |
| NFR-006 | Reliability | Reconnect tự động trong 5 phút; resync drift ≤ 500ms trong 3 giây (AC-048, AC-049) |
| NFR-007 | Security | Rate limiting và abuse protection cho tạo room, chat, vote — không yêu cầu auth OAuth v1 |
| NFR-008 | Platform | Client v1: Ubuntu Linux only |
| NFR-009 | Distribution | Phát hành open-source (chi tiết license ở architecture) |
| NFR-010 | Usability | TUI đơn giản đủ wireframe; không yêu cầu theme/animation để ship |

## Dependencies

- External APIs: YouTube (nguồn nhạc duy nhất v1)
- Internal modules: _(defer to architecture — sync server, room service, playback coordinator, client CLI/TUI)_
- Data migrations: none (v1 greenfield)

## Traceability matrix (AC → REQ)

| AC | REQ | AC | REQ | AC | REQ |
|----|-----|----|-----|----|-----|
| AC-001–003 | REQ-001 | AC-018–020 | REQ-006 | AC-035–036 | REQ-010 |
| AC-004–006 | REQ-002 | AC-021–026 | REQ-007 | AC-037–040 | REQ-011 |
| AC-007–010 | REQ-003 | AC-027–029 | REQ-008 | AC-041–044 | REQ-012 |
| AC-011–014 | REQ-004 | AC-030–032 | REQ-009 | AC-045–047 | REQ-013 |
| AC-015–016 | REQ-005 | AC-033–034 | REQ-010 | AC-048–050 | REQ-014 |
| AC-017 | REQ-006 | | | AC-051–052 | REQ-015 |
| | | | | AC-053–055 | REQ-016 |

## Gate G2 checklist

- [x] All requirements have testable AC
- [x] Edge cases documented
- [x] No implementation/tech details (defer to architecture)
- [x] Traceability: AC → REQ mapping complete
