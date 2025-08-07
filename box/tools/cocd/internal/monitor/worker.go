package monitor

import (
	"context"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/younsl/cocd/internal/scanner"
)

const (
	BaseWorkerDelay      = 1000 * time.Millisecond
	WorkerDelayIncrement = 500 * time.Millisecond
)

type WorkerPool struct {
	maxWorkers int
	scanner    scanner.Scanner
}

func NewWorkerPool(maxWorkers int, sc scanner.Scanner) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		scanner:    sc,
	}
}

// JobUpdate represents a streaming update of scan results
type JobUpdate struct {
	Jobs          []scanner.JobStatus // New jobs found
	CompletedRepo string              // Name of completed repository
	Progress      ScanProgress        // Updated progress
	Error         error               // Any error that occurred
}

func (wp *WorkerPool) ScanRepositories(ctx context.Context, repos []*github.Repository, progressChan chan<- ScanProgress, progress *ScanProgress) ([]scanner.JobStatus, error) {
	if len(repos) == 0 {
		return []scanner.JobStatus{}, nil
	}

	var allJobs []scanner.JobStatus
	completedRepos := 0
	
	// Sequential scanning for reduced server load
	for _, repo := range repos {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return allJobs, ctx.Err()
		default:
		}
		
		// Record start time for adaptive delay
		startTime := time.Now()
		
		// Scan repository
		jobs, err := wp.scanner.ScanRepository(ctx, repo)
		
		// Calculate response time
		responseTime := time.Since(startTime)
		
		// Add successful jobs to results
		if err == nil {
			allJobs = append(allJobs, jobs...)
		}
		
		completedRepos++
		
		// Update progress immediately after each repository scan
		if progress != nil {
			progress.CompletedRepos = completedRepos
			
			if progressChan != nil {
				select {
				case progressChan <- *progress:
				case <-ctx.Done():
					return allJobs, ctx.Err()
				}
			}
		}
		
		// Adaptive delay based on server response time
		var delay time.Duration
		if responseTime > 2*time.Second {
			delay = 3 * time.Second // Server is slow, use longer delay
		} else if responseTime > 1*time.Second {
			delay = BaseWorkerDelay // Normal response time
		} else {
			delay = BaseWorkerDelay / 2 // Fast response, shorter delay
		}
		
		// Apply delay before next repository (except for last one)
		if completedRepos < len(repos) {
			select {
			case <-ctx.Done():
				return allJobs, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return allJobs, nil
}

// ScanRepositoriesStreaming scans repositories and streams results in real-time
func (wp *WorkerPool) ScanRepositoriesStreaming(ctx context.Context, repos []*github.Repository, jobUpdateChan chan<- JobUpdate, progress *ScanProgress) error {
	if len(repos) == 0 {
		return nil
	}

	completedRepos := 0
	
	// Sequential scanning for reduced server load with real-time updates
	for _, repo := range repos {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Record start time for adaptive delay
		startTime := time.Now()
		
		// Scan repository
		jobs, err := wp.scanner.ScanRepository(ctx, repo)
		
		// Calculate response time
		responseTime := time.Since(startTime)
		
		completedRepos++
		
		// Update progress
		if progress != nil {
			progress.CompletedRepos = completedRepos
		}
		
		// Send immediate update with results from this repository
		update := JobUpdate{
			Jobs:          jobs,
			CompletedRepo: repo.GetName(),
			Error:         err,
		}
		if progress != nil {
			update.Progress = *progress
		}
		
		select {
		case jobUpdateChan <- update:
		case <-ctx.Done():
			return ctx.Err()
		}
		
		// Adaptive delay based on server response time
		var delay time.Duration
		if responseTime > 2*time.Second {
			delay = 3 * time.Second // Server is slow, use longer delay
		} else if responseTime > 1*time.Second {
			delay = BaseWorkerDelay // Normal response time
		} else {
			delay = BaseWorkerDelay / 2 // Fast response, shorter delay
		}
		
		// Apply delay before next repository (except for last one)
		if completedRepos < len(repos) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil
}

func SortJobsByTime(jobs []scanner.JobStatus, newest bool) {
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].StartedAt == nil && jobs[j].StartedAt == nil {
			return false
		}
		if jobs[i].StartedAt == nil {
			return !newest
		}
		if jobs[j].StartedAt == nil {
			return newest
		}
		
		if newest {
			return jobs[i].StartedAt.After(*jobs[j].StartedAt)
		} else {
			return jobs[i].StartedAt.Before(*jobs[j].StartedAt)
		}
	})
}

func LimitJobs(jobs []scanner.JobStatus, limit int) []scanner.JobStatus {
	if len(jobs) <= limit {
		return jobs
	}
	return jobs[:limit]
}