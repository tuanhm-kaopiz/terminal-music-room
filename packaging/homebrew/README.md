# Homebrew — Terminal Music Room

## Install from local formula (dev / pre-release)

After building macOS tarballs (`./packaging/build-macos.sh 0.2.1`):

```bash
# Extract the tarball for your architecture, then:
brew install --formula ./packaging/homebrew/Formula/music-room.rb
```

For a **local path install** without release URLs, temporarily edit the formula to use:

```ruby
url "file:///absolute/path/to/terminal-music-room_0.2.1_darwin_arm64.tar.gz"
sha256 "..." # shasum -a 256 < tarball
```

## Install from GitHub release (end users)

1. Install dependencies:

```bash
brew install mpv yt-dlp ffmpeg
```

2. Download the matching release tarball (`darwin_arm64` or `darwin_amd64`) from [GitHub Releases](https://github.com/tuanhm-kaopiz/terminal-music-room/releases), extract, and put `music-room` on your `PATH`.

   Or use the published formula after `v0.2.1` release (update `url` + `sha256` in the formula — see `bump-formula.sh`).

## Bump formula after release

```bash
./packaging/homebrew/bump-formula.sh 0.2.1
```

Updates `version`, `url`, and `sha256` in `Formula/music-room.rb` from `dist/*.tar.gz`.

## Tap workflow (maintainers)

To publish a tap repo, copy `Formula/music-room.rb` into `homebrew-tap/Formula/` and document:

```bash
brew tap tuanhm-kaopiz/tap
brew install music-room
```
