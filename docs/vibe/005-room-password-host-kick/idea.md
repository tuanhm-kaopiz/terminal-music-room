# Idea: Room Password & Host Kick

**Slug:** `room-password-host-kick`
**Created:** 2026-07-02
**Status:** draft

## Problem statement

Room hosts running shared listening sessions have no way to control who enters. Anyone who knows the room identifier can join immediately, which is a problem for private gatherings or when unwanted guests disrupt playback and chat. Once someone is in the room, the host also cannot remove them — they must wait for the person to leave on their own or shut down the entire room.

## Proposed solution (high level)

When creating a room, the host can optionally set a password. Joiners must supply the correct password before they are admitted. Inside an active room, the host can kick a specific member: the kicked person is disconnected from the session and must go through the join flow again if they want to re-enter (still subject to password if one is set).

## Success looks like

- [ ] Host can create a room with an optional password; rooms without a password behave as today (open join)
- [ ] A joiner who enters the wrong password is rejected with a clear message and is not admitted to the room
- [ ] Host can kick any non-host member from the room; the kicked member is disconnected promptly
- [ ] Only the host can kick; regular members have no kick action
- [ ] After being kicked, a member sees feedback that they were removed (not a silent disconnect)

## Out of scope

- User accounts or global authentication — room password only, not per-user login
- Permanent ban list or blocking re-join by identity after kick
- Moderator/delegate roles — only the room host can kick in v1
- Changing or removing room password after the room is created
- Rate-limiting or lockout policies for failed password attempts (beyond basic rejection)

## References

- Figma:
- Docs:
- Related issues:

## Raw notes

Hiện tại khi tạo room chưa có password. Giờ tôi muốn thêm password cho room. Ngoài ra host room có thể kick người ra khỏi room.

## Gate G0 checklist

- [x] Problem statement clear (who, what pain)
- [x] Success metric or "done looks like"
- [x] Out of scope (at least 1 item)
- [x] Feature folder created
