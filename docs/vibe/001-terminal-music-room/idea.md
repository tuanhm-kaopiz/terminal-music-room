# Idea: Terminal Music Room

**Slug:** `terminal-music-room`
**Created:** 2026-07-01
**Status:** draft

## Problem statement

Remote và hybrid technical teams (software engineers, DevOps, QA, sysadmins) thiếu cách tạo “không khí làm việc chung” nhẹ nhàng mà không rời workflow terminal-first. Họ làm việc chủ yếu trong terminal và IDE; việc chuyển sang browser, Spotify, Discord hoặc YouTube để nghe nhạc cùng nhau gây context switching, mất tập trung, và các giải pháp hiện có không được thiết kế cho trải nghiệm đồng bộ trong terminal.

## Proposed solution (high level)

Xây dựng **Terminal Music Room** — công cụ terminal (CLI + TUI đơn giản) cho phép nhiều người vào cùng một room, nghe nhạc online đồng bộ theo thời gian thực, quản lý queue, chat text, và vote skip/priority. Host hoặc bất kỳ thành viên có thể thêm bài, pause/play/skip; mọi người nghe cùng bài, cùng vị trí playback. Trải nghiệm gọn trong terminal, tài nguyên nhẹ, phù hợp team kỹ thuật remote.

**Phạm vi release:** bản **full v1** (không cắt xuống MVP tối thiểu) — gồm room, sync playback, queue, chat, voting, YouTube audio, reconnect; UI theo spec chỉ cần **đơn giản**, không ưu tiên polish visual.

## Success looks like

- [ ] 2–20 người trong một room nghe cùng bài với drift playback ≤ 500ms (mục tiêu ≤ 200ms)
- [ ] Luồng end-to-end hoạt động: đăng nhập nickname → tạo/join/leave room → search/URL → play/pause/skip/seek → queue → chat → vote skip
- [ ] Join room < 2s; lệnh play/pause broadcast < 500ms
- [ ] Ngắt mạng/reconnect tự phục hồi và đồng bộ lại trạng thái phát
- [ ] Người dùng Ubuntu cài tool và dùng được mà không cần rời terminal
- [ ] TUI/CLI đơn giản như wireframe trong spec là đủ để ship; không chặn release vì thiếu animation hay theme

## Out of scope

- Video streaming, voice chat, screen sharing
- Mobile app và browser-based dashboard
- Nguồn nhạc ngoài YouTube (Spotify, SoundCloud, local files, internet radio) — giai đoạn sau
- OAuth / GitHub / Google login (v1 chỉ nickname ẩn danh)
- macOS, Windows, Docker runtime, cloud-hosted rooms — Ubuntu trước, mở rộng sau
- Private/password rooms, invite links — sau v1
- AI DJ, smart playlist, personalized recommendations
- Music production / DJ mixing chuyên sâu
- Polish UI cao (rich themes, animations, accessibility nâng cao ngoài baseline terminal)

## References

- Figma: (chưa có — UI đơn giản theo text layout trong spec)
- Docs: `docs/spec.md` (PRD Terminal Music Room v0.1)
- Related issues: —

## Raw notes

> /vibe-idea Tôi có ý tưởng xây dựng terminal để nghe nhạc online cùng nhau trong một room  
> chi tiết ý tưởng xem ở file docs/spec.md  
> Tôi muốn làm bản full để release sớm nhất có thể - UI ở spec chỉ là đơn sơ thôi

Tóm tắt từ PRD:
- Đối tượng: backend/DevOps/QA/remote dev, Linux users
- Core: shared room, YouTube audio, sync playback, queue, chat, voting (>50% để skip)
- CLI commands (`/play`, `/queue`, `/chat`) + TUI layout đơn giản
- Sync là requirement quan trọng nhất; xử lý latency, jitter, reconnect
- v1 Ubuntu; RAM < 300MB idle target

## Gate G0 checklist

- [x] Problem statement clear (who, what pain)
- [x] Success metric or "done looks like"
- [x] Out of scope (at least 1 item)
- [x] Feature folder created
