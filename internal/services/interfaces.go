package services

import (
	"converzen/internal/models"
)

// FileService handles file operations
type FileService interface {
	// GetFileInfo returns information about a file
	GetFileInfo(path string) (*models.FileInfo, error)

	// ValidateFiles validates a list of file paths and returns their info
	ValidateFiles(paths []string) ([]models.FileInfo, error)

	// GetOutputFormats returns available output formats for a file type
	GetOutputFormats(fileType models.FileType) []string

	// GenerateOutputPath generates the output path for a file
	GenerateOutputPath(inputPath, outputDir, outputFormat string, namingMode models.FileNamingMode, customName string) string

	// FileExists checks if a file exists
	FileExists(path string) bool
}

// Converter handles file conversion
type Converter interface {
	// Convert converts a single file
	Convert(job models.ConversionJob, progressCallback func(progress float64)) (*models.ConversionResult, error)

	// SupportedInputFormats returns the list of supported input formats
	SupportedInputFormats() []string

	// SupportedOutputFormats returns the list of supported output formats for an input format
	SupportedOutputFormats(inputFormat string) []string

	// CanConvert checks if conversion is possible between formats
	CanConvert(inputFormat, outputFormat string) bool
}

// ConversionService orchestrates file conversions
type ConversionService interface {
	// ConvertFile converts a single file
	ConvertFile(job models.ConversionJob) (*models.ConversionResult, error)

	// ConvertBatch converts multiple files
	ConvertBatch(request models.BatchConversionRequest, progressCallback func(progress models.ConversionProgress)) (*models.BatchConversionResult, error)

	// CancelConversion cancels an ongoing conversion
	CancelConversion(id uint) error

	// GetConversionHistory retrieves conversion history
	GetConversionHistory(limit int) ([]models.Conversion, error)
}

// SettingsService handles user settings
type SettingsService interface {
	// GetSettings returns the current user settings
	GetSettings() (*models.UserSettings, error)

	// SaveSettings saves user settings
	SaveSettings(settings models.UserSettings) error

	// GetSetting retrieves a single setting value
	GetSetting(key string) (string, error)

	// SetSetting sets a single setting value
	SetSetting(key, value string) error
}
