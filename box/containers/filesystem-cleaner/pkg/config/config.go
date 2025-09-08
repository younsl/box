package config

import (
	"strings"
	"time"
)

type Config struct {
	CleanPaths       []string
	ThresholdPercent int
	CheckInterval    time.Duration
	FilePatterns     []string
	ExcludePatterns  []string
	CleanupMode      string
	DryRun           bool
}

func ParseCSV(input string) []string {
	result := []string{}
	for _, item := range strings.Split(input, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}