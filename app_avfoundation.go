//go:build darwin && appstore

package main

import (
	"os"
	"os/exec"

	"converzen/internal/logger"
	"converzen/internal/services"
	"converzen/pkg/ffmpeg"
)

// ffmpegInstance holds the FFmpeg instance for App Store builds (when system FFmpeg is available)
var ffmpegInstance *ffmpeg.FFmpeg

// activeBackend tracks whether we're using ffmpeg or avfoundation
var activeBackend = "avfoundation"

// findSystemFFmpeg looks for FFmpeg in common system locations
func findSystemFFmpeg() string {
	candidates := []string{
		"/usr/local/bin/ffmpeg",
		"/opt/homebrew/bin/ffmpeg",
		"/usr/bin/ffmpeg",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check if ffmpeg is in PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path
	}

	return ""
}

// initVideoConverter initializes the video converter for App Store builds
// Prioritizes system FFmpeg if available, falls back to AVFoundation
func (a *App) initVideoConverter(log *logger.Logger) services.Converter {
	// First, try to use system FFmpeg if available
	ffmpegPath := findSystemFFmpeg()
	if ffmpegPath != "" {
		ffmpegInstance = ffmpeg.New(ffmpegPath, log)
		if ffmpegInstance.IsAvailable() {
			if version, err := ffmpegInstance.GetVersion(); err == nil {
				log.Info("app", "Using system FFmpeg for video conversion: %s", version)
			}
			activeBackend = "ffmpeg"
			return services.NewFFmpegVideoConverter(ffmpegInstance, log)
		}
	}

	// Fall back to AVFoundation
	log.Info("app", "Using AVFoundation for video conversion (App Store build, no system FFmpeg found)")
	activeBackend = "avfoundation"
	// Pass nil for ffmpeg parameter - AVFoundation doesn't need it
	return services.NewVideoConverter(nil, log)
}

// isFFmpegAvailable returns true - either FFmpeg or AVFoundation is available
func (a *App) isFFmpegAvailable() bool {
	return true // Either FFmpeg or AVFoundation is always available on macOS
}

// getConverterBackend returns the name of the converter backend currently in use
func (a *App) getConverterBackend() string {
	return activeBackend
}

// getConverterVersion returns the version of the converter backend
func (a *App) getConverterVersion() (string, error) {
	if activeBackend == "ffmpeg" && ffmpegInstance != nil {
		return ffmpegInstance.GetVersion()
	}
	// AVFoundation doesn't have a version string like FFmpeg
	return "AVFoundation (macOS native)", nil
}
