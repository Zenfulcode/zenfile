package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"converzen/internal/logger"
)

// FFmpeg wraps FFmpeg command execution
type FFmpeg struct {
	path string
	log  *logger.ComponentLogger
}

// New creates a new FFmpeg instance
func New(ffmpegPath string, log *logger.Logger) *FFmpeg {
	return &FFmpeg{
		path: ffmpegPath,
		log:  log.WithComponent("ffmpeg"),
	}
}

// ConvertOptions holds options for video/audio conversion
type ConvertOptions struct {
	InputPath  string
	OutputPath string
	Overwrite  bool

	// Video options
	VideoCodec   string
	VideoBitrate string
	Resolution   string
	FrameRate    int

	// Audio options
	AudioCodec   string
	AudioBitrate string
	SampleRate   int
}

// ProgressCallback is called with progress updates (0-100)
type ProgressCallback func(progress float64)

// IsAvailable checks if FFmpeg is available on the system
func (f *FFmpeg) IsAvailable() bool {
	cmd := exec.Command(f.path, "-version")
	err := cmd.Run()
	available := err == nil
	if available {
		f.log.Info("FFmpeg is available at: %s", f.path)
	} else {
		f.log.Error("FFmpeg is not available: %v", err)
	}
	return available
}

// GetVersion returns the FFmpeg version
func (f *FFmpeg) GetVersion() (string, error) {
	cmd := exec.Command(f.path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get FFmpeg version: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", fmt.Errorf("empty version output")
}

// GetDuration returns the duration of a media file in seconds
func (f *FFmpeg) GetDuration(inputPath string) (float64, error) {
	f.log.Debug("Getting duration for: %s", inputPath)

	cmd := exec.Command(f.path, "-i", inputPath, "-hide_banner")
	output, _ := cmd.CombinedOutput() // FFmpeg writes info to stderr

	// Parse duration from output: Duration: 00:01:30.50
	re := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) != 5 {
		return 0, fmt.Errorf("could not parse duration from FFmpeg output")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	centiseconds, _ := strconv.Atoi(matches[4])

	duration := float64(hours)*3600 + float64(minutes)*60 + float64(seconds) + float64(centiseconds)/100

	f.log.Debug("Duration: %.2f seconds", duration)
	return duration, nil
}

// Convert performs a video/audio conversion
func (f *FFmpeg) Convert(ctx context.Context, opts ConvertOptions, progressCallback ProgressCallback) error {
	f.log.Info("Starting conversion: %s -> %s", opts.InputPath, opts.OutputPath)

	// Get input duration for progress calculation
	duration, err := f.GetDuration(opts.InputPath)
	if err != nil {
		f.log.Warn("Could not get duration, progress will not be reported: %v", err)
		duration = 0
	}

	// Build FFmpeg command
	args := []string{"-i", opts.InputPath}

	// Add overwrite flag
	if opts.Overwrite {
		args = append([]string{"-y"}, args...)
	} else {
		args = append([]string{"-n"}, args...)
	}

	// Add video options
	if opts.VideoCodec != "" {
		args = append(args, "-c:v", opts.VideoCodec)
	}
	if opts.VideoBitrate != "" {
		args = append(args, "-b:v", opts.VideoBitrate)
	}
	if opts.Resolution != "" {
		args = append(args, "-s", opts.Resolution)
	}
	if opts.FrameRate > 0 {
		args = append(args, "-r", strconv.Itoa(opts.FrameRate))
	}

	// Add audio options
	if opts.AudioCodec != "" {
		args = append(args, "-c:a", opts.AudioCodec)
	}
	if opts.AudioBitrate != "" {
		args = append(args, "-b:a", opts.AudioBitrate)
	}
	if opts.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(opts.SampleRate))
	}

	// Add progress reporting
	args = append(args, "-progress", "pipe:1", "-nostats")

	// Add output path
	args = append(args, opts.OutputPath)

	f.log.Debug("FFmpeg command: %s %s", f.path, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, f.path, args...)

	// Get stdout for progress parsing
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		f.log.Error("Failed to start FFmpeg: %v", err)
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Parse progress from stdout
	if duration > 0 && progressCallback != nil {
		go func() {
			scanner := bufio.NewScanner(stdout)
			timeRegex := regexp.MustCompile(`out_time_ms=(\d+)`)

			for scanner.Scan() {
				line := scanner.Text()
				if matches := timeRegex.FindStringSubmatch(line); len(matches) == 2 {
					timeMs, _ := strconv.ParseInt(matches[1], 10, 64)
					currentTime := float64(timeMs) / 1000000 // Convert microseconds to seconds
					progress := (currentTime / duration) * 100
					if progress > 100 {
						progress = 100
					}
					progressCallback(progress)
				}
			}
		}()
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		f.log.Error("FFmpeg conversion failed: %v", err)
		return fmt.Errorf("conversion failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback(100)
	}

	f.log.Info("Conversion completed successfully")
	return nil
}

// ConvertToGif converts a video to GIF
func (f *FFmpeg) ConvertToGif(ctx context.Context, inputPath, outputPath string, overwrite bool, progressCallback ProgressCallback) error {
	f.log.Info("Converting to GIF: %s -> %s", inputPath, outputPath)

	// Get input duration for progress calculation
	duration, _ := f.GetDuration(inputPath)

	// Build command with palette generation for better quality
	args := []string{}
	if overwrite {
		args = append(args, "-y")
	}
	args = append(args,
		"-i", inputPath,
		"-vf", "fps=10,scale=480:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
		"-loop", "0",
		"-progress", "pipe:1", "-nostats",
		outputPath,
	)

	f.log.Debug("FFmpeg GIF command: %s %s", f.path, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, f.path, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Parse progress
	if duration > 0 && progressCallback != nil {
		go func() {
			scanner := bufio.NewScanner(stdout)
			timeRegex := regexp.MustCompile(`out_time_ms=(\d+)`)

			for scanner.Scan() {
				line := scanner.Text()
				if matches := timeRegex.FindStringSubmatch(line); len(matches) == 2 {
					timeMs, _ := strconv.ParseInt(matches[1], 10, 64)
					currentTime := float64(timeMs) / 1000000
					progress := (currentTime / duration) * 100
					if progress > 100 {
						progress = 100
					}
					progressCallback(progress)
				}
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("GIF conversion failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback(100)
	}

	f.log.Info("GIF conversion completed")
	return nil
}

// GetDefaultCodec returns the default codec for a given output format
func GetDefaultCodec(format string) (videoCodec, audioCodec string) {
	format = strings.TrimPrefix(strings.ToLower(format), ".")

	switch format {
	case "mp4":
		return "libx264", "aac"
	case "webm":
		return "libvpx-vp9", "libopus"
	case "avi":
		return "mpeg4", "mp3"
	case "mkv":
		return "libx264", "aac"
	case "mov":
		return "libx264", "aac"
	default:
		return "", ""
	}
}

// Probe holds media file information
type Probe struct {
	Duration   time.Duration
	Width      int
	Height     int
	VideoCodec string
	AudioCodec string
	Bitrate    int64
}

// ProbeFile probes a media file for information
func (f *FFmpeg) ProbeFile(inputPath string) (*Probe, error) {
	f.log.Debug("Probing file: %s", inputPath)

	// Use ffprobe if available, otherwise parse ffmpeg output
	cmd := exec.Command(f.path, "-i", inputPath, "-hide_banner")
	output, _ := cmd.CombinedOutput()

	probe := &Probe{}

	// Parse duration
	durationRe := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	if matches := durationRe.FindStringSubmatch(string(output)); len(matches) == 5 {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		seconds, _ := strconv.Atoi(matches[3])
		probe.Duration = time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second
	}

	// Parse resolution
	resRe := regexp.MustCompile(`(\d{2,5})x(\d{2,5})`)
	if matches := resRe.FindStringSubmatch(string(output)); len(matches) == 3 {
		probe.Width, _ = strconv.Atoi(matches[1])
		probe.Height, _ = strconv.Atoi(matches[2])
	}

	return probe, nil
}
