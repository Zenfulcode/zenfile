package repository

import (
	"converzen/internal/models"
)

// ConversionRepository handles conversion history persistence
type ConversionRepository interface {
	// Create creates a new conversion record
	Create(conversion *models.Conversion) error

	// Update updates an existing conversion record
	Update(conversion *models.Conversion) error

	// GetByID retrieves a conversion by ID
	GetByID(id uint) (*models.Conversion, error)

	// GetHistory retrieves conversion history with a limit
	GetHistory(limit int) ([]models.Conversion, error)

	// GetPending retrieves all pending conversions
	GetPending() ([]models.Conversion, error)

	// Delete deletes a conversion record
	Delete(id uint) error

	// DeleteOlderThan deletes conversions older than the given number of days
	DeleteOlderThan(days int) error
}

// SettingsRepository handles settings persistence
type SettingsRepository interface {
	// Get retrieves a setting by key
	Get(key string) (*models.Setting, error)

	// Set sets a setting value
	Set(key, value string) error

	// GetAll retrieves all settings
	GetAll() ([]models.Setting, error)

	// Delete deletes a setting
	Delete(key string) error
}
