package monitor

import (
	"context"
	"time"

	ghclient "github.com/younsl/cocd/internal/github"
	"github.com/younsl/cocd/internal/scanner"
)

const (
	// Monitor configuration
	DefaultWorkerPoolSize    = 2                  // Default number of workers for scanning
	DefaultScanTimeout       = 60 * time.Second  // Timeout for pending jobs scan
	DefaultRecentScanTimeout = 90 * time.Second  // Timeout for recent jobs scan
	
	// Repository limits
	MaxSmartRepositories  = 100 // Maximum repositories for smart scanning
	MaxActiveRepositories = 100 // Maximum repositories for recent activity scanning
	MaxRecentJobs         = 200 // Maximum number of recent jobs to return
	
	// Cache cleanup configuration
	CacheCleanupInterval = 5 * time.Minute // Interval for cache cleanup routine
)

// Monitor coordinates repository scanning and job monitoring
type Monitor struct {
	// Core components
	client         *ghclient.Client
	repoManager    *RepositoryManager
	progressTracker *ProgressTracker
	envCache       *EnvironmentCache
	
	// Scanners
	smartScanner  *scanner.SmartScanner
	recentScanner *scanner.RecentJobsScanner
	
	// Configuration
	environment string
	interval    time.Duration
	
	// Worker pool
	smartWorkerPool *WorkerPool
}

// NewMonitor creates a new monitor instance
func NewMonitor(client *ghclient.Client, environment string, interval int) *Monitor {
	// Initialize components
	repoManager := NewRepositoryManager(client)
	progressTracker := NewProgressTracker()
	envCache := NewEnvironmentCache(client)
	
	smartScanner := scanner.NewSmartScanner(client, environment, envCache)
	recentScanner := scanner.NewRecentJobsScanner(client)
	
	smartWorkerPool := NewWorkerPool(DefaultWorkerPoolSize, smartScanner)
	
	return &Monitor{
		client:          client,
		repoManager:     repoManager,
		progressTracker: progressTracker,
		envCache:        envCache,
		smartScanner:    smartScanner,
		recentScanner:   recentScanner,
		environment:     environment,
		interval:        time.Duration(interval) * time.Second,
		smartWorkerPool: smartWorkerPool,
	}
}

// GetProgressTracker returns the progress tracker
func (m *Monitor) GetProgressTracker() *ProgressTracker {
	return m.progressTracker
}

// GetClient returns the GitHub client
func (m *Monitor) GetClient() *ghclient.Client {
	return m.client
}

// GetScanProgress returns the current scan progress with cache and memory info
func (m *Monitor) GetScanProgress() ScanProgress {
	progress := m.progressTracker.GetProgress()
	progress.CacheStatus = m.repoManager.GetCacheStatus()
	progress.MemoryUsage = m.repoManager.GetMemoryUsage()
	return progress
}

// GetUpdateInterval returns the monitoring interval in seconds
func (m *Monitor) GetUpdateInterval() int {
	return int(m.interval.Seconds())
}

// GetPendingJobs returns pending jobs using smart scanning
func (m *Monitor) GetPendingJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return m.GetPendingJobsWithProgress(ctx, nil)
}

// GetPendingJobsWithProgress returns pending jobs with progress reporting using smart scanning
func (m *Monitor) GetPendingJobsWithProgress(ctx context.Context, progressChan chan<- ScanProgress) ([]scanner.JobStatus, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultScanTimeout)
	defer cancel()

	allRepos, err := m.repoManager.GetRepositoriesWithCache(timeoutCtx)
	if err != nil {
		return nil, err
	}

	smartRepos, err := m.repoManager.GetSmartRepositories(timeoutCtx, MaxSmartRepositories)
	if err != nil {
		return nil, err
	}

	repoStats := CalculateRepoStats(allRepos)

	m.progressTracker.InitializeProgress(ScanModeSmart, len(allRepos), len(smartRepos), DefaultWorkerPoolSize, repoStats)
	
	if progressChan != nil {
		progressChan <- m.progressTracker.GetProgress()
	}

	progress := m.progressTracker.GetProgress()
	jobs, err := m.smartWorkerPool.ScanRepositories(timeoutCtx, smartRepos, progressChan, &progress)
	if err != nil {
		return nil, err
	}

	SortJobsByTime(jobs, false)

	m.progressTracker.SetIdle()

	return jobs, nil
}


// GetRecentJobs returns recent jobs
func (m *Monitor) GetRecentJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return m.GetRecentJobsWithProgress(ctx, nil)
}

// GetRecentJobsWithProgress returns recent jobs with progress reporting
func (m *Monitor) GetRecentJobsWithProgress(ctx context.Context, progressChan chan<- ScanProgress) ([]scanner.JobStatus, error) {
	// Create a context with optimized timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultRecentScanTimeout)
	defer cancel()

	// Get active repositories for faster initial results (Recent Jobs Fast Mode)
	activeRepos, err := m.repoManager.GetActiveRepositories(timeoutCtx, MaxActiveRepositories)
	if err != nil {
		return nil, err
	}

	allRepos, err := m.repoManager.GetRepositoriesWithCache(timeoutCtx)
	if err != nil {
		return nil, err
	}

	repoStats := CalculateRepoStats(allRepos)

	// Initialize progress tracking for recent jobs scan
	m.progressTracker.InitializeProgress(ScanModeRecent, len(allRepos), len(activeRepos), DefaultWorkerPoolSize, repoStats)
	
	if progressChan != nil {
		progressChan <- m.progressTracker.GetProgress()
	}

	// Create worker pool for recent jobs scanning
	recentWorkerPool := NewWorkerPool(DefaultWorkerPoolSize, m.recentScanner)
	
	// Scan repositories
	progress := m.progressTracker.GetProgress()
	jobs, err := recentWorkerPool.ScanRepositories(timeoutCtx, activeRepos, progressChan, &progress)
	if err != nil {
		return nil, err
	}

	// Sort by creation time (most recent first)
	SortJobsByTime(jobs, true)

	// Limit to most recent jobs
	jobs = LimitJobs(jobs, MaxRecentJobs)

	m.progressTracker.SetIdle()

	return jobs, nil
}

// StartMonitoring starts continuous monitoring with smart scanning
func (m *Monitor) StartMonitoring(ctx context.Context, jobChan chan<- []scanner.JobStatus) {
	go m.startCacheCleanup(ctx)
	
	nextScanAt := time.Now().Add(m.interval)
	m.progressTracker.SetNextScanTimer(nextScanAt, 1, false)
	
	go func() {
		jobs, err := m.GetPendingJobs(ctx)
		if err != nil {
			jobChan <- []scanner.JobStatus{}
			return
		}
		m.progressTracker.SetScanCompleted()
		jobChan <- jobs
	}()
	
	smartTicker := time.NewTicker(m.interval)
	defer smartTicker.Stop()

	scanCounter := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-smartTicker.C:
			scanCounter++
			
			nextScanAt := time.Now().Add(m.interval)
			m.progressTracker.SetNextScanTimer(nextScanAt, scanCounter, false)
			
			jobs, err := m.GetPendingJobs(ctx)
			if err != nil {
				continue
			}
			m.progressTracker.SetScanCompleted()
			jobChan <- jobs
		}
	}
}

// startCacheCleanup starts the cache cleanup routine
func (m *Monitor) startCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(CacheCleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.envCache.CleanupExpiredCache()
		}
	}
}