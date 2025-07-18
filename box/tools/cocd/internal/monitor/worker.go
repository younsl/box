package monitor

import (
	"context"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/younsl/cocd/internal/scanner"
)

// WorkerPool manages concurrent repository scanning
type WorkerPool struct {
	maxWorkers int
	scanner    scanner.Scanner
	perfOptimizer *PerformanceOptimizer
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int, sc scanner.Scanner) *WorkerPool {
	return &WorkerPool{
		maxWorkers:    maxWorkers,
		scanner:       sc,
		perfOptimizer: NewPerformanceOptimizer(),
	}
}

// ScanRepositories scans multiple repositories concurrently
func (wp *WorkerPool) ScanRepositories(ctx context.Context, repos []*github.Repository, progressChan chan<- ScanProgress, progress *ScanProgress) ([]scanner.JobStatus, error) {
	if len(repos) == 0 {
		return []scanner.JobStatus{}, nil
	}

	// Create channels for concurrent processing
	repoChan := make(chan *github.Repository, len(repos))
	resultChan := make(chan scanner.RepoScanResult, len(repos))

	// Start workers with throttling for GHES performance
	for i := 0; i < wp.maxWorkers; i++ {
		go func(workerID int) {
			for repo := range repoChan {
				var jobs []scanner.JobStatus
				var err error
				
				scanErr := wp.perfOptimizer.WithPerformanceTracking(ctx, func(ctx context.Context) error {
					jobs, err = wp.scanner.ScanRepository(ctx, repo)
					return err
				})
				
				if scanErr != nil {
					err = scanErr
				}
				
				resultChan <- scanner.RepoScanResult{Jobs: jobs, Err: err}
				
				baseDelay := time.Duration(300 + (workerID*100)) * time.Millisecond
				adaptiveDelay := wp.perfOptimizer.GetOptimalDelay(baseDelay)
				time.Sleep(adaptiveDelay)
			}
		}(i)
	}

	// Send repositories to workers
	for _, repo := range repos {
		repoChan <- repo
	}
	close(repoChan)

	// Collect results with progress tracking
	var allJobs []scanner.JobStatus
	completedRepos := 0
	
	for i := 0; i < len(repos); i++ {
		result := <-resultChan
		if result.Err == nil {
			allJobs = append(allJobs, result.Jobs...)
		}
		completedRepos++
		
		// Update progress tracking
		if progress != nil {
			progress.CompletedRepos = completedRepos
			
			// Send progress update
			if progressChan != nil {
				progressChan <- *progress
			}
		}
	}

	return allJobs, nil
}

// SortJobsByTime sorts jobs by creation time
func SortJobsByTime(jobs []scanner.JobStatus, newest bool) {
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].StartedAt == nil && jobs[j].StartedAt == nil {
			return false
		}
		if jobs[i].StartedAt == nil {
			return !newest // Put nil at end for newest, at start for oldest
		}
		if jobs[j].StartedAt == nil {
			return newest // Put nil at start for newest, at end for oldest
		}
		
		if newest {
			return jobs[i].StartedAt.After(*jobs[j].StartedAt)
		} else {
			return jobs[i].StartedAt.Before(*jobs[j].StartedAt)
		}
	})
}

// LimitJobs limits the number of jobs returned
func LimitJobs(jobs []scanner.JobStatus, limit int) []scanner.JobStatus {
	if len(jobs) <= limit {
		return jobs
	}
	return jobs[:limit]
}