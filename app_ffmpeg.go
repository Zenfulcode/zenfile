//go:build !appstore

package main

import (
	"converzen/internal/logger"
	"converzen/internal/services"
	"converzen/pkg/ffmpeg"
)

// ffmpegInstance holds the FFmpeg instance for non-App Store builds
var ffmpegInstance *ffmpeg.FFmpeg

// initVideoConverter initializes the video converter for non-App Store builds (using FFmpeg)
func (a *App) initVideoConverter(log *logger.Logger) services.Converter {
	// Initialize FFmpeg
	ffmpegInstance = ffmpeg.New(a.config.FFmpegPath, log)
	if ffmpegInstance.IsAvailable() {
		if version, err := ffmpegInstance.GetVersion(); err == nil {
			log.Info("app", "FFmpeg version: %s", version)
		}
	} else {
		log.Warn("app", "FFmpeg not found - video conversion will not work")
	}

	return services.NewVideoConverter(ffmpegInstance, log)
}

// isFFmpegAvailable returns whether FFmpeg is available (for non-App Store builds)
func (a *App) isFFmpegAvailable() bool {
	return ffmpegInstance != nil && ffmpegInstance.IsAvailable()
}

// getConverterBackend returns the name of the converter backend
func (a *App) getConverterBackend() string {
	return "ffmpeg"
}

// getConverterVersion returns the version of the converter backend
func (a *App) getConverterVersion() (string, error) {
	if ffmpegInstance != nil {
		return ffmpegInstance.GetVersion()
	}
	return "", nil
}
