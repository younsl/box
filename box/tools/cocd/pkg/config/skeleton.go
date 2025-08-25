package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigSkeleton represents the skeleton structure for config.yaml
type ConfigSkeleton struct {
	GitHub  GitHubSkeleton  `yaml:"github"`
	Monitor MonitorSkeleton `yaml:"monitor"`
}

type GitHubSkeleton struct {
	Token   string `yaml:"token" comment:"GitHub token (can also be set via COCD_GITHUB_TOKEN or GITHUB_TOKEN env var)\nYou can also authenticate using 'gh auth login' command"`
	BaseURL string `yaml:"base_url" comment:"GitHub API base URL\nFor GitHub Enterprise Server, use: https://github.example.com/api/v3"`
	Org     string `yaml:"org" comment:"GitHub organization name (required)\nCan also be set via COCD_GITHUB_ORG env var"`
	Repo    string `yaml:"repo" comment:"GitHub repository name (optional)\nIf not specified, monitors all repositories in the organization\nCan also be set via COCD_GITHUB_REPO env var"`
}

type MonitorSkeleton struct {
	Interval    int    `yaml:"interval" comment:"Refresh interval in seconds"`
	Timezone    string `yaml:"timezone" comment:"Timezone for displaying timestamps\nExamples: UTC, Asia/Seoul, America/New_York, Europe/London, Asia/Tokyo"`
}

func GetDefaultConfig() *ConfigSkeleton {
	return &ConfigSkeleton{
		GitHub: GitHubSkeleton{
			Token:   "",
			BaseURL: "api.github.com",
			Org:     "",
			Repo:    "",
		},
		Monitor: MonitorSkeleton{
			Interval:    5,
			Timezone:    "UTC",
		},
	}
}

// GetConfigDir returns the config directory path following XDG Base Directory specification
func GetConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "cocd")
	}
	
	// Fall back to $HOME/.config
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "cocd")
}

func GetConfigPaths() []string {
	configDir := GetConfigDir()
	homeDir, _ := os.UserHomeDir()
	
	return []string{
		filepath.Join(configDir, "config.yaml"),  // XDG config or ~/.config/cocd/
		filepath.Join(homeDir, ".cocd", "config.yaml"), // Legacy location
		"/etc/cocd/config.yaml",                  // System-wide
	}
}

func CreateSkeletonConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", path, err)
	}
	defer file.Close()

	// Write header comment
	header := `# COCD Configuration File
# 
# This file was automatically generated. Please update it with your settings.
# You can also use environment variables to override these settings.
#
# For more information, see: https://github.com/younsl/box/tree/main/box/tools/cocd

`
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create YAML encoder with custom settings
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	
	// Create default config with YAML comments
	skeleton := GetDefaultConfig()
	
	// Marshal the config with comments
	node := &yaml.Node{}
	if err := node.Encode(skeleton); err != nil {
		return fmt.Errorf("failed to encode skeleton: %w", err)
	}
	
	// Add inline comments to fields
	addComments(node)
	
	// Write the YAML with comments
	if err := encoder.Encode(node); err != nil {
		return fmt.Errorf("failed to write YAML: %w", err)
	}

	return nil
}

func addComments(node *yaml.Node) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		addComments(node.Content[0])
	} else if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			
			switch key.Value {
			case "github":
				key.HeadComment = "GitHub configuration"
			case "token":
				key.HeadComment = "GitHub token (can also be set via COCD_GITHUB_TOKEN or GITHUB_TOKEN env var)\nYou can also authenticate using 'gh auth login' command"
			case "base_url":
				key.HeadComment = "GitHub API base URL (default: api.github.com)\nFor GitHub Enterprise Server, use: github.example.com/api/v3"
			case "org":
				key.HeadComment = "GitHub organization name (required)\nCan also be set via COCD_GITHUB_ORG env var"
			case "repo":
				key.HeadComment = "GitHub repository name (optional)\nIf not specified, monitors all repositories in the organization\nCan also be set via COCD_GITHUB_REPO env var"
			case "monitor":
				key.HeadComment = "\nMonitor configuration"
			case "interval":
				key.HeadComment = "Refresh interval in seconds (default: 5)"
			case "timezone":
				key.HeadComment = "Timezone for displaying timestamps (default: UTC)\nExamples: UTC, Asia/Seoul, America/New_York, Europe/London, Asia/Tokyo"
			}
			
			if value.Kind == yaml.MappingNode {
				addComments(value)
			}
		}
	}
}

func TryCreateDefaultConfig() (string, error) {
	// Get the default config path following XDG specification
	configDir := GetConfigDir()
	defaultPath := filepath.Join(configDir, "config.yaml")
	
	if err := CreateSkeletonConfig(defaultPath); err != nil {
		return "", err
	}

	fmt.Fprintf(os.Stderr, "\n✅ Created default config file at: %s\n", defaultPath)
	fmt.Fprintf(os.Stderr, "\n📝 Next steps:\n")
	fmt.Fprintf(os.Stderr, "   1. Edit the config file and set your GitHub organization\n")
	fmt.Fprintf(os.Stderr, "   2. Set your GitHub token using one of these methods:\n")
	fmt.Fprintf(os.Stderr, "      - Edit config file (github.token field)\n")
	fmt.Fprintf(os.Stderr, "      - Set environment variable: export COCD_GITHUB_TOKEN=<token>\n")
	fmt.Fprintf(os.Stderr, "      - Use GitHub CLI: gh auth login\n\n")

	return defaultPath, nil
}

func ConfigExists() bool {
	for _, path := range GetConfigPaths() {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}