package cmd

import (
	// Import bufio
	"bytes" // Import bytes
	"fmt"
	"io" // Import io
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus" // Import logrus
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/younsl/box/tools/ol/internal/settings"
)

var (
	cfgFile  string
	logLevel string // Variable to store log level flag
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s \"message\"", settings.AppName),
	Short: "Records a one-line message with a timestamp to a daily log file.",
	Long: fmt.Sprintf(`Records a message along with the current timestamp (based on the configured timezone)
into a log file named '%s-YYYY.txt'.

Configuration is read from $XDG_CONFIG_HOME/%s/%s.%s or $HOME/.config/%s/%s.%s.
The required configuration key is '%s' (e.g., %s: "Asia/Seoul").`,
		settings.LogFileNameFormat[:strings.Index(settings.LogFileNameFormat, "-")], // Extract base name for log file example
		settings.AppName, settings.ConfigFileNameBase, settings.ConfigFileType,
		settings.AppName, settings.ConfigFileNameBase, settings.ConfigFileType,
		settings.KeyTimezone, settings.KeyTimezone),
	Example: fmt.Sprintf("  %s \"Just finished a quick meeting.\"\n  %s \"Remember to buy milk.\"", settings.AppName, settings.AppName),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger here, before any command runs
		initLogger()
	},
	Args: cobra.ExactArgs(1), // Expect exactly one argument: the message
	Run: func(cmd *cobra.Command, args []string) {
		// Trim leading/trailing whitespace and newlines from the input message FIRST
		message := strings.TrimSpace(args[0])
		if message == "" {
			log.Warn("Cannot log an empty message after trimming whitespace.")
			return // Do not log if the message becomes empty
		}
		logMessage(message) // Pass the trimmed message
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(initConfig) // initConfig still needed for viper
	err := rootCmd.Execute()
	if err != nil {
		// Use logrus for final error reporting if Execute fails
		log.Fatalf("Error executing CLI: %v", err)
		// os.Exit(1) // log.Fatalf already exits with status 1
	}
}

func init() {
	// Add persistent flag for log level to the root command
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error, fatal, panic)")
	// No other flags defined directly on root command for now
}

// initLogger initializes the global logrus logger based on the log-level flag.
func initLogger() {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Invalid log level specified: %s. Defaulting to info.", logLevel)
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{ // Or log.JSONFormatter{}
		// FullTimestamp: true,
	})
	log.Debug("Logger initialized")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Log attempts to find/read config file
	log.Debug("Initializing configuration...")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		log.Debugf("Using config file explicitly set: %s", cfgFile)
	} else {
		configFile, err := settings.GetConfigFilePath()
		if err != nil {
			log.Warnf("Could not determine standard config file path: %v. Viper will not search for a config file.", err)
		} else {
			log.Debugf("Searching for config file %s in %s", settings.ConfigFileNameBase+"."+settings.ConfigFileType, filepath.Dir(configFile))
			viper.AddConfigPath(filepath.Dir(configFile))
			viper.SetConfigName(settings.ConfigFileNameBase)
			viper.SetConfigType(settings.ConfigFileType)
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("Config file not found. Using defaults where applicable.")
		} else {
			log.Warnf("Error reading config file: %v", err)
		}
	}

	viper.SetDefault(settings.KeyTimezone, settings.DefaultTimezone)
	log.Debugf("Default timezone set to: %s", settings.DefaultTimezone)
}

func logMessage(message string) { // message is already trimmed
	// Setup timezone, timestamp, log entry, and header
	timezoneStr := viper.GetString(settings.KeyTimezone)
	location, err := time.LoadLocation(timezoneStr)
	if err != nil {
		log.Warnf("Error loading timezone '%s': %v. Using %s.", timezoneStr, err, settings.DefaultTimezone)
		location = time.UTC
	}
	now := time.Now().In(location)
	timestamp := now.Format("2006-01-02 15:04:05 MST")
	logEntry := fmt.Sprintf("%s | %s\n", timestamp, message) // logEntry ALWAYS ends with \n
	header := "# This file is generated and modified by the ol command.\n"
	headerBytes := []byte(header)

	// Determine log file path and ensure directory exists
	baseLogFileName := fmt.Sprintf(settings.LogFileNameFormat, now.Year())
	logFilePath, err := settings.GetDataFilePath(baseLogFileName)
	if err != nil {
		log.Errorf("Error determining log file path: %v", err)
		return
	}
	logDirPath := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		log.Errorf("Error creating data directory '%s': %v", logDirPath, err)
		return
	}

	// Open file with necessary flags for robust appending and checks
	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Errorf("Error opening/creating log file '%s': %v", logFilePath, err)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		log.Errorf("Error getting file info for '%s': %v", logFilePath, err)
		return
	}

	// Handle Header and Preceding Newline
	if fileInfo.Size() == 0 {
		// File is new or empty, write header
		if _, err := file.Write(headerBytes); err != nil {
			log.Errorf("Error writing header to new log file '%s': %v. Aborting.", logFilePath, err)
			return // Stop if header write fails on new file
		}
		log.Debugf("Added header to new log file: %s", logFilePath)
		// No preceding newline needed right after header
	} else {
		// File exists and is not empty. Check header presence (optional warning)
		headerCheckBytes := make([]byte, len(headerBytes))
		bytesRead, readErr := file.ReadAt(headerCheckBytes, 0)
		if readErr != nil && readErr != io.EOF {
			log.Warnf("Could not read start of '%s' to check header: %v", logFilePath, readErr)
		} else if bytesRead < len(headerBytes) || !bytes.Equal(headerCheckBytes[:bytesRead], headerBytes) {
			log.Warnf("Existing log file '%s' does not start with the expected header.", logFilePath)
			// We don't rewrite the file, just warn.
		}

		// Check if a preceding newline is needed before appending the new log entry
		lastByte := make([]byte, 1)
		_, readErr = file.ReadAt(lastByte, fileInfo.Size()-1) // Read the actual last byte

		needsPrecedingNewline := false
		if readErr != nil {
			// Couldn't read last byte, safest bet is to add a newline
			log.Warnf("Could not read last byte of '%s': %v. Adding preceding newline.", logFilePath, readErr)
			needsPrecedingNewline = true
		} else if lastByte[0] != '\n' {
			// Last byte is not a newline
			log.Debugf("Last byte of '%s' is not '\\n'. Adding preceding newline.", logFilePath)
			needsPrecedingNewline = true
		}

		if needsPrecedingNewline {
			// Write the newline (O_APPEND ensures it goes to the end)
			if _, err := file.WriteString("\n"); err != nil {
				log.Errorf("Error writing preceding newline to '%s': %v. Aborting log entry.", logFilePath, err)
				return // Stop if writing the separator fails
			}
			// Optional: Sync after writing separator if experiencing issues
			// if err := file.Sync(); err != nil { log.Warnf(...) }
		}
	}

	// Append the actual log entry (which already ends with \n)
	if _, err := file.WriteString(logEntry); err != nil {
		log.Errorf("Error writing log entry to '%s': %v", logFilePath, err)
	} else {
		log.Infof("Logged to %s", logFilePath)
	}
}
