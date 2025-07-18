package monitor

import (
	"context"
	"time"
)

// RateLimiter provides rate limiting for API calls to reduce GHES load
type RateLimiter struct {
	// Rate limiting channels
	repoListChan        chan struct{}
	workflowRunsChan    chan struct{}
	workflowJobsChan    chan struct{}
	environmentsChan    chan struct{}
	deploymentsChan     chan struct{}
	
	// Batch processing
	batchSize     int
	batchInterval time.Duration
}

// NewRateLimiter creates a new rate limiter optimized for GHES
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		// Conservative rate limits to avoid overwhelming GHES
		repoListChan:        make(chan struct{}, 1),   // 1 concurrent repo list call
		workflowRunsChan:    make(chan struct{}, 3),   // 3 concurrent workflow runs calls
		workflowJobsChan:    make(chan struct{}, 2),   // 2 concurrent job calls
		environmentsChan:    make(chan struct{}, 1),   // 1 concurrent environment call
		deploymentsChan:     make(chan struct{}, 1),   // 1 concurrent deployment call
		
		batchSize:     5,                  // Process 5 items at a time for lighter load
		batchInterval: 200 * time.Millisecond, // 200ms between batches for better spacing
	}
}

// AcquireRepoList acquires a token for repository list API calls
func (rl *RateLimiter) AcquireRepoList(ctx context.Context) error {
	select {
	case rl.repoListChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseRepoList releases a token for repository list API calls
func (rl *RateLimiter) ReleaseRepoList() {
	<-rl.repoListChan
}

// AcquireWorkflowRuns acquires a token for workflow runs API calls
func (rl *RateLimiter) AcquireWorkflowRuns(ctx context.Context) error {
	select {
	case rl.workflowRunsChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseWorkflowRuns releases a token for workflow runs API calls
func (rl *RateLimiter) ReleaseWorkflowRuns() {
	<-rl.workflowRunsChan
}

// AcquireWorkflowJobs acquires a token for workflow jobs API calls
func (rl *RateLimiter) AcquireWorkflowJobs(ctx context.Context) error {
	select {
	case rl.workflowJobsChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseWorkflowJobs releases a token for workflow jobs API calls
func (rl *RateLimiter) ReleaseWorkflowJobs() {
	<-rl.workflowJobsChan
}

// AcquireEnvironments acquires a token for environments API calls
func (rl *RateLimiter) AcquireEnvironments(ctx context.Context) error {
	select {
	case rl.environmentsChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseEnvironments releases a token for environments API calls
func (rl *RateLimiter) ReleaseEnvironments() {
	<-rl.environmentsChan
}

// AcquireDeployments acquires a token for deployments API calls
func (rl *RateLimiter) AcquireDeployments(ctx context.Context) error {
	select {
	case rl.deploymentsChan <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReleaseDeployments releases a token for deployments API calls
func (rl *RateLimiter) ReleaseDeployments() {
	<-rl.deploymentsChan
}

// BatchProcess processes items in batches with rate limiting
func (rl *RateLimiter) BatchProcess(ctx context.Context, items []interface{}, processFunc func([]interface{}) error) error {
	for i := 0; i < len(items); i += rl.batchSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		end := i + rl.batchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		if err := processFunc(batch); err != nil {
			return err
		}
		
		// Wait before processing next batch
		if end < len(items) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rl.batchInterval):
			}
		}
	}
	
	return nil
}

// GetBatchSize returns the current batch size
func (rl *RateLimiter) GetBatchSize() int {
	return rl.batchSize
}

// GetBatchInterval returns the current batch interval
func (rl *RateLimiter) GetBatchInterval() time.Duration {
	return rl.batchInterval
}