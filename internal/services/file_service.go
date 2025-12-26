package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"converzen/internal/logger"
	"converzen/internal/models"
)

// fileServiceImpl implements FileService
type fileServiceImpl struct {
	log *logger.ComponentLogger
}

// NewFileService creates a new FileService instance
func NewFileService(log *logger.Logger) FileService {
	return &fileServiceImpl{
		log: log.WithComponent("file-service"),
	}
}

// GetFileInfo returns information about a file
func (s *fileServiceImpl) GetFileInfo(path string) (*models.FileInfo, error) {
	s.log.Debug("Getting file info for: %s", path)

	// Clean the path
	path = filepath.Clean(path)

	// Check if file exists
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.log.Error("File not found: %s", path)
			return nil, fmt.Errorf("file not found: %s", path)
		}
		s.log.Error("Failed to stat file: %s, error: %v", path, err)
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// Check if it's a directory
	if stat.IsDir() {
		s.log.Error("Path is a directory, not a file: %s", path)
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(path))
	fileType := models.GetFileType(ext)

	if fileType == models.FileTypeUnknown {
		s.log.Warn("Unknown file type for: %s (extension: %s)", path, ext)
	}

	info := &models.FileInfo{
		Path:      path,
		Name:      filepath.Base(path),
		Extension: ext,
		Size:      stat.Size(),
		Type:      fileType,
	}

	s.log.Debug("File info retrieved: %s (type: %s, size: %d bytes)", info.Name, info.Type, info.Size)
	return info, nil
}

// ValidateFiles validates a list of file paths and returns their info
func (s *fileServiceImpl) ValidateFiles(paths []string) ([]models.FileInfo, error) {
	s.log.Info("Validating %d files", len(paths))

	if len(paths) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	var files []models.FileInfo
	var errors []string
	var firstType models.FileType

	for i, path := range paths {
		info, err := s.GetFileInfo(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(path), err))
			continue
		}

		if info.Type == models.FileTypeUnknown {
			errors = append(errors, fmt.Sprintf("%s: unsupported file format", info.Name))
			continue
		}

		// Check that all files are of the same type
		if i == 0 {
			firstType = info.Type
		} else if info.Type != firstType {
			errors = append(errors, fmt.Sprintf("%s: mixed file types not allowed (expected %s, got %s)", info.Name, firstType, info.Type))
			continue
		}

		files = append(files, *info)
	}

	if len(errors) > 0 {
		s.log.Warn("Validation completed with %d errors: %v", len(errors), errors)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no valid files found: %s", strings.Join(errors, "; "))
	}

	s.log.Info("Validated %d files successfully (type: %s)", len(files), firstType)
	return files, nil
}

// GetOutputFormats returns available output formats for a file type
func (s *fileServiceImpl) GetOutputFormats(fileType models.FileType) []string {
	return models.GetOutputFormats(fileType)
}

// GenerateOutputPath generates the output path for a file
func (s *fileServiceImpl) GenerateOutputPath(inputPath, outputDir, outputFormat string, namingMode models.FileNamingMode, customName string) string {
	var baseName string

	switch namingMode {
	case models.NamingModeCustom:
		if customName != "" {
			baseName = customName
		} else {
			baseName = strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		}
	default: // NamingModeOriginal
		baseName = strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	}

	// Ensure format doesn't have a leading dot
	outputFormat = strings.TrimPrefix(outputFormat, ".")

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", baseName, outputFormat))

	s.log.Debug("Generated output path: %s -> %s", inputPath, outputPath)
	return outputPath
}

// FileExists checks if a file exists
func (s *fileServiceImpl) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
