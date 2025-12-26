package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"converzen/internal/logger"
	"converzen/internal/models"
)

// conversionRepoImpl implements ConversionRepository
type conversionRepoImpl struct {
	db  *gorm.DB
	log *logger.ComponentLogger
}

// NewConversionRepository creates a new ConversionRepository
func NewConversionRepository(db *gorm.DB, log *logger.Logger) ConversionRepository {
	return &conversionRepoImpl{
		db:  db,
		log: log.WithComponent("conversion-repo"),
	}
}

// Create creates a new conversion record
func (r *conversionRepoImpl) Create(conversion *models.Conversion) error {
	r.log.Debug("Creating conversion record for: %s", conversion.InputPath)

	if err := r.db.Create(conversion).Error; err != nil {
		r.log.Error("Failed to create conversion record: %v", err)
		return fmt.Errorf("failed to create conversion record: %w", err)
	}

	r.log.Debug("Created conversion record with ID: %d", conversion.ID)
	return nil
}

// Update updates an existing conversion record
func (r *conversionRepoImpl) Update(conversion *models.Conversion) error {
	r.log.Debug("Updating conversion record ID: %d", conversion.ID)

	if err := r.db.Save(conversion).Error; err != nil {
		r.log.Error("Failed to update conversion record: %v", err)
		return fmt.Errorf("failed to update conversion record: %w", err)
	}

	return nil
}

// GetByID retrieves a conversion by ID
func (r *conversionRepoImpl) GetByID(id uint) (*models.Conversion, error) {
	r.log.Debug("Getting conversion by ID: %d", id)

	var conversion models.Conversion
	if err := r.db.First(&conversion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Error("Failed to get conversion by ID: %v", err)
		return nil, fmt.Errorf("failed to get conversion: %w", err)
	}

	return &conversion, nil
}

// GetHistory retrieves conversion history with a limit
func (r *conversionRepoImpl) GetHistory(limit int) ([]models.Conversion, error) {
	r.log.Debug("Getting conversion history (limit: %d)", limit)

	var conversions []models.Conversion
	query := r.db.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&conversions).Error; err != nil {
		r.log.Error("Failed to get conversion history: %v", err)
		return nil, fmt.Errorf("failed to get conversion history: %w", err)
	}

	r.log.Debug("Retrieved %d conversion records", len(conversions))
	return conversions, nil
}

// GetPending retrieves all pending conversions
func (r *conversionRepoImpl) GetPending() ([]models.Conversion, error) {
	r.log.Debug("Getting pending conversions")

	var conversions []models.Conversion
	if err := r.db.Where("status = ?", models.StatusPending).Find(&conversions).Error; err != nil {
		r.log.Error("Failed to get pending conversions: %v", err)
		return nil, fmt.Errorf("failed to get pending conversions: %w", err)
	}

	return conversions, nil
}

// Delete deletes a conversion record
func (r *conversionRepoImpl) Delete(id uint) error {
	r.log.Debug("Deleting conversion record ID: %d", id)

	if err := r.db.Delete(&models.Conversion{}, id).Error; err != nil {
		r.log.Error("Failed to delete conversion record: %v", err)
		return fmt.Errorf("failed to delete conversion record: %w", err)
	}

	return nil
}

// DeleteOlderThan deletes conversions older than the given number of days
func (r *conversionRepoImpl) DeleteOlderThan(days int) error {
	r.log.Info("Deleting conversions older than %d days", days)

	cutoff := time.Now().AddDate(0, 0, -days)
	result := r.db.Where("created_at < ?", cutoff).Delete(&models.Conversion{})

	if result.Error != nil {
		r.log.Error("Failed to delete old conversions: %v", result.Error)
		return fmt.Errorf("failed to delete old conversions: %w", result.Error)
	}

	r.log.Info("Deleted %d old conversion records", result.RowsAffected)
	return nil
}
