//go:build !appstore

package config

import (
	"os"
	"path/filepath"
	"runtime"

	"converzen/pkg/ffmpeg"
)

// findFFmpeg attempts to find the FFmpeg executable
// First tries to use the embedded binary, then falls back to system paths
func findFFmpeg(dataDir string) string {
	// First, try to extract and use the embedded FFmpeg binary
	if ffmpeg.HasEmbeddedFFmpeg() {
		if embeddedPath, err := ffmpeg.GetEmbeddedFFmpegPath(dataDir); err == nil {
			return embeddedPath
		}
	}

	// Fall back to checking common locations based on OS
	var candidates []string

	switch runtime.GOOS {
	case "windows":
		candidates = []string{
			"ffmpeg.exe",
			filepath.Join(os.Getenv("ProgramFiles"), "ffmpeg", "bin", "ffmpeg.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "ffmpeg", "bin", "ffmpeg.exe"),
		}
	case "darwin":
		candidates = []string{
			"ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/opt/homebrew/bin/ffmpeg",
		}
	default:
		candidates = []string{
			"ffmpeg",
			"/usr/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
		}
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fall back to assuming it's in PATH
	return "ffmpeg"
}
