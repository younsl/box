package monitor

import (
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

const (
	DefaultFullScanInterval = 10
)

type ProgressTracker struct {
	mu       sync.RWMutex
	progress ScanProgress
}

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

func (pt *ProgressTracker) GetProgress() ScanProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.progress
}

func (pt *ProgressTracker) UpdateProgress(progress ScanProgress) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.progress = progress
}

func (pt *ProgressTracker) SetMode(mode string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if pt.progress.ScanMode != mode {
		now := time.Now()
		pt.progress.CurrentStateStart = &now
		pt.progress.StateDuration = 0
	}
	
	pt.progress.ScanMode = mode
}

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

func (pt *ProgressTracker) SetCompleted() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.progress.ScanMode = "Completed"
	now := time.Now()
	pt.progress.CurrentStateStart = &now
	pt.progress.StateDuration = 0
}

func (pt *ProgressTracker) InitializeProgress(mode string, totalRepos, activeRepos, maxWorkers int, repoStats RepoStats) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	limitedRepos := activeRepos
	if limitedRepos > 100 {
		limitedRepos = 100
	}
	
	now := time.Now()
	var stateStart *time.Time
	if pt.progress.ScanMode != mode {
		stateStart = &now
	} else {
		stateStart = pt.progress.CurrentStateStart
	}
	
	preservedCompletedRepos := 0
	if pt.progress.CompletedRepos > 0 && pt.progress.CompletedRepos <= limitedRepos {
		preservedCompletedRepos = pt.progress.CompletedRepos
	}
	
	pt.progress = ScanProgress{
		ActiveWorkers:     maxWorkers,
		TotalRepos:        totalRepos,
		CompletedRepos:    preservedCompletedRepos,
		ScanMode:          mode,
		ActiveRepos:       activeRepos,
		ArchivedRepos:     repoStats.Archived,
		DisabledRepos:     repoStats.Disabled,
		ValidRepos:        repoStats.Valid,
		LimitedRepos:      limitedRepos,
		CurrentStateStart: stateStart,
		StateDuration:     0,
		NextScanAt:        pt.progress.NextScanAt,
		LastScanAt:        pt.progress.LastScanAt,
		ScanCountdown:     pt.progress.ScanCountdown,
		ScanCycleCount:    pt.progress.ScanCycleCount,
		IsNextScanFull:    pt.progress.IsNextScanFull,
	}
}

func (pt *ProgressTracker) UpdateCompleted(completed int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	// Only update if the new value is greater than current (monotonic increase)
	if completed > pt.progress.CompletedRepos {
		pt.progress.CompletedRepos = completed
	}
}


type RepoStats struct {
	Total    int
	Archived int
	Disabled int
	Valid    int
}

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

func (pt *ProgressTracker) SendProgressUpdates(progressChan chan<- ScanProgress) {
	if progressChan != nil {
		progressChan <- pt.GetProgress()
	}
}

func (pt *ProgressTracker) SetNextScanTimer(nextScanAt time.Time, cycleCount int, isFullScan bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.progress.NextScanAt = &nextScanAt
	pt.progress.ScanCycleCount = cycleCount
	
	pt.progress.IsNextScanFull = (cycleCount % DefaultFullScanInterval == 0)
	
	pt.progress.ScanCountdown = int(time.Until(nextScanAt).Seconds())
	if pt.progress.ScanCountdown < 0 {
		pt.progress.ScanCountdown = 0
	}
}

func (pt *ProgressTracker) UpdateScanCountdown() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if pt.progress.NextScanAt != nil {
		pt.progress.ScanCountdown = int(time.Until(*pt.progress.NextScanAt).Seconds())
		if pt.progress.ScanCountdown < 0 {
			pt.progress.ScanCountdown = 0
		}
	}
	
	if pt.progress.CurrentStateStart != nil {
		pt.progress.StateDuration = int(time.Since(*pt.progress.CurrentStateStart).Seconds())
		if pt.progress.StateDuration < 0 {
			pt.progress.StateDuration = 0
		}
	}
}

func (pt *ProgressTracker) SetScanCompleted() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	now := time.Now()
	pt.progress.LastScanAt = &now
}