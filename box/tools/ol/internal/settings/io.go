package settings

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// readConfig reads the configuration file from the given path.
// If the file doesn't exist, it returns an empty map and nil error.
func ReadConfig(filePath string) (map[string]interface{}, error) { // Exported
	configData := make(map[string]interface{})

	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("Config file '%s' not found, returning empty config map.", filePath)
			return configData, nil
		}
		// Other read error
		return nil, fmt.Errorf("error reading config file '%s': %w", filePath, err)
	}

	if err := yaml.Unmarshal(yamlData, &configData); err != nil {
		return nil, fmt.Errorf("error unmarshalling config file '%s': %w", filePath, err)
	}
	log.Debugf("Successfully read and parsed config file '%s'", filePath)
	return configData, nil
}

// writeConfig writes the configuration map to the given file path.
// It ensures the directory exists before writing.
func WriteConfig(filePath string, configData map[string]interface{}) error { // Exported
	configPath := filepath.Dir(filePath)
	// Ensure directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Infof("Creating directory: %s", configPath)
		if err := os.MkdirAll(configPath, 0755); err != nil {
			return fmt.Errorf("error creating config directory '%s': %w", configPath, err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking config directory status '%s': %w", configPath, err)
	}

	// Marshal the config map to YAML bytes
	yamlData, err := yaml.Marshal(&configData)
	if err != nil {
		return fmt.Errorf("error marshalling config to YAML: %w", err)
	}

	// Write the YAML data to the file
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("error writing configuration file '%s': %w", filePath, err)
	}
	log.Debugf("Successfully wrote config file '%s'", filePath)
	return nil
}