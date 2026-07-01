# Clarify: Room Host Sci-Fi TUI (v2)

**Slug:** `room-host-sci-fi-tui`
**Status:** complete
**Gate G1:** ✅ pass

## Resolved questions

| # | Question | Answer | Decided by |
|---|----------|--------|------------|
| 1 | Host mở TUI sci-fi bằng cách nào? | **Thay thế TUI member hiện tại** — khi dùng `join` / `tui`, mọi người (kể cả host) vào **một TUI sci-fi thống nhất** thay layout đơn giản v1 | user |
| 2 | Phạm vi chức năng trong TUI host v2? | **Full parity** — mọi thao tác host có trên CLI phải khả dụng trong TUI (playback, queue admin remove/reorder, thêm/search bài, chat, vote, xem members) | user |
| 3 | Hướng thị giác sci-fi ưu tiên? | **Cyberpunk / neon tối** — nền tối, accent neon (tham chiếu cảm giác Blade Runner, Cyberpunk 2077) | user |
| 4 | TUI member v1 xử lý thế nào? | **Cả host và member đều chuyển sang TUI sci-fi mới** — không giữ layout member cũ; host có đủ quyền, member thấy/làm đúng phạm vi quyền member v1 | user |
| 5 | Yêu cầu terminal cho palette? | **16 màu tối thiểu** — thiết kế cho terminal cơ bản; không bắt buộc truecolor | user |
| 6 | CLI host có còn không? | **Giữ CLI** — TUI là trải nghiệm chính khi `join`/`tui`; CLI vẫn dùng được cho scripting và fallback (suy ra từ v1, không bị user yêu cầu gỡ) | AI (assumption — confirm in spec if needed) |
| 7 | Kích thước terminal tối thiểu? | **80×24** trở lên — theo idea.md; layout phải usable ở kích thước này | idea.md |
| 8 | Mức animation? | **Nhẹ / micro-feedback only** — không animation nặng gây lag (theo idea.md out of scope) | idea.md |

## Open questions (blocking)

| # | Question | Owner | Blocking? |
|---|----------|-------|-----------|

> Không còn câu hỏi blocking.

## Scope

### In scope

- **Một TUI sci-fi thống nhất** thay TUI member đơn giản v1 — áp dụng cho **cả host và member** khi vào room qua `join` / `tui`
- **Phong cách cyberpunk**: nền tối, accent neon, typography/layout gợi HUD điều khiển — vẫn đọc được trên terminal 16 màu
- **Host full parity trong TUI**: play/pause/skip/seek, thêm bài (search/URL), xem/sắp queue, **remove/reorder queue** (host-only), xem members, chat, vote skip/priority, reactions — tương đương CLI host v1
- **Member** trong cùng shell sci-fi: thao tác theo quyền member v1 (playback democratic, add queue, chat, vote, react); **không** có queue admin host-only trừ khi là host
- Thông tin chính **hiển thị ngay màn hình mặc định**: bài đang phát, tiến độ, queue kế tiếp, số người nghe
- Palette và component **tối ưu cho 16 màu** (ANSI); degrade an toàn nếu terminal hạn chế hơn
- Ubuntu terminal tiêu chuẩn; không GUI/browser
- **Không đổi** logic backend/sync playback — chỉ lớp client TUI + điều khiển

### Out of scope

- **Nhiều theme / skin** do user chọn — v2 một visual direction cyberpunk
- **Truecolor bắt buộc** hoặc gradient phức tạp phụ thuộc 24-bit
- Animation nặng, spinner toàn màn hình, hoặc hiệu ứng làm lag terminal / CPU cao
- Web dashboard, mobile app, GUI ngoài terminal
- Tính năng nhạc/sync mới (nguồn khác YouTube, voice, video, private room, OAuth, v.v.)
- Thay đổi server-authoritative sync, voting rules, hoặc permission model backend
- Accessibility nâng cao ngoài baseline terminal (screen reader đầy đủ, high-contrast mode riêng) — có thể phase sau
- macOS / Windows client (giữ Ubuntu như v1)

### Scope change vs idea.md

`idea.md` ghi out of scope: *"TUI sci-fi cho thành viên không phải host"*. User đã **mở rộng**: cả member dùng TUI sci-fi mới. **clarify.md wins** — spec v2 cover unified sci-fi TUI với phân quyền host/member trong cùng giao diện.

## Actors / users

| Actor | Role | Key actions trong TUI sci-fi |
|-------|------|------------------------------|
| Room host | Người tạo room | Mọi action member + remove/reorder queue; dashboard đầy đủ |
| Room member | Người đã join | Play/pause/skip/seek, add queue, chat, vote, react, xem queue/members — không queue admin |
| Anonymous visitor | Chưa trong room | Không thấy TUI room — login/join qua CLI flow như v1 trước khi vào TUI |

## Assumptions

1. **Entry point**: `music-room join <slug>` (và lệnh `tui` tương đương nếu có) mở **TUI sci-fi mới** thay TUI đơn giản v1 — không cần subcommand `host` riêng.
2. **CLI coexistence**: Lệnh CLI v1 vẫn hoạt động song song; TUI không thay thế hoàn toàn CLI cho automation.
3. **Permission model** giữ nguyên v1: democratic playback cho member; queue admin host-only — chỉ khác **presentation** và **một màn hình thống nhất**.
4. **Cyberpunk trên 16 màu**: Dùng ANSI palette cố định (vd. nền đen/xám đậm, accent magenta/cyan/yellow neon) — chấp nhận giới hạn so với truecolor reference CP2077.
5. **Terminal 80×24** là layout baseline; terminal rộng/cao hơn có thể tận dụng không gian nhưng không bắt buộc cho ship.
6. **v1 functional reference**: `docs/vibe/001-terminal-music-room/` (spec/clarify) định nghĩa hành vi; feature này là **UI v2** trên cùng capability set.
7. Đánh giá “cảm giác sci-fi/cyberpunk” dùng **user feedback nội bộ** (≥3 người) như success metric trong idea.md.

## Risks & constraints

| Risk | Impact | Mitigation |
|------|--------|------------|
| Cyberpunk neon trên **16 màu** trông “phẳng” hoặc khó phân biệt panel | Không đạt cảm giác “trending” | Palette ANSI cố định + border/box-drawing rõ; test trên GNOME Terminal / default Ubuntu |
| **Full parity** host trong một màn hình 80×24 | Chật, phải scroll hoặc tab — trái idea “nhìn thấy ngay” | Spec định layout ưu tiên (now playing + queue + members); chat/vote có thể panel phụ hoặc toggle — không ẩn hoàn toàn |
| Scope mở rộng sang **member TUI** tăng effort vs idea ban đầu | Slip timeline v2 | Spec tách host-only vs shared UI components; ship một shell, role-based actions |
| Unicode box-drawing / emoji trên terminal cũ | Vỡ layout | Spec ghi fallback ASCII borders; emoji theo baseline v1 |
| Animation quá mức | Lag trên máy yếu | Chỉ micro-feedback (highlight, blink nhẹ); no full-screen effects |

## Gate G1 checklist

- [x] No blocking open questions
- [x] Scope bounded (in/out explicit)
- [x] Actors identified
- [x] Assumptions listed
