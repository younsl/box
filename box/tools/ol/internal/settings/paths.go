package settings

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// GetConfigFilePath determines the absolute path to the configuration file.
func GetConfigFilePath() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not find home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	configPath := filepath.Join(configHome, AppName)
	configFile := filepath.Join(configPath, fmt.Sprintf("%s.%s", ConfigFileNameBase, ConfigFileType))
	return configFile, nil
}

// ExpandPath expands environment variables (like $HOME or ${VAR})
// and then expands tilde (~) to the user's home directory.
func ExpandPath(path string) (string, error) {
	// 1. Expand environment variables first
	expandedEnvPath := os.ExpandEnv(path)

	// 2. Expand tilde (if present)
	if !strings.HasPrefix(expandedEnvPath, "~") {
		return expandedEnvPath, nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get current user: %w", err)
	}
	// Join home directory with the rest of the path (after '~')
	return filepath.Join(currentUser.HomeDir, expandedEnvPath[1:]), nil
}

// GetDataDirPath determines the absolute path for the data directory.
func GetDataDirPath() (string, error) {
	dataDir := viper.GetString(KeyDataDirectory) // Use KeyDataDirectory

	if dataDir != "" {
		// Expand environment variables and tilde
		expandedDir, err := ExpandPath(dataDir) // ExpandPath now handles both
		if err != nil {
			return "", fmt.Errorf("error expanding configured dataDirectory '%s': %w", dataDir, err) // Updated error message
		}
		// Use the configured directory
		return filepath.Clean(expandedDir), nil
	}

	// Fallback to XDG Base Directory Specification
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not find home directory: %w", err)
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	dataPath := filepath.Join(dataHome, AppName) // Use constant
	return dataPath, nil
}

// GetDataFilePath determines the absolute path for a specific data file (like the log file).
func GetDataFilePath(fileName string) (string, error) {
	dataPath, err := GetDataDirPath()
	if err != nil {
		return "", err
	}
	dataFilePath := filepath.Join(dataPath, fileName)
	return dataFilePath, nil
} 