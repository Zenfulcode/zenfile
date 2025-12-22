package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"zenfile/internal/config"
	"zenfile/internal/database"
	"zenfile/internal/logger"
	"zenfile/internal/models"
	"zenfile/internal/repository"
	"zenfile/internal/services"
)

// App struct holds the application state and dependencies
type App struct {
	ctx context.Context

	// Configuration
	config *config.Config

	// Logger
	log *logger.Logger

	// Database
	db *database.Database

	// Services
	fileService       services.FileService
	conversionService services.ConversionService
	settingsService   services.SettingsService
	formatProvider    services.FormatProvider
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration
	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		return
	}
	a.config = cfg

	// Initialize logger
	logLevel := logger.INFO
	if cfg.Debug {
		logLevel = logger.DEBUG
	}
	log, err := logger.New(cfg.LogFile, logLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	a.log = log

	log.Info("app", "Starting %s v%s", cfg.AppName, cfg.Version)
	log.Debug("app", "Data directory: %s", cfg.DataDir)
	log.Debug("app", "Log file: %s", cfg.LogFile)

	// Initialize database
	db, err := database.New(cfg.DatabaseURL, log)
	if err != nil {
		log.Error("app", "Failed to initialize database: %v", err)
		return
	}
	a.db = db

	// Initialize repositories
	conversionRepo := repository.NewConversionRepository(db.DB, log)
	settingsRepo := repository.NewSettingsRepository(db.DB, log)

	// Initialize services
	a.fileService = services.NewFileService(log)
	videoConverter := a.initVideoConverter(log)
	imageConverter := services.NewImageConverter(log)
	a.conversionService = services.NewConversionService(
		a.fileService,
		videoConverter,
		imageConverter,
		conversionRepo,
		log,
	)
	a.settingsService = services.NewSettingsService(settingsRepo, log)
	a.formatProvider = services.NewFormatProvider(videoConverter, imageConverter, a.getConverterBackend())

	log.Info("app", "Application startup complete")
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.log != nil {
		a.log.Info("app", "Application shutting down")
	}

	if a.db != nil {
		a.db.Close()
	}

	if a.log != nil {
		a.log.Close()
	}
}

// SelectFiles opens a file dialog for selecting files
func (a *App) SelectFiles() ([]models.FileInfo, error) {
	a.log.Debug("app", "Opening file selection dialog")

	// Build file filters dynamically based on supported formats
	filters := a.buildFileFilters()

	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select Files to Convert",
		Filters: filters,
	})

	if err != nil {
		a.log.Error("app", "File selection error: %v", err)
		return nil, err
	}

	if len(files) == 0 {
		a.log.Debug("app", "No files selected")
		return []models.FileInfo{}, nil
	}

	// Validate the selected files
	validFiles, err := a.fileService.ValidateFiles(files)
	if err != nil {
		a.log.Error("app", "File validation error: %v", err)
		return nil, err
	}

	a.log.Info("app", "Selected %d valid files", len(validFiles))
	return validFiles, nil
}

// buildFileFilters creates file filters based on supported formats from the formatProvider
func (a *App) buildFileFilters() []runtime.FileFilter {
	videoFormats := a.formatProvider.GetSupportedVideoInputFormats()
	imageFormats := a.formatProvider.GetSupportedImageInputFormats()

	// Build pattern strings (e.g., "*.mp4;*.mov;*.m4v")
	videoPattern := buildPatternFromFormats(videoFormats)
	imagePattern := buildPatternFromFormats(imageFormats)

	var filters []runtime.FileFilter

	if videoPattern != "" {
		filters = append(filters, runtime.FileFilter{
			DisplayName: "Video Files",
			Pattern:     videoPattern,
		})
	}

	if imagePattern != "" {
		filters = append(filters, runtime.FileFilter{
			DisplayName: "Image Files",
			Pattern:     imagePattern,
		})
	}

	// Add "All Supported Files" option if we have formats
	if videoPattern != "" || imagePattern != "" {
		allPatterns := videoPattern
		if allPatterns != "" && imagePattern != "" {
			allPatterns += ";" + imagePattern
		} else if imagePattern != "" {
			allPatterns = imagePattern
		}
		filters = append(filters, runtime.FileFilter{
			DisplayName: "All Supported Files",
			Pattern:     allPatterns,
		})
	}

	return filters
}

// buildPatternFromFormats converts a slice of format strings to a file pattern
func buildPatternFromFormats(formats []string) string {
	if len(formats) == 0 {
		return ""
	}

	patterns := make([]string, len(formats))
	for i, format := range formats {
		patterns[i] = "*." + format
	}
	return strings.Join(patterns, ";")
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() (string, error) {
	a.log.Debug("app", "Opening directory selection dialog")

	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Output Directory",
	})

	if err != nil {
		a.log.Error("app", "Directory selection error: %v", err)
		return "", err
	}

	if dir != "" {
		a.log.Info("app", "Selected output directory: %s", dir)
		// Save as last used directory
		a.settingsService.SetSetting(models.SettingLastOutputDir, dir)
	}

	return dir, nil
}

// GetOutputFormats returns available output formats for a file type
func (a *App) GetOutputFormats(fileType string) []string {
	ft := models.FileType(fileType)
	formats := a.fileService.GetOutputFormats(ft)
	a.log.Debug("app", "Output formats for %s: %v", fileType, formats)
	return formats
}

// ConvertFiles converts multiple files
func (a *App) ConvertFiles(request models.BatchConversionRequest) (*models.BatchConversionResult, error) {
	a.log.Info("app", "Starting batch conversion: %d files to %s", len(request.Files), request.OutputFormat)

	result, err := a.conversionService.ConvertBatch(request, func(progress models.ConversionProgress) {
		// Emit progress event to frontend
		runtime.EventsEmit(a.ctx, "conversion:progress", progress)
	})

	if err != nil {
		a.log.Error("app", "Batch conversion error: %v", err)
		return nil, err
	}

	// Emit completion event
	runtime.EventsEmit(a.ctx, "conversion:complete", result)

	a.log.Info("app", "Batch conversion complete: %d success, %d failed",
		result.SuccessCount, result.FailCount)

	return result, nil
}

// GetConversionHistory retrieves the conversion history
func (a *App) GetConversionHistory(limit int) ([]models.Conversion, error) {
	a.log.Debug("app", "Getting conversion history (limit: %d)", limit)
	return a.conversionService.GetConversionHistory(limit)
}

// GetSettings retrieves user settings
func (a *App) GetSettings() (*models.UserSettings, error) {
	return a.settingsService.GetSettings()
}

// SaveSettings saves user settings
func (a *App) SaveSettings(settings models.UserSettings) error {
	return a.settingsService.SaveSettings(settings)
}

// CheckFFmpeg checks if the video converter backend is available
func (a *App) CheckFFmpeg() bool {
	return a.isFFmpegAvailable()
}

// GetFFmpegVersion returns the video converter backend version string
func (a *App) GetFFmpegVersion() (string, error) {
	return a.getConverterVersion()
}

// GetAppInfo returns application information
func (a *App) GetAppInfo() map[string]string {
	info := map[string]string{
		"name":             a.config.AppName,
		"version":          a.config.Version,
		"dataDir":          a.config.DataDir,
		"logFile":          a.config.LogFile,
		"converterBackend": a.getConverterBackend(),
	}

	if version, err := a.getConverterVersion(); err == nil {
		info["ffmpegVersion"] = version
	}

	return info
}

// SupportedFormatsResponse contains the supported formats for the frontend
type SupportedFormatsResponse struct {
	VideoFormats []string `json:"videoFormats"`
	ImageFormats []string `json:"imageFormats"`
	Backend      string   `json:"backend"`
}

// GetSupportedFormats returns the supported output formats for the current backend
// This allows the frontend to dynamically show only formats that the backend supports
func (a *App) GetSupportedFormats() SupportedFormatsResponse {
	return SupportedFormatsResponse{
		VideoFormats: a.formatProvider.GetSupportedVideoOutputFormats(),
		ImageFormats: a.formatProvider.GetSupportedImageOutputFormats(),
		Backend:      a.formatProvider.GetBackendName(),
	}
}

// GetSupportedVideoFormats returns the supported video output formats
func (a *App) GetSupportedVideoFormats() []string {
	return a.formatProvider.GetSupportedVideoOutputFormats()
}

// GetSupportedImageFormats returns the supported image output formats
func (a *App) GetSupportedImageFormats() []string {
	return a.formatProvider.GetSupportedImageOutputFormats()
}

// CanConvert checks if conversion from input to output format is supported
func (a *App) CanConvert(fileType string, outputFormat string) bool {
	switch models.FileType(fileType) {
	case models.FileTypeVideo:
		return a.formatProvider.CanConvertVideo(outputFormat)
	case models.FileTypeImage:
		return a.formatProvider.CanConvertImage(outputFormat)
	default:
		return false
	}
}
