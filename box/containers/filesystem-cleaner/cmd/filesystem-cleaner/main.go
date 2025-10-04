package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/younsl/box/containers/filesystem-cleaner/pkg/cleaner"
	"github.com/younsl/box/containers/filesystem-cleaner/pkg/config"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	buildInfo = fmt.Sprintf("Version: %s, Commit: %s, Date: %s", version, commit, date)
)

func main() {
	var (
		paths            string
		thresholdPercent int
		intervalMinutes  int
		filePatterns     string
		excludePatterns  string
		cleanupMode      string
		dryRun           bool
		logLevel         string
		showVersion      bool
	)

	flag.StringVar(&paths, "target-paths", "/home/runner/_work", "Target filesystem paths to clean (comma-separated)")
	flag.IntVar(&thresholdPercent, "usage-threshold-percent", 80, "Disk usage percentage threshold to trigger cleanup (0-100)")
	flag.IntVar(&intervalMinutes, "check-interval-minutes", 10, "Interval between cleanup checks in minutes (used with cleanup-mode=interval)")
	flag.StringVar(&filePatterns, "include-patterns", "*", "File patterns to include for deletion (comma-separated)")
	flag.StringVar(&excludePatterns, "exclude-patterns", ".git,node_modules,*.log", "File/directory patterns to exclude from deletion (comma-separated)")
	flag.StringVar(&cleanupMode, "cleanup-mode", "interval", "Cleanup mode: 'once' for single run (initContainer), 'interval' for periodic cleanup")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode (don't delete files)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Println(buildInfo)
		os.Exit(0)
	}

	setupLogger(logLevel)

	// Validate cleanup mode
	if cleanupMode != "once" && cleanupMode != "interval" {
		logrus.Fatal("Invalid cleanup-mode. Must be 'once' or 'interval'")
	}

	cfg := config.Config{
		CleanPaths:       config.ParseCSV(paths),
		ThresholdPercent: thresholdPercent,
		CheckInterval:    time.Duration(intervalMinutes) * time.Minute,
		FilePatterns:     config.ParseCSV(filePatterns),
		ExcludePatterns:  config.ParseCSV(excludePatterns),
		CleanupMode:      cleanupMode,
		DryRun:           dryRun,
	}

	// Log startup with version info
	logrus.WithFields(logrus.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("Starting filesystem-cleaner")

	// Log complete configuration in single entry
	configFields := logrus.Fields{
		"target_paths":             cfg.CleanPaths,
		"usage_threshold_percent":  cfg.ThresholdPercent,
		"cleanup_mode":             cfg.CleanupMode,
		"include_patterns":         cfg.FilePatterns,
		"exclude_patterns":         cfg.ExcludePatterns,
		"dry_run":                  cfg.DryRun,
		"log_level":                logLevel,
	}

	// Add interval only for interval mode
	if cfg.CleanupMode == "interval" {
		configFields["check_interval_minutes"] = int(cfg.CheckInterval.Minutes())
	}

	logrus.WithFields(configFields).Info("Configuration loaded")

	if cfg.DryRun {
		logrus.Warn("Running in DRY-RUN mode - no files will be deleted")
	}

	c := cleaner.New(cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logrus.Info("Received shutdown signal, stopping cleaner...")
		c.Stop()
	}()

	if err := c.Run(); err != nil {
		logrus.WithError(err).Fatal("Failed to run cleaner")
	}
}

func setupLogger(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     false,
		DisableColors:   true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyFunc: "component",
		},
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			// Show package.function
			funcName := frame.Function
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
			return funcName, ""
		},
	})
	
	// Enable caller reporting
	logrus.SetReportCaller(true)

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
}