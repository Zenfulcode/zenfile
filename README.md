# Converzen

A cross-platform file converter for video and image files, built with [Wails](https://wails.io/), SvelteKit, and Go.

## Features

- 🎬 Video conversion (MP4, MOV, WebM, AVI, MKV, etc.)
- 🖼️ Image conversion (PNG, JPG, WebP, GIF, etc.)
- 📦 Batch conversion support
- 🎯 Drag and drop interface
- ⚡ Native performance

## Building

Converzen supports multiple build configurations for different distribution channels.

### Prerequisites

- [Go 1.21+](https://go.dev/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- [Bun](https://bun.sh/) (for frontend)
- FFmpeg (for non-App Store builds)

### Build Options

| Build Type      | Command                          | Video Backend   | Size  | Use Case                |
| --------------- | -------------------------------- | --------------- | ----- | ----------------------- |
| Standard        | `wails build`                    | System FFmpeg   | ~16MB | Development, Homebrew   |
| Embedded FFmpeg | `wails build -tags embed_ffmpeg` | Embedded FFmpeg | ~90MB | Standalone distribution |
| App Store       | `wails build -tags appstore`     | AVFoundation    | ~16MB | macOS App Store         |

### Standard Build (uses system FFmpeg)

```bash
wails build
```

Requires FFmpeg to be installed on the system:

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg

# Windows (with Chocolatey)
choco install ffmpeg
```

### Embedded FFmpeg Build

Bundles FFmpeg inside the application for standalone distribution. FFmpeg is compiled from source to ensure LGPL compliance.

```bash
# 1. Install build dependencies (macOS)
brew install nasm yasm pkg-config
# Optional for better codec support:
brew install x264 x265 libvpx aom opus libvorbis

# 2. Build FFmpeg from source for your target platform
./scripts/download-ffmpeg.sh darwin_arm64  # or darwin_amd64, linux_amd64, windows_amd64

# 3. Build with embedded FFmpeg
wails build -tags embed_ffmpeg
```

> **Note:** Only the FFmpeg binary for the target platform is embedded, not all platforms.

#### Build Script Options

```bash
# Build for current platform (auto-detected)
./scripts/download-ffmpeg.sh

# Build for specific platforms
./scripts/download-ffmpeg.sh darwin_arm64 darwin_amd64

# Show required dependencies
./scripts/download-ffmpeg.sh --deps

# Clean and rebuild
./scripts/download-ffmpeg.sh --clean darwin_arm64
```

The build script saves the FFmpeg source tarball at `pkg/ffmpeg/binaries/source/` for LGPL redistribution compliance.

### App Store Build (macOS only)

Uses Apple's native AVFoundation framework instead of FFmpeg. This is required for App Store distribution due to licensing requirements.

```bash
wails build -tags appstore
```

**Limitations of AVFoundation:**

- Output formats limited to: MP4, MOV, M4V
- Fewer codec options than FFmpeg
- macOS only

## Development

```bash
# Run in development mode
wails dev

# Run frontend only
cd frontend && bun run dev
```

## Architecture

```
├── app.go                 # Main application logic
├── app_ffmpeg.go          # FFmpeg initialization (non-App Store)
├── app_avfoundation.go    # AVFoundation initialization (App Store)
├── internal/
│   ├── config/            # Configuration management
│   ├── services/          # Business logic
│   │   ├── video_converter.go              # FFmpeg video converter
│   │   └── video_converter_avfoundation.go # AVFoundation converter
│   └── ...
├── pkg/
│   └── ffmpeg/            # FFmpeg wrapper and embedding
│       ├── ffmpeg.go      # FFmpeg command wrapper
│       ├── embed_*.go     # Platform-specific embedding
│       └── binaries/      # FFmpeg binaries (for embedding)
└── frontend/              # SvelteKit frontend
```

## License

This project is licensed under the GPL-3.0 License - see [LICENSE.md](LICENSE.md).

### Third-Party Licenses

See [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md) for information about FFmpeg and other dependencies.

**Important:**

- App Store builds use AVFoundation (no FFmpeg) and are fully compliant with App Store guidelines.
- Non-App Store builds that include FFmpeg must comply with LGPL 2.1 requirements.

## Tech Stack

- **Backend:** Go, Wails v2
- **Frontend:** SvelteKit, Svelte 5, TypeScript, Tailwind CSS 4
- **Video Processing:** FFmpeg (standard) / AVFoundation (App Store)
