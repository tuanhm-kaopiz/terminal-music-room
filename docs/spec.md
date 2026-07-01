# Product Requirement Document (PRD)
# Terminal Music Room (Working Title)

Version: 0.1  
Status: Draft  
Author: Product Team  
Date: 2026-07-01

---

# 1. Product Overview

## 1.1 Product Name
Terminal Music Room (working title)

Possible branding:
- TermMusic
- CLI Jam
- SyncBeat
- DevTunes
- Terminal FM

---

# 2. Product Vision

Terminal Music Room là một nền tảng cho phép nhiều người nghe nhạc cùng nhau theo thời gian thực trong một room chia sẻ, với trải nghiệm hoàn toàn trên terminal (CLI/TUI).

Sản phẩm hướng đến nhóm người dùng kỹ thuật như:
- Software Engineers
- QA/Testers
- DevOps
- SysAdmins
- Remote technical teams

Thay vì mở browser, Spotify hoặc Discord, người dùng có thể:
- Join room
- Nghe nhạc đồng bộ
- Chat text
- Vote bài hát
- Quản lý queue

Toàn bộ trải nghiệm diễn ra trong terminal.

Ví dụ:

```bash
music-room join backend-team
```

---

# 3. Problem Statement

Hiện tại, khi làm việc remote hoặc hybrid, các team kỹ thuật thường gặp các vấn đề:

## 3.1 Lack of shared working atmosphere
Remote team thiếu cảm giác làm việc cùng nhau.

Ví dụ:
- Không còn cảm giác "cùng ngồi làm"
- Thiếu social interaction nhẹ nhàng

---

## 3.2 Context switching
Người dùng kỹ thuật thường làm việc chủ yếu trong:
- Terminal
- IDE
- Browser

Việc chuyển qua app khác để:
- nghe nhạc
- chat
- social interaction

gây mất tập trung.

---

## 3.3 Existing solutions are not optimized for developers

Các nền tảng hiện tại:
- Spotify Jam
- Discord
- YouTube
- Apple Music SharePlay

Không được thiết kế cho workflow terminal-first.

---

# 4. Goals

## 4.1 Business Goals
- Build niche product cho cộng đồng developer
- Tạo sản phẩm open-source hoặc SaaS có khả năng viral
- Xây dựng community around terminal culture

---

## 4.2 Product Goals
- Real-time music synchronization
- Low-latency playback
- Fully terminal-native experience
- Minimal resource usage
- Cross-platform support

---

# 5. Non-Goals (MVP)

Các tính năng KHÔNG thuộc phạm vi MVP:

- Video streaming
- Voice chat
- Screen sharing
- Mobile app
- Browser-based dashboard
- Music production / DJ mixing
- AI recommendation engine

---

# 6. Target Users

## Primary Users
- Backend Engineers
- DevOps Engineers
- QA Engineers
- Remote Developers
- Linux Users

---

## Secondary Users
- MacOS developers
- Open-source contributors
- Indie hackers
- Startup teams

---

# 7. Supported Platforms

## MVP
- Ubuntu Linux

---

## Future Expansion
- MacOS
- Windows
- Docker runtime
- Cloud-hosted rooms

---

# 8. Core Product Concept

Người dùng cài đặt CLI tool.

Ví dụ:

```bash
brew install music-room
```

hoặc

```bash
apt install music-room
```

Sau khi cài đặt, user có thể:
- Create room
- Join room
- Play music from YouTube
- Sync playback with room members

Room hoạt động theo shared playback state.

Tức là:
- cùng bài nhạc
- cùng playback position
- cùng pause/play state

---

# 9. User Stories

## Room Management

### US-001
As a user, I want to create a room so others can join.

### US-002
As a user, I want to join a room using room ID.

### US-003
As a user, I want to leave a room anytime.

---

## Playback

### US-004
As a user, I want everyone in the room to hear the same song.

### US-005
As a user, I want playback to stay synchronized.

### US-006
As a user, I want pause/play actions to sync.

---

## Queue

### US-007
As a user, I want to add songs to queue.

### US-008
As a user, I want to see upcoming songs.

### US-009
As a user, I want to vote skip.

---

## Social

### US-010
As a user, I want to chat inside room.

### US-011
As a user, I want to react to songs.

---

# 10. Functional Requirements

# 10.1 Authentication

MVP:
- Anonymous login via nickname

Example:
```bash
music-room login --name kaopiz
```

Optional future:
- OAuth login
- GitHub login
- Google login

---

# 10.2 Room Management

Supported actions:
- Create room
- Join room
- Leave room
- List active users

Room attributes:
- Room ID
- Room Name
- Host
- Created Time
- Members

---

# 10.3 Music Playback

Supported actions:
- Play
- Pause
- Resume
- Skip
- Seek

Playback state:
- Current song
- Current timestamp
- Duration
- Status

Possible statuses:
- Playing
- Paused
- Buffering
- Ended

---

# 10.4 Music Source

MVP source:
- YouTube audio streaming

Supported:
- Search by keyword
- Play by URL

Examples:

```bash
/play lofi hip hop
```

or

```bash
/play https://youtube.com/xxxxx
```

---

Future sources:
- Spotify
- SoundCloud
- Local files
- Internet radio

---

# 10.5 Queue System

Queue actions:
- Add
- Remove
- Reorder
- Skip

Queue metadata:
- Song title
- Duration
- Added by
- Source
- Added time

---

# 10.6 Chat

Users can send text messages inside room.

Examples:

```bash
/chat hello team
```

Chat supports:
- text only
- emoji
- system messages

Examples:
- User joined
- Song changed
- Vote started

---

# 10.7 Voting System

Supported voting:
- Vote skip current song
- Vote next song priority

Rules:
- >50% members required

Example:
- 5 users online
- 3 votes required

---

# 10.8 Terminal UI

System supports:

## CLI mode
Command-based interaction

Example:
```bash
/play
/queue
/chat
```

---

## TUI mode
Visual interactive mode

Example layout:

```text
+------------------------------------------------+
| Room: Backend Team                             |
| Song: Lofi Coding Mix                          |
| Time: 02:31 / 12:20                            |
+------------------------------------------------+
| Online Users                                   |
| - kaopiz                                       |
| - tester01                                     |
+------------------------------------------------+
| Queue                                          |
| 1. Night Drive                                 |
| 2. Rain Sounds                                 |
+------------------------------------------------+
| Chat                                           |
| tester01: bug nhiều quá                        |
+------------------------------------------------+
```

---

# 11. Synchronization Requirements

Đây là requirement quan trọng nhất.

System phải đảm bảo:
- playback sync giữa users
- low drift
- minimal latency

Acceptable sync drift:
- <= 500ms (acceptable)
- <= 200ms (ideal)

System phải xử lý:
- network delay
- reconnect
- jitter

---

# 12. Performance Requirements

## Latency
- Join room < 2 seconds
- Play command broadcast < 500ms

---

## Scalability
MVP:
- 1 room: 2–20 users

Future:
- 1000+ concurrent rooms

---

## Resource Usage
Ubuntu machine target:
- RAM < 300MB
- CPU minimal while idle

---

# 13. Security Requirements

- Secure room access
- Prevent room hijacking
- Abuse protection
- Rate limiting

Future:
- Private rooms
- Password protected rooms

---

# 14. Error Handling

System should handle:

- User disconnect
- Internet loss
- Playback failure
- Source unavailable
- YouTube throttling

Recovery:
- Auto reconnect
- Resume playback state
- Sync correction

---

# 15. Constraints

Known constraints:
- YouTube streaming limitations
- Network latency variance
- OS-level audio dependencies
- Cross-platform audio differences

---

# 16. Success Metrics

## Product Metrics
- Daily active rooms
- Average session duration
- Average room size

---

## Technical Metrics
- Playback sync accuracy
- Command latency
- Crash-free rate

---

# 17. Future Roadmap

## Phase 1 — MVP
- Ubuntu support
- Room system
- YouTube audio
- Playback sync
- Queue
- Chat

---

## Phase 2
- MacOS support
- Voting
- Reactions
- Better TUI

---

## Phase 3
- Spotify integration
- Private rooms
- Invite links
- Cloud rooms

---

## Phase 4
- AI DJ
- Smart playlist
- Personalized recommendations

---

# 18. Risks

Major risks:
- Audio synchronization complexity
- YouTube legal/technical limitations
- Cross-platform playback issues

---

# 19. Open Questions

- Host-authoritative hay server-authoritative sync?
- YouTube stream via API hay indirect?
- Open-source hay commercial?
- Self-hosted hay managed cloud?