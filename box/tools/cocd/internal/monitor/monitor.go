package monitor

import (
	"context"
	"time"

	ghclient "github.com/younsl/cocd/internal/github"
	"github.com/younsl/cocd/internal/scanner"
)

const (
	DefaultWorkerPoolSize    = 2
	DefaultScanTimeout       = 60 * time.Second
	DefaultRecentScanTimeout = 90 * time.Second
	
	MaxSmartRepositories  = 100
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
	envCache       *EnvironmentCache
	
	smartScanner  *scanner.SmartScanner
	recentScanner *scanner.RecentJobsScanner
	
	environment string
	interval    time.Duration
	
	smartWorkerPool *WorkerPool
}

func NewMonitor(client *ghclient.Client, environment string, interval int) *Monitor {
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

func (m *Monitor) GetPendingJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return m.GetPendingJobsWithProgress(ctx, nil)
}

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

	jobs = LimitJobs(jobs, MaxRecentJobs)

	m.progressTracker.SetIdle()

	return jobs, nil
}

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