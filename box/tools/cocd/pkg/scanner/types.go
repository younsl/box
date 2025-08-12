package scanner

import (
	"fmt"
	"strings"
	"time"
)

// JobStatus represents the status of a GitHub Actions job
type JobStatus struct {
	ID           int64
	Name         string
	RunID        int64
	RunNumber    int
	Status       string
	Conclusion   string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Environment  string
	WorkflowName string
	Branch       string
	Event        string
	Actor        string
	Repository   string
	
	// UI highlighting for newly scanned jobs
	IsNewlyScanned bool      `json:"-"` // Track if this job was just discovered
	HighlightUntil *time.Time `json:"-"` // When to stop highlighting this job
}

// RepoScanResult represents the result of scanning a repository
type RepoScanResult struct {
	Jobs []JobStatus
	Err  error
}

// GetActionsURL returns the GitHub Actions URL for this job
func (js JobStatus) GetActionsURL(baseURL, org string) string {
	// Remove /api/v3 suffix if present (for GitHub Enterprise)
	cleanBaseURL := baseURL
	if strings.HasSuffix(cleanBaseURL, "/api/v3") {
		cleanBaseURL = strings.TrimSuffix(cleanBaseURL, "/api/v3")
	}
	
	// Default to github.com if baseURL is GitHub.com API
	if cleanBaseURL == "https://api.github.com" || cleanBaseURL == "" {
		cleanBaseURL = "https://github.com"
	}
	
	// Generate GitHub Actions URL with organization
	return fmt.Sprintf("%s/%s/%s/actions/runs/%d", cleanBaseURL, org, js.Repository, js.RunID)
}