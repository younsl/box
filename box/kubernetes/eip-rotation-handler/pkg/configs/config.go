package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Configuration constants
const (
	DefaultRotationMinutes = 10
	MinRotationMinutes     = 1
	MaxRotationMinutes     = 1440 // 24 hours
	DefaultLogLevel        = "info"
	DefaultMetadataURL     = "http://169.254.169.254/latest/meta-data"
	DefaultIMDSVersion     = IMDSVersionAuto // auto, v1, v2
)

// IMDS version constants
const (
	IMDSVersionAuto = "auto"
	IMDSVersionV1   = "v1"
	IMDSVersionV2   = "v2"
)

// Valid IMDS versions
var ValidIMDSVersions = []string{
	IMDSVersionAuto,
	IMDSVersionV1,
	IMDSVersionV2,
}

// Environment variable names
const (
	EnvRotationInterval = "ROTATION_INTERVAL_MINUTES"
	EnvLogLevel         = "LOG_LEVEL"
	EnvMetadataURL      = "METADATA_URL"
	EnvIMDSVersion      = "IMDS_VERSION" // auto, v1, v2
)

// Config app settings data
type Config struct {
	RotationInterval time.Duration
	LogLevel         string
	MetadataURL      string
	IMDSVersion      string // auto, v1, v2
}

// Load reads settings from environment with validation
func Load() (*Config, error) {
	rotationInterval, err := parseRotationInterval()
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	imdsVersion := getEnvOrDefault(EnvIMDSVersion, DefaultIMDSVersion)
	if err := validateIMDSVersion(imdsVersion); err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	return &Config{
		RotationInterval: rotationInterval,
		LogLevel:         getEnvOrDefault(EnvLogLevel, DefaultLogLevel),
		MetadataURL:      getEnvOrDefault(EnvMetadataURL, DefaultMetadataURL),
		IMDSVersion:      imdsVersion,
	}, nil
}

// getEnvOrDefault gets env value or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseRotationInterval parses and validates rotation interval from env
func parseRotationInterval() (time.Duration, error) {
	intervalStr := getEnvOrDefault(EnvRotationInterval, strconv.Itoa(DefaultRotationMinutes))

	minutes, err := strconv.Atoi(intervalStr)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid number, got: %s", EnvRotationInterval, intervalStr)
	}

	if err := validateRotationMinutes(minutes); err != nil {
		return 0, err
	}

	return time.Duration(minutes) * time.Minute, nil
}

// validateRotationMinutes checks if rotation minutes is within valid range
func validateRotationMinutes(minutes int) error {
	if minutes < MinRotationMinutes {
		return fmt.Errorf("%s must be at least %d minute(s), got: %d",
			EnvRotationInterval, MinRotationMinutes, minutes)
	}

	if minutes > MaxRotationMinutes {
		return fmt.Errorf("%s must be at most %d minutes (24 hours), got: %d",
			EnvRotationInterval, MaxRotationMinutes, minutes)
	}

	return nil
}

// validateIMDSVersion checks if IMDS version is valid
func validateIMDSVersion(version string) error {
	for _, valid := range ValidIMDSVersions {
		if version == valid {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %v, got: %s", EnvIMDSVersion, ValidIMDSVersions, version)
}
