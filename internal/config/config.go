package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the application configuration
type Config struct {
	AppName     string
	Version     string
	DataDir     string
	LogDir      string
	LogFile     string
	DatabaseDir string
	DatabaseURL string
	FFmpegPath  string
	Debug       bool
}

// New creates a new Config with default values
func New() (*Config, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(dataDir, "logs")
	dbDir := filepath.Join(dataDir, "data")

	// Create directories if they don't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	return &Config{
		AppName:     "Converzen",
		Version:     "1.0.0",
		DataDir:     dataDir,
		LogDir:      logDir,
		LogFile:     filepath.Join(logDir, "app.log"),
		DatabaseDir: dbDir,
		DatabaseURL: filepath.Join(dbDir, "converzen.db"),
		FFmpegPath:  findFFmpeg(dataDir),
		Debug:       os.Getenv("DEBUG") == "true",
	}, nil
}

// getDataDir returns the appropriate data directory for the current OS
func getDataDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("APPDATA")
		}
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(homeDir, "Library", "Application Support")
	default: // Linux and others
		baseDir = os.Getenv("XDG_DATA_HOME")
		if baseDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(homeDir, ".local", "share")
		}
	}

	return filepath.Join(baseDir, "Converzen"), nil
}
