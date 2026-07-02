# Spec: Room Password & Host Kick

**Slug:** `room-password-host-kick`
**Status:** approved
**Gate G2:** ✅ pass

## Overview

Bổ sung **kiểm soát truy cập room** bằng password tùy chọn khi tạo room, và **quyền kick** cho host để đuổi member khỏi session. Joiner phải cung cấp password đúng trước khi vào room có password; host kick member từ panel Members trong TUI. Người bị kick nhận thông báo rõ và có thể join lại ngay nếu biết room ID + password.

Room **không password** giữ hành vi join mở như hiện tại. Không có tài khoản user, ban vĩnh viễn, hay đổi password sau khi tạo room.

**Tham chiếu:** `docs/vibe/005-room-password-host-kick/clarify.md` (scope đã chốt); `docs/vibe/002-room-host-sci-fi-tui/spec.md` (panel Members, TUI baseline).

## User stories / requirements

### REQ-001: Tạo room với password tùy chọn

**As a** room host  
**I want** đặt password khi tạo room (hoặc bỏ trống để room mở)  
**So that** tôi kiểm soát ai được vào session riêng tư

**Acceptance criteria:**

- [ ] AC-001: Given host tạo room **không** cung cấp password (bỏ trống / không flag), When room được tạo thành công, Then room ở trạng thái **mở** — joiner chỉ cần room ID để vào, hành vi giống trước feature này
- [ ] AC-002: Given host tạo room với password hợp lệ (1–32 ký tự), When room được tạo thành công, Then room yêu cầu password khi join; host vào room **không** cần nhập lại password trong cùng phiên tạo
- [ ] AC-003: Given host tạo room qua CLI, When dùng tùy chọn password (vd. `--password`), Then password được áp dụng cho room theo AC-002
- [ ] AC-004: Given host tạo room qua TUI, When được prompt nhập password, Then ký tự nhập **bị ẩn** (masked); bỏ trống prompt = room mở theo AC-001
- [ ] AC-005: Given host nhập password **dài hơn 32 ký tự** hoặc **0 ký tự khi cố “có password”** (chỉ khoảng trắng), When submit, Then tạo room **thất bại** với thông báo lỗi validation rõ ràng

### REQ-002: Join room có password

**As a** joiner  
**I want** nhập password khi join room được bảo vệ  
**So that** chỉ người được host chia sẻ mật khẩu mới vào được

**Acceptance criteria:**

- [ ] AC-006: Given room **có** password và joiner biết room ID + password đúng, When join (CLI hoặc TUI), Then joiner **được admit** vào room và thấy TUI room bình thường
- [ ] AC-007: Given room **không** password, When joiner join chỉ với room ID, Then joiner **được admit** không cần password
- [ ] AC-008: Given room có password, When joiner join qua CLI với flag password đúng, Then admit thành công theo AC-006
- [ ] AC-009: Given room có password, When joiner join qua TUI và được prompt password, Then ký tự nhập **bị ẩn**; nhập đúng → admit theo AC-006
- [ ] AC-010: Given room có password, When joiner **không** cung cấp password (bỏ trống prompt / không flag), Then join **bị từ chối** với thông báo yêu cầu password

### REQ-003: Từ chối join sai password

**As a** joiner  
**I want** biết rõ khi password sai  
**So that** tôi có thể thử lại hoặc liên hệ host

**Acceptance criteria:**

- [ ] AC-011: Given room có password, When joiner cung cấp password **sai**, Then join **bị từ chối**; joiner **không** vào room và **không** thấy nội dung room (queue, members, playback)
- [ ] AC-012: Given join bị từ chối vì sai password, When xem thông báo lỗi, Then message **không tiết lộ** password đúng và **không** phân biệt “room không tồn tại” vs “sai password” nếu product chọn message thống nhất — nhưng phải rõ là **lỗi xác thực / không được phép vào**
- [ ] AC-013: Given joiner nhập sai password nhiều lần liên tiếp, When mỗi lần thử, Then mỗi lần đều bị từ chối; **không** có lockout vĩnh viễn hay cooldown bắt buộc trong v1 (chỉ reject từng lần)

### REQ-004: Host kick member từ panel Members

**As a** room host  
**I want** kick một member đã chọn trong panel Members  
**So that** tôi loại người gây rối khỏi session ngay lập tức

**Acceptance criteria:**

- [ ] AC-014: Given host đang trong room với ≥1 member (không tính host), When host focus panel Members, chọn một member, và nhấn phím kick (vd. `K` hoặc `Del`), Then member đó **bị ngắt kết nối** khỏi room trong vòng **3 giây**
- [ ] AC-015: Given member vừa bị kick, When các client còn lại xem panel Members, Then member đó **không còn** trong danh sách active
- [ ] AC-016: Given host kick member, When kick thành công, Then **không** có bước xác nhận bắt buộc trong v1 — kick thực thi ngay (theo clarify)
- [ ] AC-017: Given chỉ còn host trong room, When host mở panel Members, Then **không có** member nào để kick (hoặc action kick không khả dụng khi không có selection hợp lệ)

### REQ-005: Quyền kick chỉ dành cho host

**As a** room member thường  
**I want** không thấy hoặc không dùng được kick  
**So that** chỉ host quản lý membership

**Acceptance criteria:**

- [ ] AC-018: Given user là **member** (không phải host), When xem panel Members, Then **không** có action kick khả dụng (phím kick không có hiệu lực / không hiển thị hint kick cho member)
- [ ] AC-019: Given member cố gửi yêu cầu kick (nếu có đường bypass), When server xử lý, Then yêu cầu **bị từ chối** — member không kick được ai
- [ ] AC-020: Given host đang trong panel Members, When chọn **chính host** (nếu host xuất hiện trong list), Then action kick **không** áp dụng cho host / không kick được chính mình

### REQ-006: Trải nghiệm member bị kick và re-join

**As a** member bị kick  
**I want** biết mình bị host remove và có thể join lại nếu cần  
**So that** tôi hiểu chuyện gì xảy ra, không tưởng mất mạng

**Acceptance criteria:**

- [ ] AC-021: Given member bị host kick, When client nhận sự kiện kick, Then hiển thị thông báo rõ (vd. “Removed from room by host” hoặc tương đương tiếng Việt/Anh) — **không** silent disconnect
- [ ] AC-022: Given member bị kick, When kick hoàn tất, Then client **rời** TUI room (màn join / thoát session) và **không** tiếp tục nhận sync playback/chat của room đó
- [ ] AC-023: Given member vừa bị kick, When join lại cùng room ID với password đúng (nếu room có password), Then **được admit** lại — không có ban vĩnh viễn trong v1
- [ ] AC-024: Given member bị kick và join lại thành công, When vào room, Then xuất hiện trong panel Members như member mới; **không** giữ trạng thái vote/reaction cũ từ phiên trước (state session mới)

### REQ-007: Bảo mật hiển thị password (product level)

**As a** host hoặc joiner  
**I want** password không lộ trên UI công khai hoặc log hiển thị cho user  
**So that** mật khẩu room không bị lộ cho người khác trong cùng phòng

**Acceptance criteria:**

- [ ] AC-025: Given host đã set password cho room, When member khác (đã join) xem TUI/CLI, Then **không** hiển thị plaintext password của room ở bất kỳ panel nào (header, members, help)
- [ ] AC-026: Given user nhập password qua TUI prompt, When đang gõ, Then chỉ thấy ký tự masked (`*` hoặc tương đương), không echo plaintext
- [ ] AC-027: Given tài liệu help / `--help` cho lệnh create và join, When user đọc, Then có **cảnh báo ngắn** rằng password trên dòng lệnh có thể lưu trong shell history — khuyến nghị dùng prompt TUI hoặc phương thức an toàn hơn (chi tiết kỹ thuật defer architecture)

## Functional requirements

| ID | Requirement | Priority | Trace |
|----|-------------|----------|-------|
| FR-001 | Tạo room không password → room mở, join như cũ | Must | REQ-001 |
| FR-002 | Tạo room với password 1–32 ký tự → room yêu cầu password khi join | Must | REQ-001 |
| FR-003 | Host tạo room không cần nhập lại password trong phiên tạo | Must | REQ-001 |
| FR-004 | Password qua CLI flag và/hoặc TUI masked prompt khi tạo | Must | REQ-001 |
| FR-005 | Validation password: từ chối >32 ký tự và chuỗi chỉ whitespace | Must | REQ-001 |
| FR-006 | Join room có password với credential đúng → admit | Must | REQ-002 |
| FR-007 | Join room không password chỉ cần room ID | Must | REQ-002 |
| FR-008 | Join room có password qua CLI flag và/hoặc TUI masked prompt | Must | REQ-002 |
| FR-009 | Join room có password mà thiếu password → từ chối | Must | REQ-002 |
| FR-010 | Sai password → từ chối, không expose nội dung room | Must | REQ-003 |
| FR-011 | Thông báo lỗi join rõ, không leak password đúng | Must | REQ-003 |
| FR-012 | Không lockout/cooldown bắt buộc cho sai password v1 | Must | REQ-003 |
| FR-013 | Host kick member đã chọn từ panel Members (phím kick) | Must | REQ-004 |
| FR-014 | Kicked member disconnect ≤3s; biến mất khỏi member list | Must | REQ-004 |
| FR-015 | Kick ngay, không confirm bắt buộc v1 | Must | REQ-004 |
| FR-016 | Chỉ host có action kick; member bị từ chối nếu cố kick | Must | REQ-005 |
| FR-017 | Host không kick được chính mình | Must | REQ-005 |
| FR-018 | Kicked member nhận message rõ; rời TUI room | Must | REQ-006 |
| FR-019 | Kicked member được phép re-join ngay với ID + password đúng | Must | REQ-006 |
| FR-020 | Re-join sau kick = session mới, không giữ state phiên cũ | Should | REQ-006 |
| FR-021 | Password không hiển thị plaintext trên UI room cho non-host | Must | REQ-007 |
| FR-022 | TUI mask input password; help cảnh báo CLI history | Should | REQ-007 |

## Edge cases & error scenarios

| Scenario | Expected behavior |
|----------|-------------------|
| Tạo room, password = chuỗi rỗng / không nhập | Room mở; không yêu cầu password khi join |
| Tạo room, password chỉ khoảng trắng | Từ chối; thông báo validation |
| Tạo room, password 33+ ký tự | Từ chối; thông báo độ dài tối đa 32 |
| Join room mở, không gửi password | Admit bình thường |
| Join room có password, không gửi password | Từ chối; yêu cầu password |
| Join room có password, sai password | Từ chối; message lỗi xác thực; có thể thử lại |
| Join room không tồn tại | Từ chối; message lỗi (không admit) |
| Host kick member duy nhất | Kick thành công; room còn host |
| Host kick khi không chọn member | Không kick; không side effect |
| Member nhấn phím kick | Không có hiệu lực / bị server từ chối |
| Kicked member join lại ngay | Admit nếu password đúng (hoặc room mở) |
| Host tạo room có password rồi disconnect | Room vẫn yêu cầu password cho joiner mới (password gắn room, không đổi trong v1) |
| Nhiều joiner cùng lúc, một người sai password | Chỉ người sai bị từ chối; người đúng vẫn vào |
| Kick member đang vote/react | Kick vẫn thành công; state vote/reaction của member đó không còn active sau khi rời |

## Non-functional requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-001 | Security | Password room không hiển thị plaintext trên UI cho members; không ghi password vào log user-facing |
| NFR-002 | Security | Chi tiết lưu trữ và truyền password (hash, transport) — định nghĩa ở architecture, không plaintext lưu lâu dài nếu tránh được |
| NFR-003 | Usability | Thông báo lỗi join và kick bằng ngôn ngữ ngắn, actionable (thử lại password / liên hệ host) |
| NFR-004 | Performance | Kick disconnect member trong ≤3s p95 trong mạng LAN bình thường |
| NFR-005 | Compatibility | Hành vi room không password tương thích ngược 100% với flow join trước feature |

## Dependencies

- **Baseline TUI:** Panel Members và flow create/join room từ feature 002 (`docs/vibe/002-room-host-sci-fi-tui/`)
- **Room/session model:** Host identity và member list đã có trong hub hiện tại
- **CLI commands:** Lệnh create room và join room hiện có (bổ sung tùy chọn password)

## Out of scope (spec lock)

- User accounts / global auth
- Ban list / block re-join theo identity
- Moderator role
- Đổi/gỡ password sau khi tạo room
- Rate-limit / lockout sai password
- Kick qua CLI riêng (chỉ TUI panel Members v1)
- Confirm dialog trước kick
- Web / GUI client

## Traceability matrix

| AC | REQ | FR |
|----|-----|-----|
| AC-001 | REQ-001 | FR-001 |
| AC-002 | REQ-001 | FR-002, FR-003 |
| AC-003 | REQ-001 | FR-004 |
| AC-004 | REQ-001 | FR-004 |
| AC-005 | REQ-001 | FR-005 |
| AC-006 | REQ-002 | FR-006 |
| AC-007 | REQ-002 | FR-007 |
| AC-008 | REQ-002 | FR-008 |
| AC-009 | REQ-002 | FR-008 |
| AC-010 | REQ-002 | FR-009 |
| AC-011 | REQ-003 | FR-010 |
| AC-012 | REQ-003 | FR-011 |
| AC-013 | REQ-003 | FR-012 |
| AC-014 | REQ-004 | FR-013, FR-014, NFR-004 |
| AC-015 | REQ-004 | FR-014 |
| AC-016 | REQ-004 | FR-015 |
| AC-017 | REQ-004 | FR-013 |
| AC-018 | REQ-005 | FR-016 |
| AC-019 | REQ-005 | FR-016 |
| AC-020 | REQ-005 | FR-017 |
| AC-021 | REQ-006 | FR-018 |
| AC-022 | REQ-006 | FR-018 |
| AC-023 | REQ-006 | FR-019 |
| AC-024 | REQ-006 | FR-020 |
| AC-025 | REQ-007 | FR-021, NFR-001 |
| AC-026 | REQ-007 | FR-022 |
| AC-027 | REQ-007 | FR-022 |

## Gate G2 checklist

- [x] All requirements have testable AC
- [x] Edge cases documented
- [x] No implementation/tech details (defer to architecture)
- [x] Traceability: AC → REQ mapping complete
