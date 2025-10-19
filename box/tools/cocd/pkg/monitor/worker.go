package monitor

import (
	"context"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/younsl/cocd/pkg/scanner"
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
	
	for _, repo := range repos {
		select {
		case <-ctx.Done():
			return allJobs, ctx.Err()
		default:
		}
		
		startTime := time.Now()
		
		jobs, err := wp.scanner.ScanRepository(ctx, repo)
		
		responseTime := time.Since(startTime)
		
		if err == nil {
			allJobs = append(allJobs, jobs...)
		}
		
		completedRepos++
		
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
		
		var delay time.Duration
		if responseTime > 2*time.Second {
			delay = 3 * time.Second
		} else if responseTime > 1*time.Second {
			delay = BaseWorkerDelay
		} else {
			delay = BaseWorkerDelay / 2
		}
		
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


func (wp *WorkerPool) ScanRepositoriesStreamingWithTracker(ctx context.Context, repos []*github.Repository, jobUpdateChan chan<- JobUpdate, progressTracker *ProgressTracker) error {
	if len(repos) == 0 {
		return nil
	}

	completedRepos := 0
	
	// Sequential scanning for reduced server load with real-time updates
	for _, repo := range repos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		startTime := time.Now()
		
		jobs, err := wp.scanner.ScanRepository(ctx, repo)
		
		responseTime := time.Since(startTime)
		
		completedRepos++
		
			if progressTracker != nil {
			progressTracker.UpdateCompleted(completedRepos)
		}
		
		update := JobUpdate{
			Jobs:          jobs,
			CompletedRepo: repo.GetName(),
			Error:         err,
		}
		if progressTracker != nil {
			update.Progress = progressTracker.GetProgress()
		}
		
		select {
		case jobUpdateChan <- update:
		case <-ctx.Done():
			return ctx.Err()
		}
		
		var delay time.Duration
		if responseTime > 2*time.Second {
			delay = 3 * time.Second
		} else if responseTime > 1*time.Second {
			delay = BaseWorkerDelay
		} else {
			delay = BaseWorkerDelay / 2
		}
		
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