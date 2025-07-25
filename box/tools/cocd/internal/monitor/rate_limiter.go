package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ResourceType string

const (
	ResourceRepoList      ResourceType = "repo_list"
	ResourceWorkflowRuns  ResourceType = "workflow_runs"
	ResourceWorkflowJobs  ResourceType = "workflow_jobs"
	ResourceEnvironments  ResourceType = "environments"
	ResourceDeployments   ResourceType = "deployments"
)

var defaultLimits = map[ResourceType]int{
	ResourceRepoList:      1,
	ResourceWorkflowRuns:  1,
	ResourceWorkflowJobs:  1,
	ResourceEnvironments:  1,
	ResourceDeployments:   1,
}

const (
	DefaultBatchSize     = 2
	DefaultBatchInterval = 800 * time.Millisecond
)

type Config struct {
	Limits        map[ResourceType]int
	BatchSize     int
	BatchInterval time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Limits:        defaultLimits,
		BatchSize:     DefaultBatchSize,
		BatchInterval: DefaultBatchInterval,
	}
}

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

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	<-s.ch
}

type RateLimiter struct {
	semaphores    map[ResourceType]*Semaphore
	mu            sync.RWMutex
	batchSize     int
	batchInterval time.Duration
}

func NewRateLimiter() *RateLimiter {
	return NewRateLimiterWithConfig(DefaultConfig())
}

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

func (rl *RateLimiter) Acquire(ctx context.Context, resourceType ResourceType) error {
	rl.mu.RLock()
	semaphore, exists := rl.semaphores[resourceType]
	rl.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	
	return semaphore.Acquire(ctx)
}

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

func (rl *RateLimiter) AcquireRepoList(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceRepoList)
}

func (rl *RateLimiter) ReleaseRepoList() {
	rl.Release(ResourceRepoList)
}

func (rl *RateLimiter) AcquireWorkflowRuns(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceWorkflowRuns)
}

func (rl *RateLimiter) ReleaseWorkflowRuns() {
	rl.Release(ResourceWorkflowRuns)
}

func (rl *RateLimiter) AcquireWorkflowJobs(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceWorkflowJobs)
}

func (rl *RateLimiter) ReleaseWorkflowJobs() {
	rl.Release(ResourceWorkflowJobs)
}

func (rl *RateLimiter) AcquireEnvironments(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceEnvironments)
}

func (rl *RateLimiter) ReleaseEnvironments() {
	rl.Release(ResourceEnvironments)
}

func (rl *RateLimiter) AcquireDeployments(ctx context.Context) error {
	return rl.Acquire(ctx, ResourceDeployments)
}

func (rl *RateLimiter) ReleaseDeployments() {
	rl.Release(ResourceDeployments)
}

type BatchProcessor[T any] func([]T) error

func (rl *RateLimiter) BatchProcess(ctx context.Context, items []interface{}, processFunc func([]interface{}) error) error {
	return rl.batchProcessInternal(ctx, items, rl.batchSize, rl.batchInterval, processFunc)
}

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

func (rl *RateLimiter) GetBatchSize() int {
	return rl.batchSize
}

func (rl *RateLimiter) GetBatchInterval() time.Duration {
	return rl.batchInterval
}

func (rl *RateLimiter) UpdateLimit(resourceType ResourceType, newLimit int) error {
	if newLimit <= 0 {
		return fmt.Errorf("limit must be positive, got %d", newLimit)
	}
	
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.semaphores[resourceType] = NewSemaphore(newLimit)
	return nil
}

func (rl *RateLimiter) GetLimit(resourceType ResourceType) (int, error) {
	rl.mu.RLock()
	semaphore, exists := rl.semaphores[resourceType]
	rl.mu.RUnlock()
	
	if !exists {
		return 0, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	
	return cap(semaphore.ch), nil
}

func (rl *RateLimiter) GetResourceTypes() []ResourceType {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	resourceTypes := make([]ResourceType, 0, len(rl.semaphores))
	for resourceType := range rl.semaphores {
		resourceTypes = append(resourceTypes, resourceType)
	}
	return resourceTypes
}