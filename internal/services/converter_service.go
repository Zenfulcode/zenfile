package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"converzen/internal/logger"
	"converzen/internal/models"
	"converzen/internal/repository"
)

// conversionServiceImpl orchestrates file conversions
type conversionServiceImpl struct {
	fileService    FileService
	videoConverter Converter
	imageConverter Converter
	repo           repository.ConversionRepository
	log            *logger.ComponentLogger

	// Active conversions tracking
	activeConversions map[uint]context.CancelFunc
	mu                sync.Mutex
}

// NewConversionService creates a new ConversionService
func NewConversionService(
	fileService FileService,
	videoConverter Converter,
	imageConverter Converter,
	repo repository.ConversionRepository,
	log *logger.Logger,
) ConversionService {
	return &conversionServiceImpl{
		fileService:       fileService,
		videoConverter:    videoConverter,
		imageConverter:    imageConverter,
		repo:              repo,
		log:               log.WithComponent("conversion-service"),
		activeConversions: make(map[uint]context.CancelFunc),
	}
}

// ConvertFile converts a single file
func (s *conversionServiceImpl) ConvertFile(job models.ConversionJob) (*models.ConversionResult, error) {
	s.log.Info("Converting file: %s", job.InputPath)

	// Get file info to determine converter
	fileInfo, err := s.fileService.GetFileInfo(job.InputPath)
	if err != nil {
		return nil, err
	}

	// Create database record
	now := time.Now()
	conversion := &models.Conversion{
		InputPath:    job.InputPath,
		OutputPath:   job.OutputPath,
		InputFormat:  fileInfo.Extension,
		OutputFormat: job.OutputFormat,
		FileType:     fileInfo.Type,
		FileSize:     fileInfo.Size,
		Status:       models.StatusProcessing,
		StartedAt:    &now,
	}

	if err := s.repo.Create(conversion); err != nil {
		s.log.Error("Failed to create conversion record: %v", err)
	}

	// Select appropriate converter
	var converter Converter
	switch fileInfo.Type {
	case models.FileTypeVideo:
		converter = s.videoConverter
	case models.FileTypeImage:
		converter = s.imageConverter
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileInfo.Type)
	}

	// Perform conversion
	result, err := converter.Convert(job, func(progress float64) {
		conversion.Progress = progress
		s.repo.Update(conversion)
	})

	// Update database record
	completedAt := time.Now()
	conversion.CompletedAt = &completedAt

	if err != nil {
		conversion.Status = models.StatusFailed
		conversion.ErrorMessage = err.Error()
	} else {
		conversion.Status = models.StatusCompleted
		conversion.OutputSize = result.OutputSize
	}

	if updateErr := s.repo.Update(conversion); updateErr != nil {
		s.log.Error("Failed to update conversion record: %v", updateErr)
	}

	return result, err
}

// ConvertBatch converts multiple files
func (s *conversionServiceImpl) ConvertBatch(request models.BatchConversionRequest, progressCallback func(progress models.ConversionProgress)) (*models.BatchConversionResult, error) {
	s.log.Info("Starting batch conversion of %d files", len(request.Files))
	startTime := time.Now()

	result := &models.BatchConversionResult{
		TotalFiles: len(request.Files),
		Results:    make([]models.ConversionResult, 0, len(request.Files)),
	}

	for i, inputPath := range request.Files {
		// Validate file exists
		if _, err := s.fileService.GetFileInfo(inputPath); err != nil {
			result.Results = append(result.Results, models.ConversionResult{
				InputPath:    inputPath,
				ErrorMessage: err.Error(),
			})
			result.FailCount++
			continue
		}

		// Generate output path
		var customName string
		if request.NamingMode == models.NamingModeCustom && i < len(request.CustomNames) {
			customName = request.CustomNames[i]
		}
		outputPath := s.fileService.GenerateOutputPath(
			inputPath,
			request.OutputDirectory,
			request.OutputFormat,
			request.NamingMode,
			customName,
		)

		// Create conversion job
		job := models.ConversionJob{
			InputPath:       inputPath,
			OutputPath:      outputPath,
			OutputFormat:    request.OutputFormat,
			OverwriteOutput: !request.MakeCopies,
		}

		// For copies, we always create new files, so allow overwrite if needed
		if request.MakeCopies {
			job.OverwriteOutput = true
		}

		// Convert file
		convResult, err := s.ConvertFile(job)

		if err != nil {
			result.Results = append(result.Results, models.ConversionResult{
				InputPath:    inputPath,
				OutputPath:   outputPath,
				ErrorMessage: err.Error(),
			})
			result.FailCount++
		} else {
			result.Results = append(result.Results, *convResult)
			result.SuccessCount++
		}

		// Report progress
		if progressCallback != nil {
			progressCallback(models.ConversionProgress{
				InputPath: inputPath,
				Progress:  float64(i+1) / float64(len(request.Files)) * 100,
				Status:    string(models.StatusCompleted),
			})
		}

		s.log.Debug("Batch progress: %d/%d files completed", i+1, len(request.Files))
	}

	result.TotalDuration = time.Since(startTime).Milliseconds()

	s.log.Info("Batch conversion completed: %d success, %d failed, %dms total",
		result.SuccessCount, result.FailCount, result.TotalDuration)

	return result, nil
}

// CancelConversion cancels an ongoing conversion
func (s *conversionServiceImpl) CancelConversion(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.activeConversions[id]; exists {
		cancel()
		delete(s.activeConversions, id)

		// Update database record
		conversion, err := s.repo.GetByID(id)
		if err == nil && conversion != nil {
			conversion.Status = models.StatusCancelled
			s.repo.Update(conversion)
		}

		s.log.Info("Conversion %d cancelled", id)
		return nil
	}

	return fmt.Errorf("conversion %d not found or already completed", id)
}

// GetConversionHistory retrieves conversion history
func (s *conversionServiceImpl) GetConversionHistory(limit int) ([]models.Conversion, error) {
	return s.repo.GetHistory(limit)
}
