package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus" // Import logrus
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/younsl/box/tools/ol/internal/settings" // Import settings package
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage ol configuration settings.`,
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	Long: fmt.Sprintf(`Creates a default configuration file at the standard location
($XDG_CONFIG_HOME/%s/%s.%s or $HOME/.config/%s/%s.%s)
if one does not already exist. The default timezone is set to %s.`,
		settings.AppName, settings.ConfigFileNameBase, settings.ConfigFileType,
		settings.AppName, settings.ConfigFileNameBase, settings.ConfigFileType,
		settings.DefaultTimezone),
	Run: func(cmd *cobra.Command, args []string) {
		createDefaultConfig()
	},
	Example: fmt.Sprintf("  %s config init", settings.AppName),
}

var setConfigCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: fmt.Sprintf("Set a configuration value (e.g., '%s', '%s')", settings.KeyTimezone, settings.KeyDataDirectory),
	Long: fmt.Sprintf(`Sets a specific configuration value in the configuration file.
Supported keys: '%s', '%s'.`,
		settings.KeyTimezone, settings.KeyDataDirectory),
	Example: fmt.Sprintf(`  %[1]s config set %[2]s Asia/Seoul
  %[1]s config set %[3]s ~/Documents/OlData
  %[1]s config set %[3]s "/path/with spaces/My Data"`, settings.AppName, settings.KeyTimezone, settings.KeyDataDirectory),
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		setConfigValue(key, value)
	},
}

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Show current configuration paths and values",
	Long:  `Displays the location of the configuration file, data directory, and the currently active configuration values.`,
	Run: func(cmd *cobra.Command, args []string) {
		getCurrentConfig()
	},
	Example: fmt.Sprintf("  %s config get", settings.AppName),
}

// This init function belongs to the config_cmd.go file
func init() {
	rootCmd.AddCommand(configCmd)       // Add config command to root
	configCmd.AddCommand(initConfigCmd) // Add init subcommand to config
	configCmd.AddCommand(setConfigCmd)  // Add set subcommand to config
	configCmd.AddCommand(getConfigCmd)  // Add get subcommand to config
}

func createDefaultConfig() {
	configFile, err := settings.GetConfigFilePath()
	if err != nil {
		log.Errorf("Error determining config file path: %v", err)
		return
	}

	fileExists := false
	if _, err := os.Stat(configFile); err == nil {
		fileExists = true
		fmt.Printf("Configuration file already exists: %s\n", configFile)
		fmt.Print("Do you want to overwrite it with defaults? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("Error reading input: %v", err)
			return
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			log.Info("Configuration file overwrite cancelled.")
			return
		}
	} else if !os.IsNotExist(err) {
		log.Errorf("Error checking config file status: %v", err)
		return
	}

	if !fileExists {
		fmt.Printf("Configuration file not found. Create default file at %s? (y/n): ", configFile)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("Error reading input: %v", err)
			return
		}

		input = strings.ToLower(strings.TrimSpace(input))
		if input != "y" && input != "yes" {
			log.Info("Configuration file creation cancelled.")
			return
		}
	}

	defaultTimezone := viper.GetString(settings.KeyTimezone)
	defaultConfig := map[string]interface{}{
		settings.KeyTimezone:      defaultTimezone,
		settings.KeyDataDirectory: "",
	}

	if err := settings.WriteConfig(configFile, defaultConfig); err != nil {
		log.Error(err)
		return
	}

	if fileExists {
		log.Infof("Default configuration file overwritten: %s", configFile)
	} else {
		log.Infof("Default configuration file created: %s", configFile)
	}
}

func setConfigValue(key string, value string) {
	switch key {
	case settings.KeyTimezone:
		if _, err := time.LoadLocation(value); err != nil {
			log.Errorf("Invalid timezone '%s'. Please use a valid IANA Time Zone name. %v", value, err)
			return
		}
	case settings.KeyDataDirectory:
		if _, err := settings.ExpandPath(value); err != nil {
			log.Errorf("Error expanding path '%s': %v", value, err)
			return
		}
	default:
		log.Errorf("Unsupported configuration key '%s'. Supported keys: %s, %s", key, settings.KeyTimezone, settings.KeyDataDirectory)
		return
	}

	configFile, err := settings.GetConfigFilePath()
	if err != nil {
		log.Errorf("Error determining config file path: %v", err)
		return
	}

	configData, err := settings.ReadConfig(configFile)
	if err != nil {
		log.Error(err)
		return
	}

	configData[key] = value

	if err := settings.WriteConfig(configFile, configData); err != nil {
		log.Error(err)
		return
	}

	log.Infof("Configuration updated: Set '%s' to '%s' in %s", key, value, configFile)
}

func getCurrentConfig() {
	// --- Paths Section --- (Refactored with tabwriter)
	fmt.Println("Current Configuration Paths:")
	// Initialize tabwriter for paths
	pathTw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) // 2 space padding

	// Write Paths header
	fmt.Fprintln(pathTw, "TYPE\tPATH\tSTATUS")
	fmt.Fprintln(pathTw, "----\t----\t------")

	configFile, err := settings.GetConfigFilePath()
	if err != nil {
		// Print error directly if path calculation fails significantly
		fmt.Fprintf(os.Stderr, "  Error determining config file path: %v\n", err)
	} else {
		var existsStatus string
		if _, statErr := os.Stat(configFile); statErr == nil {
			existsStatus = "Exists"
		} else if os.IsNotExist(statErr) {
			existsStatus = "Not Found"
		} else {
			existsStatus = fmt.Sprintf("Error (%v)", statErr)
		}
		fmt.Fprintf(pathTw, "Config File\t%s\t%s\n", configFile, existsStatus)
	}

	dataPath, err := settings.GetDataDirPath() // This call now reflects config
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error determining data directory path: %v\n", err)
	} else {
		var existsStatus string
		if _, statErr := os.Stat(dataPath); statErr == nil {
			existsStatus = "Exists"
		} else if os.IsNotExist(statErr) {
			existsStatus = "Not Found (will be created on use)"
		} else {
			existsStatus = fmt.Sprintf("Error (%v)", statErr)
		}
		fmt.Fprintf(pathTw, "Data Directory\t%s\t%s\n", dataPath, existsStatus)
	}
	pathTw.Flush() // Flush path table

	// --- Values Section --- (Uses a separate tabwriter)
	fmt.Println("\nCurrent Configuration Values:")

	configFileUsed := viper.ConfigFileUsed()
	if configFileUsed != "" {
		fmt.Printf("  (Source: %s)\n", configFileUsed)
	} else {
		if _, err := os.Stat(configFile); err == nil { // Re-check configFile path determination result
			fmt.Printf("  (Source: Config file exists at %s but might have errors or wasn't loaded by Viper)\n", configFile)
		} else {
			fmt.Println("  (Source: No config file loaded, using defaults where applicable)")
		}
	}

	// Initialize tabwriter for values
	valueTw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) // 2 space padding
	defer valueTw.Flush()                                      // Ensure buffer is flushed at the end

	// Write Values header
	fmt.Fprintln(valueTw, "KEY\tVALUE\tSOURCE\tEFFECTIVE PATH")
	fmt.Fprintln(valueTw, "---\t-----\t------\t--------------")

	// Iterate through defined settings
	for _, item := range settings.AppSettings {
		value := viper.GetString(item.Key)
		source := settings.GetValueSource(item.Key, configFileUsed)
		effectivePath := "-" // Default for non-path or unset path

		if item.IsPath {
			if item.Key == settings.KeyDataDirectory {
				// Use dataPath calculated earlier (dataPath itself already reflects config)
				if value == "" {
					value = "[Not Set]" // Display explicitly unset
					effectivePath = dataPath
				} else {
					effectivePath = dataPath // Always show the calculated path
				}
			}
			// Add similar logic here if more path settings are introduced
		}

		// Write row to tabwriter
		fmt.Fprintf(valueTw, "%s\t%s\t%s\t%s\n",
			item.Key, value, source, effectivePath)
	}
}
