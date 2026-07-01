# Supported platforms — Terminal Music Room

Reference for **V0.2.1** client support. Server (`music-roomd`) remains Linux/Docker for operators.

## Client matrix

| OS | Architectures | Install | Playback deps |
|----|---------------|---------|---------------|
| **Ubuntu** 22.04 / 24.04 | amd64 | `.deb`, tarball, source | `apt install mpv yt-dlp ffmpeg` |
| **Debian** 12+ and Debian-based (Mint, Pop!_OS, …) | amd64 | `.deb`, tarball, source | same apt packages |
| **macOS** 13 Ventura+ | arm64 (Apple Silicon), amd64 (Intel) | GitHub Release tarball, Homebrew | `brew install mpv yt-dlp ffmpeg` |

## Out of scope (V0.2.1)

- Windows
- Fedora, Arch, NixOS, and other non-Debian Linux families
- macOS App Store / notarized distribution
- Bundled mpv/ffmpeg inside the `music-room` binary

## Release artifacts (tag `v0.2.1`)

| Asset | Contents |
|-------|----------|
| `terminal-music-room_{VERSION}_linux_amd64.tar.gz` | `music-room`, `music-roomd`, `SHA256SUMS` |
| `terminal-music-room_{VERSION}_darwin_arm64.tar.gz` | `music-room`, `SHA256SUMS` |
| `terminal-music-room_{VERSION}_darwin_amd64.tar.gz` | `music-room`, `SHA256SUMS` |
| `music-room_{VERSION}-1_amd64.deb` | client + apt deps metadata |
| `music-roomd_{VERSION}-1_amd64.deb` | server (Linux operators) |

## Cross-platform rooms

Any combination of in-scope clients may share a room:

- Host macOS ↔ guest Ubuntu/Debian
- Host Ubuntu/Debian ↔ guest macOS

Sync is **server-authoritative**; each client plays audio locally via **mpv**. See [E2E.md](E2E.md) §Cross-platform.

## macOS Gatekeeper (unsigned binary)

Release binaries are **not** code-signed or notarized. On first run, macOS may block the app:

1. **System Settings → Privacy & Security → Open Anyway**, or
2. Remove quarantine: `xattr -dr com.apple.quarantine /path/to/music-room`

## Minimum tool versions

Documented in README; pin in your environment:

- `mpv` — recent stable from apt/Homebrew
- `yt-dlp` — weekly releases recommended (YouTube changes often)
- `ffmpeg` — recommended on all platforms
