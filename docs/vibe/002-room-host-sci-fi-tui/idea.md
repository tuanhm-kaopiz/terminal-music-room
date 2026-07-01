# Idea: Room Host Sci-Fi TUI (v2)

**Slug:** `room-host-sci-fi-tui`
**Created:** 2026-07-01
**Status:** draft

## Problem statement

**Chủ room (host)** hiện quản lý phòng nhạc chủ yếu qua lệnh CLI — không có giao diện TUI dành riêng. Trong khi thành viên có thể có trải nghiệm terminal cơ bản, host phải nhớ nhiều lệnh, thiếu bức tranh tổng thể (queue, trạng thái phát, người nghe, vote) trên một màn hình. Điều này làm việc điều phối room (play/pause/skip, duyệt queue, theo dõi phòng) chậm và kém “cảm giác” so với kỳ vọng của một công cụ nhạc terminal hiện đại.

## Proposed solution (high level)

Trong **v2**, bổ sung **TUI dành cho chủ room** với phong cách **khoa học viễn tưởng / sci-fi trending** — cảm giác như bảng điều khiển tàu vũ trụ hoặc HUD trong phim cyberpunk, nhưng vẫn dùng được trong terminal thật. Host mở một màn hình duy nhất để xem và điều khiển room: trạng thái phát, queue, danh sách người nghe, chat/vote, và các thao tác điều phối thường dùng — không cần gõ lệnh rời rạc. Trải nghiệm phải **ấn tượng về mặt thị giác** (màu, typography, layout, micro-feedback) trong khi vẫn **đọc được và thao tác nhanh** trên terminal 80×24 trở lên.

## Success looks like

- [ ] Host có thể mở TUI và thực hiện các thao tác điều phối chính (play/pause/skip, xem/sắp queue, xem người trong room) mà không cần thoát ra CLI
- [ ] Ít nhất 3 người dùng thử nội bộ mô tả giao diện là “sci-fi / futuristic / trending” mà không cần giải thích thêm
- [ ] Thông tin quan trọng (bài đang phát, tiến độ, queue kế tiếp, số người nghe) nhìn thấy ngay trên màn hình mặc định, không cần tab ẩn
- [ ] Host hoàn thành luồng “mở room → thêm bài → skip → xem phản hồi vote” trong TUI nhanh hơn hoặc bằng CLI hiện tại (cảm nhận chủ quan + ít bước hơn)
- [ ] Giao diện vẫn dùng được trên terminal Ubuntu tiêu chuẩn (không yêu cầu GUI hay trình duyệt)

## Out of scope

- TUI sci-fi cho **thành viên không phải host** — có thể làm sau, v2 tập trung host trước
- Thay đổi logic backend/sync playback — chỉ nâng cấp lớp hiển thị và điều khiển host
- Theme tùy chỉnh do user chọn (nhiều skin) — v2 chỉ một visual direction sci-fi
- Hiệu ứng animation nặng làm lag terminal hoặc tốn CPU đáng kể
- Web dashboard, mobile app, hoặc GUI ngoài terminal
- Video, voice chat, hoặc tính năng nhạc mới (chỉ UI/UX cho host trên tính năng v1 đã có)

## References

- Figma: (chưa có — có thể tham khảo mood board sci-fi terminal / cyberpunk HUD)
- Docs: `docs/vibe/001-terminal-music-room/` (v1 — host thiếu TUI, UI member đơn giản)
- Related issues: —

## Raw notes

> /vibe-idea Hiện tại chủ room không có giao diện tui. Ở v2 tôi muốn nâng cấp giao diện TUI thật "trending" kiểu cảm giác khoa học, viễn tưởng

Ý định sản phẩm:
- **Actor chính:** chủ room (host), không phải guest/member thường
- **Phiên bản:** v2 (sau v1 đã ship chức năng cốt lõi)
- **Cảm xúc mong muốn:** trending, sci-fi, futuristic — “cool” nhưng vẫn là terminal tool thực dụng
- **Pain:** host không có TUI, phải dùng CLI rời rạc

## Gate G0 checklist

- [x] Problem statement clear (who, what pain)
- [x] Success metric or "done looks like"
- [x] Out of scope (at least 1 item)
- [x] Feature folder created
