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

func (wp *WorkerPool) ScanRepositories(ctx context.Context, repos []*github.Repository, progressChan chan<- ScanProgress, progress *ScanProgress) ([]scanner.JobStatus, error) {
	if len(repos) == 0 {
		return []scanner.JobStatus{}, nil
	}

	repoChan := make(chan *github.Repository, len(repos))
	resultChan := make(chan scanner.RepoScanResult, len(repos))

	for i := 0; i < wp.maxWorkers; i++ {
		go func(workerID int) {
			for repo := range repoChan {
				jobs, err := wp.scanner.ScanRepository(ctx, repo)
				
				resultChan <- scanner.RepoScanResult{Jobs: jobs, Err: err}
				
				delay := BaseWorkerDelay + time.Duration(workerID)*WorkerDelayIncrement
				time.Sleep(delay)
			}
		}(i)
	}

	for _, repo := range repos {
		repoChan <- repo
	}
	close(repoChan)

	var allJobs []scanner.JobStatus
	completedRepos := 0
	
	for i := 0; i < len(repos); i++ {
		result := <-resultChan
		if result.Err == nil {
			allJobs = append(allJobs, result.Jobs...)
		}
		completedRepos++
		
		if progress != nil {
			progress.CompletedRepos = completedRepos
			
			if progressChan != nil {
				progressChan <- *progress
			}
		}
	}

	return allJobs, nil
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