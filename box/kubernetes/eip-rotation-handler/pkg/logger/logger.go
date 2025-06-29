package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// New creates new logger instance
func New(level string) *logrus.Logger {
	log := logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// Use JSON format (good for Kubernetes)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
	})

	// Output to stdout
	log.SetOutput(os.Stdout)

	return log
}
