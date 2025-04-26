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
		message := args[0]
		logMessage(message)
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

func logMessage(message string) {
	// --- Timezone and Timestamp ---
	timezoneStr := viper.GetString(settings.KeyTimezone)
	location, err := time.LoadLocation(timezoneStr)
	if err != nil {
		log.Warnf("Error loading timezone '%s': %v. Using %s.", timezoneStr, err, settings.DefaultTimezone)
		location = time.UTC
		timezoneStr = settings.DefaultTimezone
	}
	now := time.Now().In(location)
	timestamp := now.Format("2006-01-02 15:04:05 MST")
	logEntry := fmt.Sprintf("%s | %s\n", timestamp, message)
	header := "# This file is generated and modified by the ol command.\n"
	headerBytes := []byte(header)

	// --- File Path and Directory ---
	baseLogFileName := fmt.Sprintf(settings.LogFileNameFormat, now.Year())
	logFilePath, err := settings.GetDataFilePath(baseLogFileName)
	if err != nil {
		log.Errorf("Error determining log file path: %v", err)
		return
	}

	logDirPath := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDirPath, 0755); err != nil { // Ensure directory exists
		log.Errorf("Error creating data directory '%s': %v", logDirPath, err)
		return
	}

	// --- Check File Status and Header ---
	var needsHeaderPrepended bool
	var originalContent []byte

	_, statErr := os.Stat(logFilePath)
	fileExists := !os.IsNotExist(statErr)

	if fileExists && statErr == nil {
		// File exists, check for header
		readFile, err := os.Open(logFilePath)
		if err != nil {
			log.Errorf("Error opening existing log file '%s' for reading: %v", logFilePath, err)
			return // Cannot proceed without reading
		}

		// Read just enough for the header check
		firstBytes := make([]byte, len(headerBytes))
		n, readErr := io.ReadFull(readFile, firstBytes) // Use ReadFull for predictable read size

		// Check if the first part matches the header
		if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
			// Unexpected error reading the file start
			readFile.Close() // Close before returning
			log.Errorf("Error reading start of log file '%s': %v", logFilePath, readErr)
			return
		}

		if n < len(headerBytes) || !bytes.Equal(firstBytes, headerBytes) {
			// Header is missing or file is smaller than header
			needsHeaderPrepended = true
			log.Debugf("Header missing or incomplete in '%s'. Prepending header.", logFilePath)

			// We need the original content to rewrite the file
			// Reset reader to beginning
			_, err = readFile.Seek(0, io.SeekStart)
			if err != nil {
				readFile.Close()
				log.Errorf("Error seeking to start of log file '%s': %v", logFilePath, err)
				return
			}
			originalContent, err = io.ReadAll(readFile)
			if err != nil {
				readFile.Close()
				log.Errorf("Error reading entire existing log file '%s': %v", logFilePath, err)
				return // Cannot proceed without original content
			}
		} else {
			log.Debugf("Header found in '%s'. Appending normally.", logFilePath)
		}

		readFile.Close() // Close the read file handle

	} else if !fileExists {
		log.Debugf("Log file '%s' does not exist. Will create with header.", logFilePath)
		// File doesn't exist, header will be added by append logic below
	} else {
		// Stat error other than NotExist
		log.Errorf("Error checking log file status '%s': %v", logFilePath, statErr)
		return // Cannot determine file state
	}

	// --- Write to File ---
	if needsHeaderPrepended {
		// Open for writing, truncating the file
		writeFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Errorf("Error opening log file '%s' for writing (prepending header): %v", logFilePath, err)
			return
		}
		defer writeFile.Close()

		// Write Header
		if _, err := writeFile.Write(headerBytes); err != nil {
			log.Errorf("Error writing header to log file '%s': %v", logFilePath, err)
			return // Abort if header write fails
		}
		// Write New Log Entry
		if _, err := writeFile.WriteString(logEntry); err != nil {
			log.Errorf("Error writing new log entry to file '%s' after header: %v", logFilePath, err)
			return // Abort if new entry write fails
		}
		// Write Original Content
		if _, err := writeFile.Write(originalContent); err != nil {
			log.Errorf("Error writing original content back to log file '%s': %v", logFilePath, err)
			// Log error but consider the main operation (adding new log) successful
		}
		log.Infof("Prepended header and logged to %s", logFilePath)

	} else {
		// Standard append logic (handles file creation automatically)
		appendFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Error opening log file '%s' for appending: %v", logFilePath, err)
			return
		}
		defer appendFile.Close()

		// Add header ONLY if the file did not exist before this operation
		if !fileExists {
			if _, err := appendFile.Write(headerBytes); err != nil {
				log.Errorf("Error writing header to new log file '%s': %v", logFilePath, err)
				// Continue to write log entry anyway
			} else {
				log.Debugf("Added header to new log file: %s", logFilePath)
			}
		}

		// Write New Log Entry
		if _, err := appendFile.WriteString(logEntry); err != nil {
			log.Errorf("Error writing log entry to file '%s': %v", logFilePath, err)
		} else {
			// Adjust log message based on whether header was already present or added now
			if fileExists {
				log.Infof("Appended log to %s", logFilePath)
			} else {
				log.Infof("Created new log file and logged to %s", logFilePath)
			}
		}
	}
}
