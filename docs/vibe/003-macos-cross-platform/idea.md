# Idea: macOS Cross-Platform Support (V0.2.1)

**Slug:** `macos-cross-platform`
**Created:** 2026-07-01
**Status:** draft

## Problem statement

Người dùng macOS muốn dùng Terminal Music Room nhưng hiện chỉ có thể chạy trên Linux/Ubuntu. Họ không thể host phòng từ máy Mac hoặc tham gia phòng một cách đáng tin cậy khi bạn bè/colleague dùng OS khác.

Pain cụ thể:
- Host trên Ubuntu, khách trên Mac (hoặc ngược lại) không có trải nghiệm được hỗ trợ rõ ràng cho V0.2.1.
- Nhóm mixed-OS (phổ biến trong team dev) bị loại khỏi use case "nghe nhạc chung qua terminal".
- Thiếu binary/cách cài trên macOS khiến macOS user không tham gia được ecosystem hiện có.

## Proposed solution (high level)

Mở rộng Terminal Music Room sang macOS trong phạm vi V0.2.1: người dùng macOS có thể cài và chạy CLI/TUI như trên Linux. Phòng nhạc phải hoạt động khi host và khách ở hai nền tảng khác nhau — ví dụ host macOS + khách Ubuntu, hoặc host Ubuntu + khách macOS — với cùng luồng tạo/join phòng, đồng bộ queue và điều khiển phát nhạc.

## Success looks like

- [ ] Người dùng macOS cài và chạy được `music-room` từ Terminal (host hoặc join) mà không cần Linux
- [ ] Host macOS + khách Ubuntu: khách join được, thấy queue/trạng thái phát, thao tác điều khiển cơ bản hoạt động
- [ ] Host Ubuntu + khách macOS: tương tự, hai chiều đều pass
- [ ] Release V0.2.1 publish artifact macOS (cùng mức "sẵn sàng dùng" như bản Linux hiện tại)
- [ ] README/hướng dẫn cài trên macOS được cập nhật cho end user

## Out of scope

- Hỗ trợ Windows trong V0.2.1
- GUI native macOS (chỉ terminal/TUI như hiện tại)
- Phân phối qua Mac App Store hoặc notarization phức tạp (trừ khi clarify/spec yêu cầu tối thiểu)
- Thay đổi giao thức/phòng cho web client hoặc mobile
- Tối ưu audio backend ngoài phạm vi cần cho parity host/guest cross-platform

## References

- Figma: —
- Docs: `docs/vibe/001-terminal-music-room/`, `docs/vibe/002-room-host-sci-fi-tui/`
- Related issues: V0.2.1 milestone — mở rộng terminal sang macOS

## Raw notes

> V0.2.1 tôi muốn mở rộng terminal sang mac os
> Host có thể là terminal của mac os, khách có thể là ubuntu và ngược lại

## Gate G0 checklist

- [x] Problem statement clear (who, what pain)
- [x] Success metric or "done looks like"
- [x] Out of scope (at least 1 item)
- [x] Feature folder created
