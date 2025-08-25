package utils

import (
	"regexp"
)

// SanitizeRegex helps remove characters unsafe for filenames.
var SanitizeRegex = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

// SanitizeFilename replaces characters unsafe for filenames with underscores.
func SanitizeFilename(name string) string {
	return SanitizeRegex.ReplaceAllString(name, "_")
}
