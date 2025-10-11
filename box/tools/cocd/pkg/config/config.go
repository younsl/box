package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GitHub GitHubConfig `mapstructure:"github"`
	Monitor MonitorConfig `mapstructure:"monitor"`
}

type GitHubConfig struct {
	Token    string `mapstructure:"token"`
	BaseURL  string `mapstructure:"base_url"`
	Org      string `mapstructure:"org"`
	Repo     string `mapstructure:"repo"`
}

type MonitorConfig struct {
	Interval int `mapstructure:"interval"`
	Timezone string `mapstructure:"timezone"`
}

func Load() (*Config, error) {
	// Check if config exists, if not create skeleton
	if !ConfigExists() {
		configPath, err := TryCreateDefaultConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create default config: %v\n", err)
		} else {
			// Set the config file explicitly to use the newly created one
			viper.SetConfigFile(configPath)
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add config paths in priority order
	for _, path := range GetConfigPaths() {
		dir := filepath.Dir(path)
		viper.AddConfigPath(dir)
	}

	viper.SetEnvPrefix("COCD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("github.base_url", "api.github.com")
	viper.SetDefault("monitor.interval", 5)
	viper.SetDefault("monitor.timezone", "UTC")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Ensure base_url has https:// prefix
	if config.GitHub.BaseURL != "" && !strings.HasPrefix(config.GitHub.BaseURL, "http://") && !strings.HasPrefix(config.GitHub.BaseURL, "https://") {
		config.GitHub.BaseURL = "https://" + config.GitHub.BaseURL
	}

	if config.GitHub.Token == "" {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			config.GitHub.Token = token
		} else if token := getGHToken(); token != "" {
			config.GitHub.Token = token
		} else {
			return nil, fmt.Errorf("GitHub token is required. Please set GITHUB_TOKEN environment variable or login with 'gh auth login'")
		}
	}

	return &config, nil
}

func getGHToken() string {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}