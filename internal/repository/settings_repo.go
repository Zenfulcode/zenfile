package repository

import (
	"fmt"

	"gorm.io/gorm"

	"converzen/internal/logger"
	"converzen/internal/models"
)

// settingsRepoImpl implements SettingsRepository
type settingsRepoImpl struct {
	db  *gorm.DB
	log *logger.ComponentLogger
}

// NewSettingsRepository creates a new SettingsRepository
func NewSettingsRepository(db *gorm.DB, log *logger.Logger) SettingsRepository {
	return &settingsRepoImpl{
		db:  db,
		log: log.WithComponent("settings-repo"),
	}
}

// Get retrieves a setting by key
func (r *settingsRepoImpl) Get(key string) (*models.Setting, error) {
	r.log.Debug("Getting setting: %s", key)

	var setting models.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Error("Failed to get setting: %v", err)
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	return &setting, nil
}

// Set sets a setting value (creates or updates)
func (r *settingsRepoImpl) Set(key, value string) error {
	r.log.Debug("Setting: %s = %s", key, value)

	// Try to find existing setting
	var setting models.Setting
	result := r.db.Where("key = ?", key).First(&setting)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new setting
		setting = models.Setting{
			Key:   key,
			Value: value,
		}
		if err := r.db.Create(&setting).Error; err != nil {
			r.log.Error("Failed to create setting: %v", err)
			return fmt.Errorf("failed to create setting: %w", err)
		}
	} else if result.Error != nil {
		r.log.Error("Failed to query setting: %v", result.Error)
		return fmt.Errorf("failed to query setting: %w", result.Error)
	} else {
		// Update existing setting
		setting.Value = value
		if err := r.db.Save(&setting).Error; err != nil {
			r.log.Error("Failed to update setting: %v", err)
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	return nil
}

// GetAll retrieves all settings
func (r *settingsRepoImpl) GetAll() ([]models.Setting, error) {
	r.log.Debug("Getting all settings")

	var settings []models.Setting
	if err := r.db.Find(&settings).Error; err != nil {
		r.log.Error("Failed to get all settings: %v", err)
		return nil, fmt.Errorf("failed to get all settings: %w", err)
	}

	return settings, nil
}

// Delete deletes a setting
func (r *settingsRepoImpl) Delete(key string) error {
	r.log.Debug("Deleting setting: %s", key)

	if err := r.db.Where("key = ?", key).Delete(&models.Setting{}).Error; err != nil {
		r.log.Error("Failed to delete setting: %v", err)
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	return nil
}
