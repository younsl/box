package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceType represents different API resource types for rate limiting
type ResourceType string

const (
	ResourceRepoList      ResourceType = "repo_list"
	ResourceWorkflowRuns  ResourceType = "workflow_runs"
	ResourceWorkflowJobs  ResourceType = "workflow_jobs"
	ResourceEnvironments  ResourceType = "environments"
	ResourceDeployments   ResourceType = "deployments"
)

// Default rate limiting configuration for GHES optimization
var defaultLimits = map[ResourceType]int{
	ResourceRepoList:      1, // Maximum concurrent repository list calls
	ResourceWorkflowRuns:  2, // Maximum concurrent workflow runs calls
	ResourceWorkflowJobs:  1, // Maximum concurrent workflow job calls
	ResourceEnvironments:  1, // Maximum concurrent environment calls
	ResourceDeployments:   1, // Maximum concurrent deployment calls
}

const (
	// Batch processing configuration
	DefaultBatchSize     = 3                   // Items processed per batch
	DefaultBatchInterval = 400 * time.Millisecond // Interval between batches
)

// Config holds rate limiter configuration
type Config struct {
	Limits        map[ResourceType]int
	BatchSize     int
	BatchInterval time.Duration
}

// DefaultConfig returns the default rate limiter configuration
func DefaultConfig() *Config {
	return &Config{
		Limits:        defaultLimits,
		BatchSize:     DefaultBatchSize,
		BatchInterval: DefaultBatchInterval,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive, got %d", c.BatchSize)
	}
	if c.BatchInterval <= 0 {
		return fmt.Errorf("batch interval must be positive, got %v", c.BatchInterval)
	}
	for resourceType, limit := range c.Limits {
		if limit <= 0 {
			return fmt.Errorf("limit for %s must be positive, got %d", resourceType, limit)
		}
	}
	return nil
}

// Semaphore represents a semaphore for rate limiting
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given capacity
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire acquires a token from the semaphore
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a token back to the semaphore
func (s *Semaphore) Release() {
	<-s.ch
}

// RateLimiter provides rate limiting for API calls to reduce GHES load
type RateLimiter struct {
	semaphores    map[ResourceType]*Semaphore
	mu            sync.RWMutex
	batchSize     int
	batchInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with default configuration
func NewRateLimiter() *RateLimiter {
	return NewRateLimiterWithConfig(DefaultConfig())
}

// NewRateLimiterWithConfig creates a new rate limiter with the given configuration
func NewRateLimiterWithConfig(config *Config) *RateLimiter {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid rate limiter config: %v", err))
	}
	
	semaphores := make(map[ResourceType]*Semaphore)
	for resourceType, limit := range config.Limits {
		semaphores[resourceType] = NewSemaphore(limit)
	}
	
	return &RateLimiter{
		semaphores:    semaphores,
		batchSize:     config.BatchSize,
		batchInterval: config.BatchInterval,
	}
}

// Acquire acquires a token for the specified resource type
func (rl *RateLimiter) Acquire(ctx context.Context, resourceType ResourceType) error {
	rl.mu.RLock()
	semaphore, exists := rl.semaphores[resourceType]
	rl.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	
	return semaphore.Acquire(ctx)
}

// Release releases a token for the specified resource type
func (rl *RateLimiter) Release(resourceType ResourceType) error {
	rl.mu.RLock()
	semaphore, exists := rl.semaphores[resourceType]
	rl.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	
	semaphore.Release()
	return nil
}

// AcquireRepoList acquires a token for repository list API calls
func (rl *RateLimiter) AcquireRepoList(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceRepoList)
}

// ReleaseRepoList releases a token for repository list API calls
func (rl *RateLimiter) ReleaseRepoList() {
	rl.Release(ResourceRepoList)
}

// AcquireWorkflowRuns acquires a token for workflow runs API calls
func (rl *RateLimiter) AcquireWorkflowRuns(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceWorkflowRuns)
}

// ReleaseWorkflowRuns releases a token for workflow runs API calls
func (rl *RateLimiter) ReleaseWorkflowRuns() {
	rl.Release(ResourceWorkflowRuns)
}

// AcquireWorkflowJobs acquires a token for workflow jobs API calls
func (rl *RateLimiter) AcquireWorkflowJobs(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceWorkflowJobs)
}

// ReleaseWorkflowJobs releases a token for workflow jobs API calls
func (rl *RateLimiter) ReleaseWorkflowJobs() {
	rl.Release(ResourceWorkflowJobs)
}

// AcquireEnvironments acquires a token for environments API calls
func (rl *RateLimiter) AcquireEnvironments(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceEnvironments)
}

// ReleaseEnvironments releases a token for environments API calls
func (rl *RateLimiter) ReleaseEnvironments() {
	rl.Release(ResourceEnvironments)
}

// AcquireDeployments acquires a token for deployments API calls
func (rl *RateLimiter) AcquireDeployments(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceDeployments)
}

// ReleaseDeployments releases a token for deployments API calls
func (rl *RateLimiter) ReleaseDeployments() {
	rl.Release(ResourceDeployments)
}

// BatchProcessor defines a function type for processing batches
type BatchProcessor[T any] func([]T) error

// BatchProcess processes items in batches with rate limiting using generics
func (rl *RateLimiter) BatchProcess(ctx context.Context, items []interface{}, processFunc func([]interface{}) error) error {
	return rl.batchProcessInternal(ctx, items, rl.batchSize, rl.batchInterval, processFunc)
}

// BatchProcessTyped processes typed items in batches with rate limiting
func BatchProcessTyped[T any](ctx context.Context, items []T, batchSize int, batchInterval time.Duration, processFunc BatchProcessor[T]) error {
	for i := 0; i < len(items); i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		if err := processFunc(batch); err != nil {
			return err
		}
		
		if end < len(items) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(batchInterval):
			}
		}
	}
	
	return nil
}

// batchProcessInternal is the internal implementation of batch processing
func (rl *RateLimiter) batchProcessInternal(ctx context.Context, items []interface{}, batchSize int, batchInterval time.Duration, processFunc func([]interface{}) error) error {
	for i := 0; i < len(items); i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		if err := processFunc(batch); err != nil {
			return err
		}
		
		if end < len(items) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(batchInterval):
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

// UpdateLimit updates the limit for a specific resource type
func (rl *RateLimiter) UpdateLimit(resourceType ResourceType, newLimit int) error {
	if newLimit <= 0 {
		return fmt.Errorf("limit must be positive, got %d", newLimit)
	}
	
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.semaphores[resourceType] = NewSemaphore(newLimit)
	return nil
}

// GetLimit returns the current limit for a specific resource type
func (rl *RateLimiter) GetLimit(resourceType ResourceType) (int, error) {
	rl.mu.RLock()
	semaphore, exists := rl.semaphores[resourceType]
	rl.mu.RUnlock()
	
	if !exists {
		return 0, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	
	return cap(semaphore.ch), nil
}

// GetResourceTypes returns all available resource types
func (rl *RateLimiter) GetResourceTypes() []ResourceType {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	resourceTypes := make([]ResourceType, 0, len(rl.semaphores))
	for resourceType := range rl.semaphores {
		resourceTypes = append(resourceTypes, resourceType)
	}
	return resourceTypes
}