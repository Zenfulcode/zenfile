package services

import (
	"strconv"

	"converzen/internal/logger"
	"converzen/internal/models"
	"converzen/internal/repository"
)

// settingsServiceImpl implements SettingsService
type settingsServiceImpl struct {
	repo repository.SettingsRepository
	log  *logger.ComponentLogger
}

// NewSettingsService creates a new SettingsService
func NewSettingsService(repo repository.SettingsRepository, log *logger.Logger) SettingsService {
	return &settingsServiceImpl{
		repo: repo,
		log:  log.WithComponent("settings-service"),
	}
}

// GetSettings returns the current user settings
func (s *settingsServiceImpl) GetSettings() (*models.UserSettings, error) {
	s.log.Debug("Getting all user settings")

	settings := models.DefaultUserSettings()

	// Get last output directory
	if setting, err := s.repo.Get(models.SettingLastOutputDir); err == nil && setting != nil {
		settings.LastOutputDirectory = setting.Value
	}

	// Get default naming mode
	if setting, err := s.repo.Get(models.SettingDefaultNaming); err == nil && setting != nil {
		settings.DefaultNamingMode = models.FileNamingMode(setting.Value)
	}

	// Get default make copies setting
	if setting, err := s.repo.Get(models.SettingDefaultMakeCopy); err == nil && setting != nil {
		settings.DefaultMakeCopies = setting.Value == "true"
	}

	// Get theme
	if setting, err := s.repo.Get(models.SettingTheme); err == nil && setting != nil {
		settings.Theme = setting.Value
	}

	return &settings, nil
}

// SaveSettings saves user settings
func (s *settingsServiceImpl) SaveSettings(settings models.UserSettings) error {
	s.log.Info("Saving user settings")

	if err := s.repo.Set(models.SettingLastOutputDir, settings.LastOutputDirectory); err != nil {
		return err
	}

	if err := s.repo.Set(models.SettingDefaultNaming, string(settings.DefaultNamingMode)); err != nil {
		return err
	}

	if err := s.repo.Set(models.SettingDefaultMakeCopy, strconv.FormatBool(settings.DefaultMakeCopies)); err != nil {
		return err
	}

	if err := s.repo.Set(models.SettingTheme, settings.Theme); err != nil {
		return err
	}

	s.log.Info("User settings saved successfully")
	return nil
}

// GetSetting retrieves a single setting value
func (s *settingsServiceImpl) GetSetting(key string) (string, error) {
	setting, err := s.repo.Get(key)
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "", nil
	}
	return setting.Value, nil
}

// SetSetting sets a single setting value
func (s *settingsServiceImpl) SetSetting(key, value string) error {
	return s.repo.Set(key, value)
}
