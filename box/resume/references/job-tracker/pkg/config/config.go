package config

import (
	"os"

	"github.com/younsl/box/resume/references/job-tracker/pkg/crypto"
	"github.com/younsl/box/resume/references/job-tracker/pkg/logging"
)

type Config struct {
	Port         string
	GPGRecipient string
	DataFile     string
}

func Load() *Config {
	cfg := &Config{
		Port:         getEnv("PORT", "1314"),
		GPGRecipient: getEnv("GPG_RECIPIENT", ""),
		DataFile:     getEnv("DATA_FILE", "data.db.gpg"),
	}

	// Validate and setup GPG recipient - fail fast if not configured properly
	recipient, err := crypto.ValidateAndSetupGPG(cfg.GPGRecipient)
	if err != nil {
		logging.Logger.WithError(err).WithField("gpg_recipient", cfg.GPGRecipient).Fatal("GPG setup required but failed. Please ensure: 1. GPG is installed and configured, 2. Set GPG_RECIPIENT environment variable, 3. Your GPG key is trusted")
	}
	cfg.GPGRecipient = recipient

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}