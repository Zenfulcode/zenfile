package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"converzen/internal/logger"
	"converzen/internal/models"
	"converzen/pkg/ffmpeg"
)

// ffmpegVideoConverter handles video file conversion using FFmpeg
// This is available in all builds and can be used when system FFmpeg is available
type ffmpegVideoConverter struct {
	ffmpeg *ffmpeg.FFmpeg
	log    *logger.ComponentLogger
}

// NewFFmpegVideoConverter creates a new video converter using FFmpeg
// This is used when system FFmpeg is available in the App Store build
func NewFFmpegVideoConverter(ff *ffmpeg.FFmpeg, log *logger.Logger) Converter {
	return &ffmpegVideoConverter{
		ffmpeg: ff,
		log:    log.WithComponent("ffmpeg-video-converter"),
	}
}

// Convert converts a video file to another format using FFmpeg
func (c *ffmpegVideoConverter) Convert(job models.ConversionJob, progressCallback func(progress float64)) (*models.ConversionResult, error) {
	c.log.Info("Starting FFmpeg video conversion: %s -> %s", job.InputPath, job.OutputPath)
	startTime := time.Now()

	result := &models.ConversionResult{
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
	}

	// Validate input file exists
	if _, err := os.Stat(job.InputPath); os.IsNotExist(err) {
		result.ErrorMessage = fmt.Sprintf("Input file not found: %s", job.InputPath)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	// Check output directory exists
	outputDir := filepath.Dir(job.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create output directory: %v", err)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	// Check if output file already exists
	if !job.OverwriteOutput {
		if _, err := os.Stat(job.OutputPath); err == nil {
			result.ErrorMessage = fmt.Sprintf("Output file already exists: %s", job.OutputPath)
			c.log.Error("%s", result.ErrorMessage)
			return result, fmt.Errorf("%s", result.ErrorMessage)
		}
	}

	// Get output format
	outputFormat := strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")

	// Handle GIF conversion separately
	if outputFormat == "gif" {
		err := c.ffmpeg.ConvertToGif(context.Background(), job.InputPath, job.OutputPath, job.OverwriteOutput, progressCallback)
		if err != nil {
			result.ErrorMessage = err.Error()
			c.log.Error("GIF conversion failed: %v", err)
			return result, err
		}
	} else {
		// Get default codecs for the format
		videoCodec, audioCodec := ffmpeg.GetDefaultCodec(outputFormat)

		opts := ffmpeg.ConvertOptions{
			InputPath:  job.InputPath,
			OutputPath: job.OutputPath,
			Overwrite:  job.OverwriteOutput,
			VideoCodec: videoCodec,
			AudioCodec: audioCodec,
		}

		err := c.ffmpeg.Convert(context.Background(), opts, progressCallback)
		if err != nil {
			result.ErrorMessage = err.Error()
			c.log.Error("Video conversion failed: %v", err)
			return result, err
		}
	}

	// Get output file size
	if stat, err := os.Stat(job.OutputPath); err == nil {
		result.OutputSize = stat.Size()
	}

	result.Success = true
	result.Duration = time.Since(startTime).Milliseconds()

	c.log.Info("FFmpeg video conversion completed in %dms: %s", result.Duration, job.OutputPath)
	return result, nil
}

// SupportedInputFormats returns the list of supported input video formats
func (c *ffmpegVideoConverter) SupportedInputFormats() []string {
	formats := make([]string, 0, len(models.VideoFormats))
	for format := range models.VideoFormats {
		formats = append(formats, strings.TrimPrefix(format, "."))
	}
	return formats
}

// SupportedOutputFormats returns the list of supported output formats for video
func (c *ffmpegVideoConverter) SupportedOutputFormats(inputFormat string) []string {
	return models.VideoOutputFormats
}

// CanConvert checks if conversion is possible between formats
func (c *ffmpegVideoConverter) CanConvert(inputFormat, outputFormat string) bool {
	inputFormat = strings.ToLower(strings.TrimPrefix(inputFormat, "."))
	outputFormat = strings.ToLower(strings.TrimPrefix(outputFormat, "."))

	// Check input is a video format
	if !models.VideoFormats["."+inputFormat] {
		return false
	}

	// Check output is a valid video output format
	for _, format := range models.VideoOutputFormats {
		if format == outputFormat {
			return true
		}
	}

	return false
}
