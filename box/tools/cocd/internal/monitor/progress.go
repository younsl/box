package monitor

import (
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

// ProgressTracker manages scan progress tracking
type ProgressTracker struct {
	mu       sync.RWMutex
	progress ScanProgress
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	now := time.Now()
	return &ProgressTracker{
		progress: ScanProgress{
			ScanMode: ScanModeIdle,
			CurrentStateStart: &now,
			StateDuration: 0,
		},
	}
}

// GetProgress returns the current progress (thread-safe)
func (pt *ProgressTracker) GetProgress() ScanProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.progress
}

// UpdateProgress updates the progress (thread-safe)
func (pt *ProgressTracker) UpdateProgress(progress ScanProgress) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.progress = progress
}

// SetMode sets the scan mode
func (pt *ProgressTracker) SetMode(mode string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	// If mode is changing, update the state start time
	if pt.progress.ScanMode != mode {
		now := time.Now()
		pt.progress.CurrentStateStart = &now
		pt.progress.StateDuration = 0
	}
	
	pt.progress.ScanMode = mode
}

// SetIdle sets the tracker to idle state
func (pt *ProgressTracker) SetIdle() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	now := time.Now()
	pt.progress = ScanProgress{
		ScanMode: ScanModeIdle,
		CurrentStateStart: &now,
		StateDuration: 0,
	}
}

// InitializeProgress initializes progress for a new scan
func (pt *ProgressTracker) InitializeProgress(mode string, totalRepos, activeRepos, maxWorkers int, repoStats RepoStats) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	// Calculate limited repos (capped at 100)
	limitedRepos := activeRepos
	if limitedRepos > 100 {
		limitedRepos = 100
	}
	
	// If mode is changing, update the state start time
	now := time.Now()
	var stateStart *time.Time
	if pt.progress.ScanMode != mode {
		stateStart = &now
	} else {
		stateStart = pt.progress.CurrentStateStart
	}
	
	pt.progress = ScanProgress{
		ActiveWorkers:     maxWorkers,
		TotalRepos:        totalRepos,
		CompletedRepos:    0,
		ScanMode:          mode,
		ActiveRepos:       activeRepos,
		ArchivedRepos:     repoStats.Archived,
		DisabledRepos:     repoStats.Disabled,
		ValidRepos:        repoStats.Valid,
		LimitedRepos:      limitedRepos,
		CurrentStateStart: stateStart,
		StateDuration:     0,
	}
}

// UpdateCompleted updates the number of completed repositories
func (pt *ProgressTracker) UpdateCompleted(completed int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.progress.CompletedRepos = completed
}

// RepoStats holds repository statistics
type RepoStats struct {
	Total    int
	Archived int
	Disabled int
	Valid    int
}

// CalculateRepoStats calculates repository statistics
func CalculateRepoStats(repos []*github.Repository) RepoStats {
	stats := RepoStats{Total: len(repos)}
	
	for _, repo := range repos {
		if repo.GetArchived() {
			stats.Archived++
		} else if repo.GetDisabled() {
			stats.Disabled++
		} else {
			stats.Valid++
		}
	}
	
	return stats
}

// SendProgressUpdates sends progress updates through a channel
func (pt *ProgressTracker) SendProgressUpdates(progressChan chan<- ScanProgress) {
	if progressChan != nil {
		progressChan <- pt.GetProgress()
	}
}

// SetNextScanTimer sets the next scan timer information
func (pt *ProgressTracker) SetNextScanTimer(nextScanAt time.Time, cycleCount int, isFullScan bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.progress.NextScanAt = &nextScanAt
	pt.progress.ScanCycleCount = cycleCount
	pt.progress.IsNextScanFull = isFullScan
	
	// Calculate countdown in seconds
	pt.progress.ScanCountdown = int(time.Until(nextScanAt).Seconds())
	if pt.progress.ScanCountdown < 0 {
		pt.progress.ScanCountdown = 0
	}
}

// UpdateScanCountdown updates the countdown timer and state duration
func (pt *ProgressTracker) UpdateScanCountdown() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	// Update scan countdown
	if pt.progress.NextScanAt != nil {
		pt.progress.ScanCountdown = int(time.Until(*pt.progress.NextScanAt).Seconds())
		if pt.progress.ScanCountdown < 0 {
			pt.progress.ScanCountdown = 0
		}
	}
	
	// Update state duration
	if pt.progress.CurrentStateStart != nil {
		pt.progress.StateDuration = int(time.Since(*pt.progress.CurrentStateStart).Seconds())
		if pt.progress.StateDuration < 0 {
			pt.progress.StateDuration = 0
		}
	}
}

// SetScanCompleted marks a scan as completed
func (pt *ProgressTracker) SetScanCompleted() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	now := time.Now()
	pt.progress.LastScanAt = &now
}