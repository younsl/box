package monitor

import (
	"context"
	"time"

	ghclient "github.com/younsl/cocd/pkg/github"
	"github.com/younsl/cocd/pkg/scanner"
)

const (
	DefaultWorkerPoolSize    = 1
	DefaultScanTimeout       = 60 * time.Second
	DefaultRecentScanTimeout = 90 * time.Second
	
	MaxSmartRepositories  = 50
	MaxActiveRepositories = 100
	MaxRecentJobs         = 100
	
	CacheCleanupInterval = 5 * time.Minute
	
	MinScanInterval         = 10 * time.Second
	MaxScanInterval         = 60 * time.Second
	ScanIntervalIncrement   = 5 * time.Second
)

type Monitor struct {
	client         *ghclient.Client
	repoManager    *RepositoryManager
	progressTracker *ProgressTracker
	
	recentScanner *scanner.RecentJobsScanner
	
	interval    time.Duration
	
}

func NewMonitor(client *ghclient.Client, interval int) *Monitor {
	repoManager := NewRepositoryManager(client)
	progressTracker := NewProgressTracker()
	
	recentScanner := scanner.NewRecentJobsScanner(client)
	
	return &Monitor{
		client:          client,
		repoManager:     repoManager,
		progressTracker: progressTracker,
		recentScanner:   recentScanner,
		interval:        time.Duration(interval) * time.Second,
	}
}

func (m *Monitor) GetProgressTracker() *ProgressTracker {
	return m.progressTracker
}

func (m *Monitor) GetClient() *ghclient.Client {
	return m.client
}

func (m *Monitor) GetScanProgress() ScanProgress {
	progress := m.progressTracker.GetProgress()
	progress.CacheStatus = m.repoManager.GetCacheStatus()
	progress.MemoryUsage = m.repoManager.GetMemoryUsage()
	return progress
}

func (m *Monitor) GetUpdateInterval() int {
	return int(m.interval.Seconds())
}

func (m *Monitor) GetAuthenticatedUser(ctx context.Context) (string, error) {
	user, _, err := m.client.GetAuthenticatedUser(ctx)
	if err != nil {
		return "", err
	}
	if user.Login != nil {
		return *user.Login, nil
	}
	return "", nil
}

func (m *Monitor) GetPendingJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return m.GetPendingJobsWithProgress(ctx, nil)
}

func (m *Monitor) GetPendingJobsWithProgress(ctx context.Context, progressChan chan<- ScanProgress) ([]scanner.JobStatus, error) {
	recentJobs, err := m.GetRecentJobsWithProgress(ctx, progressChan)
	if err != nil {
		return nil, err
	}

	var waitingJobs []scanner.JobStatus
	for _, job := range recentJobs {
		if job.Status == "waiting" {
			waitingJobs = append(waitingJobs, job)
		}
	}

	SortJobsByTime(waitingJobs, false)

	return waitingJobs, nil
}


// GetRecentJobsWithStreaming gets recent jobs with real-time streaming updates
func (m *Monitor) GetRecentJobsWithStreaming(ctx context.Context, jobUpdateChan chan<- JobUpdate) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultRecentScanTimeout)
	defer cancel()

	activeRepos, err := m.repoManager.GetActiveRepositories(timeoutCtx, MaxActiveRepositories)
	if err != nil {
		return err
	}

	allRepos, err := m.repoManager.GetRepositoriesWithCache(timeoutCtx)
	if err != nil {
		return err
	}

	repoStats := CalculateRepoStats(allRepos)

	m.progressTracker.InitializeProgress(ScanModeRecent, len(allRepos), len(activeRepos), DefaultWorkerPoolSize, repoStats)
	
	recentWorkerPool := NewWorkerPool(DefaultWorkerPoolSize, m.recentScanner)
	
	err = recentWorkerPool.ScanRepositoriesStreamingWithTracker(timeoutCtx, activeRepos, jobUpdateChan, m.progressTracker)
	if err != nil {
		return err
	}

	m.progressTracker.SetCompleted()

	return nil
}


func (m *Monitor) GetRecentJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return m.GetRecentJobsWithProgress(ctx, nil)
}

func (m *Monitor) GetRecentJobsWithProgress(ctx context.Context, progressChan chan<- ScanProgress) ([]scanner.JobStatus, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultRecentScanTimeout)
	defer cancel()

	activeRepos, err := m.repoManager.GetActiveRepositories(timeoutCtx, MaxActiveRepositories)
	if err != nil {
		return nil, err
	}

	allRepos, err := m.repoManager.GetRepositoriesWithCache(timeoutCtx)
	if err != nil {
		return nil, err
	}

	repoStats := CalculateRepoStats(allRepos)

	m.progressTracker.InitializeProgress(ScanModeRecent, len(allRepos), len(activeRepos), DefaultWorkerPoolSize, repoStats)
	
	if progressChan != nil {
		progressChan <- m.progressTracker.GetProgress()
	}

	recentWorkerPool := NewWorkerPool(DefaultWorkerPoolSize, m.recentScanner)
	
	progress := m.progressTracker.GetProgress()
	jobs, err := recentWorkerPool.ScanRepositories(timeoutCtx, activeRepos, progressChan, &progress)
	if err != nil {
		return nil, err
	}

	SortJobsByTime(jobs, true)

	m.progressTracker.SetCompleted()

	return jobs, nil
}

func (m *Monitor) StartMonitoring(ctx context.Context, jobChan chan<- []scanner.JobStatus) {
	go m.startCacheCleanup(ctx)
	
	nextScanAt := time.Now().Add(m.interval)
	m.progressTracker.SetNextScanTimer(nextScanAt, 1, false)
	
	
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

func (m *Monitor) startCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(CacheCleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Repository cache cleanup handled by repository manager
		}
	}
}
