# Clarify: Room Password & Host Kick

**Slug:** `room-password-host-kick`
**Status:** complete
**Gate G1:** ✅ pass

## Resolved questions

| # | Question | Answer | Decided by |
|---|----------|--------|------------|
| 1 | Password khi tạo room: bắt buộc hay tùy chọn? | **Tùy chọn** — không nhập password = room mở như hiện tại | user |
| 2 | Sau khi bị kick, member có join lại được không? | **Được join lại ngay** nếu biết room ID + password đúng; không có ban vĩnh viễn | user |
| 3 | Host kick member qua đâu trong TUI? | **Panel Members** — chọn member rồi nhấn phím kick (vd. `K` hoặc `Del`) | user |
| 4 | Nhập password khi tạo/join (CLI + TUI)? | **CLI flag `--password`** khi tạo/join + **prompt TUI** tương tác khi cần (ẩn ký tự khi gõ) | user |
| 5 | Quy tắc password? | **1–32 ký tự**, không ràng buộc ký tự đặc biệt; **chuỗi rỗng = không password** | user |
| 6 | Joiner nhập sai password? | **Từ chối join** với thông báo rõ; không rate-limit/lockout ngoài reject cơ bản (theo `idea.md`) | AI (suy ra từ idea.md) |
| 7 | Ai được kick? | **Chỉ non-host members**; host không kick chính mình; member thường không có action kick | AI (suy ra từ idea.md) |
| 8 | Phạm vi client? | **CLI + TUI terminal** — cùng flow tạo/join room hiện có, bổ sung password và kick | AI (suy ra từ product context) |

## Open questions (blocking)

| # | Question | Owner | Blocking? |
|---|----------|-------|-----------|

> Không còn câu hỏi blocking.

## Scope

### In scope

- **Tạo room với password tùy chọn** — host có thể đặt password qua CLI flag `--password` hoặc prompt TUI; bỏ trống = room mở
- **Join room có password** — joiner phải cung cấp password đúng qua CLI flag hoặc prompt TUI trước khi vào room
- **Từ chối join sai password** — thông báo lỗi rõ ràng, không admit vào room
- **Host kick member** từ **panel Members** trong TUI (chọn member + phím kick)
- **Ngắt kết nối ngay** khi bị kick; member thấy **feedback** rõ (không silent disconnect)
- **Re-join sau kick** — được phép join lại ngay với room ID + password đúng (không ban list)
- **Quyền kick** — chỉ host; members không có action kick

### Out of scope

- Tài khoản user / global authentication
- Ban vĩnh viễn hoặc block re-join theo identity sau kick
- Role moderator/delegate — chỉ host kick trong v1
- Đổi hoặc gỡ password sau khi room đã tạo
- Rate-limit / lockout cho sai password nhiều lần
- Kick qua CLI command riêng (chỉ TUI panel Members trong v1)
- Web UI / GUI ngoài terminal

### Scope alignment vs idea.md

`idea.md` khớp với các quyết định user. `clarify.md` bổ sung chi tiết: password **tùy chọn**, **1–32 ký tự**, input qua **CLI flag + TUI prompt**, kick qua **panel Members**, **re-join ngay** sau kick.

## Actors / users

| Actor | Role | Key actions |
|-------|------|-------------|
| Room host | Người tạo và sở hữu room | Tạo room (có/không password); kick member từ panel Members |
| Joiner | Người muốn vào room | Nhập room ID + password (nếu room có); nhận lỗi nếu sai password |
| Room member | Đã join thành công | Bị kick bởi host; nhận thông báo bị remove; có thể join lại |
| Server | Hub điều phối room | Xác thực password khi join; thực thi kick và ngắt kết nối member |

## Assumptions

1. **Room ID** vẫn là cách chính để xác định room; password là lớp bảo vệ thêm, không thay room ID.
2. **Password lưu server-side** cho phiên room; chi tiết hash/transport thuộc phase Architecture — product yêu cầu không lộ password trong log/UI công khai.
3. **Prompt TUI ẩn ký tự** khi nhập password (mask input) — chuẩn terminal cho secret.
4. **Kick disconnect WebSocket/session** hiện có; kicked client quay về màn hình join hoặc exit với message.
5. **Host identity** đã có từ hệ thống room hiện tại — feature không đổi mô hình host transfer.
6. **Panel Members** đã tồn tại trong TUI (feature 002) — feature này thêm action kick, không build panel mới từ đầu.
7. **CLI `--password`** optional trên lệnh create/join hiện có; không password flag + room không password = join như cũ.

## Risks & constraints

| Risk | Impact | Mitigation |
|------|--------|------------|
| Password plaintext trên CLI history | Lộ secret qua shell history | Spec ghi cảnh báo; ưu tiên TUI masked prompt; architecture cân nhắc env var hoặc `--password-stdin` |
| Kick gây confuse member (tưởng mất mạng) | UX kém | Message rõ "removed by host"; spec AC cho kicked state |
| Re-join ngay sau kick — harass loop | Host kick lặp lại cùng người | Chấp nhận v1 (no ban); ghi risk; có thể backlog cooldown/ban |
| Password yếu (1 ký tự) | Room dễ đoán | User chọn rules đơn giản; document trong help text |
| Host kick nhầm | Member mất session đột ngột | Confirm optional hoặc undo window — **out of scope v1**; kick ngay lập tức |

## Gate G1 checklist

- [x] No blocking open questions
- [x] Scope bounded (in/out explicit)
- [x] Actors identified
- [x] Assumptions listed
