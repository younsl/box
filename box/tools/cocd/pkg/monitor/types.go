package monitor

import (
	"time"
)


// ScanProgress represents the progress of repository scanning
type ScanProgress struct {
	ActiveWorkers      int
	TotalRepos         int
	CompletedRepos     int
	ScanMode           string // "Smart", "Recent", "Idle"
	ActiveRepos        int    // Number of repos being scanned (for fast mode)
	ArchivedRepos      int    // Number of archived repositories
	DisabledRepos      int    // Number of disabled repositories
	ValidRepos         int    // Number of valid (non-archived, non-disabled) repositories
	LimitedRepos       int    // Number of limited repos (capped at 200 for GHES load reduction)
	CacheStatus        string // Cache status information
	MemoryUsage        string // Memory usage information
	
	// Timer information
	NextScanAt         *time.Time // Next scan scheduled time
	LastScanAt         *time.Time // Last scan completion time
	ScanCountdown      int        // Seconds until next scan
	ScanCycleCount     int        // Current cycle count (1-6)
	IsNextScanFull     bool       // Whether next scan will be full scan
	
	// State duration tracking
	CurrentStateStart  *time.Time // When the current state started
	StateDuration      int        // Seconds since current state started
}

// ScanMode constants
const (
	ScanModeIdle   = "Idle"
	ScanModeOrg    = "Organization"
	ScanModeSmart  = "Smart"
	ScanModeRecent = "Recent"
)

// Repository filter criteria
type RepoFilter struct {
	IncludeArchived bool
	IncludeDisabled bool
	MaxAge          time.Duration // For fast scanning, repos with recent activity
}

