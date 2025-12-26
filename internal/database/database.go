package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"converzen/internal/logger"
	"converzen/internal/models"
)

// Database wraps the GORM database connection
type Database struct {
	DB  *gorm.DB
	log *logger.ComponentLogger
}

// New creates a new database connection
func New(dbPath string, log *logger.Logger) (*Database, error) {
	componentLog := log.WithComponent("database")
	componentLog.Info("Initializing database at: %s", dbPath)

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		componentLog.Error("Failed to create database directory: %v", err)
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	// Open SQLite database
	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		componentLog.Error("Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{
		DB:  db,
		log: componentLog,
	}

	// Run migrations
	if err := database.migrate(); err != nil {
		componentLog.Error("Failed to run migrations: %v", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	componentLog.Info("Database initialized successfully")
	return database, nil
}

// migrate runs database migrations
func (d *Database) migrate() error {
	d.log.Info("Running database migrations")

	err := d.DB.AutoMigrate(
		&models.Conversion{},
		&models.Setting{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	d.log.Info("Database migrations completed")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	d.log.Info("Closing database connection")

	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
