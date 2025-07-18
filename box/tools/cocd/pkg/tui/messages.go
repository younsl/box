package tui

import (
	"time"
	
	"github.com/younsl/cocd/internal/scanner"
)

// Messages for Bubble Tea
type (
	jobsMsg         []scanner.JobStatus
	recentJobsMsg   []scanner.JobStatus
	errorMsg        string
	tickMsg         time.Time
	scanProgressMsg ScanProgress
	updateUIMsg     struct{} // For forcing UI updates
	cancelSuccessMsg struct{} // For successful cancellation
)

// ScanProgress represents the progress of repository scanning
type ScanProgress struct {
	ActiveWorkers int
	TotalRepos    int
	CompletedRepos int
	CurrentView   string // "pending" or "recent"
}