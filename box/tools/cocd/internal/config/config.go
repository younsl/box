package config

import (
	"fmt"
	"os"
	"os/exec"
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
	Environment string `mapstructure:"environment"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.cocd")
	viper.AddConfigPath("/etc/cocd")

	viper.SetEnvPrefix("COCD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("github.base_url", "https://api.github.com")
	viper.SetDefault("monitor.interval", 5)
	viper.SetDefault("monitor.environment", "prod")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
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