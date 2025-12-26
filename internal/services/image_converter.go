package services

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"

	"converzen/internal/logger"
	"converzen/internal/models"
)

// imageConverter handles image file conversion
type imageConverter struct {
	log *logger.ComponentLogger
}

// NewImageConverter creates a new image converter
func NewImageConverter(log *logger.Logger) Converter {
	return &imageConverter{
		log: log.WithComponent("image-converter"),
	}
}

// Convert converts an image file to another format
func (c *imageConverter) Convert(job models.ConversionJob, progressCallback func(progress float64)) (*models.ConversionResult, error) {
	c.log.Info("Starting image conversion: %s -> %s", job.InputPath, job.OutputPath)
	startTime := time.Now()

	result := &models.ConversionResult{
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
	}

	// Report initial progress
	if progressCallback != nil {
		progressCallback(10)
	}

	// Open input file
	inputFile, err := os.Open(job.InputPath)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to open input file: %v", err)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}
	defer inputFile.Close()

	if progressCallback != nil {
		progressCallback(20)
	}

	// Decode input image
	var img image.Image
	inputFormat := strings.ToLower(strings.TrimPrefix(filepath.Ext(job.InputPath), "."))

	switch inputFormat {
	case "png":
		img, err = png.Decode(inputFile)
	case "jpg", "jpeg":
		img, err = jpeg.Decode(inputFile)
	case "gif":
		img, err = gif.Decode(inputFile)
	case "webp":
		img, err = webp.Decode(inputFile)
	case "bmp":
		img, err = bmp.Decode(inputFile)
	case "tiff", "tif":
		img, err = tiff.Decode(inputFile)
	default:
		// Try generic decode
		img, _, err = image.Decode(inputFile)
	}

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to decode image: %v", err)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	if progressCallback != nil {
		progressCallback(50)
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

	// Create output file
	outputFile, err := os.Create(job.OutputPath)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create output file: %v", err)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}
	defer outputFile.Close()

	if progressCallback != nil {
		progressCallback(70)
	}

	// Encode to output format
	outputFormat := strings.ToLower(strings.TrimPrefix(filepath.Ext(job.OutputPath), "."))

	switch outputFormat {
	case "png":
		err = png.Encode(outputFile, img)
	case "jpg", "jpeg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 90})
	case "gif":
		err = gif.Encode(outputFile, img, nil)
	case "bmp":
		err = bmp.Encode(outputFile, img)
	case "tiff", "tif":
		err = tiff.Encode(outputFile, img, nil)
	case "webp":
		// WebP encoding requires a different library, fall back to PNG for now
		// In production, use github.com/chai2010/webp or similar
		result.ErrorMessage = "WebP encoding is not supported as output format"
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	default:
		result.ErrorMessage = fmt.Sprintf("Unsupported output format: %s", outputFormat)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to encode image: %v", err)
		c.log.Error("%s", result.ErrorMessage)
		return result, fmt.Errorf("%s", result.ErrorMessage)
	}

	if progressCallback != nil {
		progressCallback(100)
	}

	// Get output file size
	if stat, err := os.Stat(job.OutputPath); err == nil {
		result.OutputSize = stat.Size()
	}

	result.Success = true
	result.Duration = time.Since(startTime).Milliseconds()

	c.log.Info("Image conversion completed in %dms: %s", result.Duration, job.OutputPath)
	return result, nil
}

// SupportedInputFormats returns the list of supported input image formats
func (c *imageConverter) SupportedInputFormats() []string {
	formats := make([]string, 0, len(models.ImageFormats))
	for format := range models.ImageFormats {
		formats = append(formats, strings.TrimPrefix(format, "."))
	}
	return formats
}

// SupportedOutputFormats returns the list of supported output formats for images
func (c *imageConverter) SupportedOutputFormats(inputFormat string) []string {
	// Exclude webp from output since we can't encode it without external library
	formats := make([]string, 0)
	for _, f := range models.ImageOutputFormats {
		if f != "webp" {
			formats = append(formats, f)
		}
	}
	return formats
}

// CanConvert checks if conversion is possible between formats
func (c *imageConverter) CanConvert(inputFormat, outputFormat string) bool {
	inputFormat = strings.ToLower(strings.TrimPrefix(inputFormat, "."))
	outputFormat = strings.ToLower(strings.TrimPrefix(outputFormat, "."))

	// Check input is an image format
	if !models.ImageFormats["."+inputFormat] {
		return false
	}

	// Check output is a valid image output format (excluding webp)
	validOutputs := []string{"png", "jpg", "jpeg", "gif", "bmp", "tiff", "tif"}
	for _, format := range validOutputs {
		if format == outputFormat {
			return true
		}
	}

	return false
}
