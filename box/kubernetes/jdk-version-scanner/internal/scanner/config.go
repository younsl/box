package scanner

import "time"

// Config holds the configuration for the JDK version scanner
type Config struct {
	Namespaces    []string
	MaxGoroutines int
	Timeout       time.Duration
	SkipDaemonSet bool
	Verbose       bool
}
