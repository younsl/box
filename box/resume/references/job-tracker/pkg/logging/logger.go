package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	
	// Set output to stdout
	Logger.SetOutput(os.Stdout)
	
	// Set log format
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})
	
	// Set log level from environment variable, default to Info
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
}