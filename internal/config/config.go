package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds runtime configuration for the Atlas application.
type Config struct {
	Host           string
	Port           int
	DataDir        string
	DatabasePath   string
	LogLevel       string
	AllowedOrigins string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	dataDir := getEnv("ATLAS_DATA_DIR", "./data")
	dbPath := getEnv("ATLAS_DATABASE_PATH", filepath.Join(dataDir, "atlas.duckdb"))
	port, err := strconv.Atoi(getEnv("ATLAS_PORT", "8080"))
	if err != nil {
		port = 8080
	}

	return &Config{
		Host:           getEnv("ATLAS_HOST", "0.0.0.0"),
		Port:           port,
		DataDir:        dataDir,
		DatabasePath:   dbPath,
		LogLevel:       getEnv("ATLAS_LOG_LEVEL", "INFO"),
		AllowedOrigins: getEnv("ATLAS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
