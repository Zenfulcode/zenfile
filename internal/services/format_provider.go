package services

import "converzen/internal/models"

// FormatProvider defines the contract for getting supported formats
// This follows the Interface Segregation Principle (ISP) - clients only depend on methods they use
type FormatProvider interface {
	// GetSupportedVideoInputFormats returns the list of supported video input formats
	GetSupportedVideoInputFormats() []string

	// GetSupportedImageInputFormats returns the list of supported image input formats
	GetSupportedImageInputFormats() []string

	// GetSupportedVideoOutputFormats returns the list of supported video output formats
	GetSupportedVideoOutputFormats() []string

	// GetSupportedImageOutputFormats returns the list of supported image output formats
	GetSupportedImageOutputFormats() []string

	// GetSupportedFormats returns all supported formats for a given file type
	GetSupportedFormats(fileType models.FileType) []string

	// CanConvertVideo checks if video conversion to the specified format is supported
	CanConvertVideo(outputFormat string) bool

	// CanConvertImage checks if image conversion to the specified format is supported
	CanConvertImage(outputFormat string) bool

	// GetBackendName returns the name of the conversion backend (e.g., "ffmpeg", "avfoundation")
	GetBackendName() string
}

// formatProvider implements FormatProvider by delegating to the actual converters
// This follows the Single Responsibility Principle (SRP) - only handles format queries
type formatProvider struct {
	videoConverter Converter
	imageConverter Converter
	backendName    string
}

// NewFormatProvider creates a new FormatProvider
// This follows the Dependency Inversion Principle (DIP) - depends on Converter interface, not concrete implementations
func NewFormatProvider(videoConverter Converter, imageConverter Converter, backendName string) FormatProvider {
	return &formatProvider{
		videoConverter: videoConverter,
		imageConverter: imageConverter,
		backendName:    backendName,
	}
}

// GetSupportedVideoInputFormats returns the list of supported video input formats
func (p *formatProvider) GetSupportedVideoInputFormats() []string {
	if p.videoConverter == nil {
		return []string{}
	}
	return p.videoConverter.SupportedInputFormats()
}

// GetSupportedImageInputFormats returns the list of supported image input formats
func (p *formatProvider) GetSupportedImageInputFormats() []string {
	if p.imageConverter == nil {
		return []string{}
	}
	return p.imageConverter.SupportedInputFormats()
}

// GetSupportedVideoOutputFormats returns the list of supported video output formats
func (p *formatProvider) GetSupportedVideoOutputFormats() []string {
	if p.videoConverter == nil {
		return []string{}
	}
	return p.videoConverter.SupportedOutputFormats("")
}

// GetSupportedImageOutputFormats returns the list of supported image output formats
func (p *formatProvider) GetSupportedImageOutputFormats() []string {
	if p.imageConverter == nil {
		return []string{}
	}
	return p.imageConverter.SupportedOutputFormats("")
}

// GetSupportedFormats returns all supported formats for a given file type
func (p *formatProvider) GetSupportedFormats(fileType models.FileType) []string {
	switch fileType {
	case models.FileTypeVideo:
		return p.GetSupportedVideoOutputFormats()
	case models.FileTypeImage:
		return p.GetSupportedImageOutputFormats()
	default:
		return []string{}
	}
}

// CanConvertVideo checks if video conversion to the specified format is supported
func (p *formatProvider) CanConvertVideo(outputFormat string) bool {
	if p.videoConverter == nil {
		return false
	}
	for _, format := range p.GetSupportedVideoOutputFormats() {
		if format == outputFormat {
			return true
		}
	}
	return false
}

// CanConvertImage checks if image conversion to the specified format is supported
func (p *formatProvider) CanConvertImage(outputFormat string) bool {
	if p.imageConverter == nil {
		return false
	}
	for _, format := range p.GetSupportedImageOutputFormats() {
		if format == outputFormat {
			return true
		}
	}
	return false
}

// GetBackendName returns the name of the conversion backend
func (p *formatProvider) GetBackendName() string {
	return p.backendName
}
