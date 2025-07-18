package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/internal/github"
)

// EnvironmentCache manages environment detection with caching to reduce API calls
type EnvironmentCache struct {
	client      *ghclient.Client
	rateLimiter *RateLimiter
	mu          sync.RWMutex
	
	// Cache for repository environments
	repoEnvs map[string][]string
	repoEnvsTTL map[string]time.Time
	
	// Cache for workflow run environments
	runEnvs map[string]string
	runEnvsTTL map[string]time.Time
	
	// Cache for deployment environments
	deploymentEnvs map[string][]DeploymentEnv
	deploymentEnvsTTL map[string]time.Time
	
	// Cache TTL settings
	repoEnvCacheTTL   time.Duration
	runEnvCacheTTL    time.Duration
	deploymentCacheTTL time.Duration
}

// DeploymentEnv represents deployment environment information
type DeploymentEnv struct {
	Environment string
	SHA         string
	Ref         string
	State       string
	CreatedAt   time.Time
}

// NewEnvironmentCache creates a new environment cache
func NewEnvironmentCache(client *ghclient.Client) *EnvironmentCache {
	return &EnvironmentCache{
		client:      client,
		rateLimiter: NewRateLimiter(),
		repoEnvs: make(map[string][]string),
		repoEnvsTTL: make(map[string]time.Time),
		runEnvs: make(map[string]string),
		runEnvsTTL: make(map[string]time.Time),
		deploymentEnvs: make(map[string][]DeploymentEnv),
		deploymentEnvsTTL: make(map[string]time.Time),
		
		// Conservative TTL settings to reduce GHES load
		repoEnvCacheTTL:   10 * time.Minute,  // Repository environments change infrequently
		runEnvCacheTTL:    2 * time.Minute,   // Workflow run environments are more dynamic
		deploymentCacheTTL: 1 * time.Minute,  // Deployment states change frequently
	}
}

// GetRepositoryEnvironments gets environments for a repository with caching
func (ec *EnvironmentCache) GetRepositoryEnvironments(ctx context.Context, repo string) ([]string, error) {
	ec.mu.RLock()
	if envs, exists := ec.repoEnvs[repo]; exists {
		if time.Now().Before(ec.repoEnvsTTL[repo]) {
			ec.mu.RUnlock()
			return envs, nil
		}
	}
	ec.mu.RUnlock()
	
	// Fetch from API with rate limiting
	if err := ec.rateLimiter.AcquireEnvironments(ctx); err != nil {
		return nil, err
	}
	defer ec.rateLimiter.ReleaseEnvironments()
	
	envResponse, _, err := ec.client.ListEnvironments(ctx, repo)
	if err != nil {
		return nil, err
	}
	
	var envs []string
	if envResponse != nil {
		for _, env := range envResponse.Environments {
			envs = append(envs, env.GetName())
		}
	}
	
	// Cache the result
	ec.mu.Lock()
	ec.repoEnvs[repo] = envs
	ec.repoEnvsTTL[repo] = time.Now().Add(ec.repoEnvCacheTTL)
	ec.mu.Unlock()
	
	return envs, nil
}

// GetWorkflowRunEnvironment gets environment for a workflow run with caching and deployment detection
func (ec *EnvironmentCache) GetWorkflowRunEnvironment(ctx context.Context, repo string, runID int64) (string, error) {
	cacheKey := fmt.Sprintf("%s:%d", repo, runID)
	
	ec.mu.RLock()
	if env, exists := ec.runEnvs[cacheKey]; exists {
		if time.Now().Before(ec.runEnvsTTL[cacheKey]) {
			ec.mu.RUnlock()
			return env, nil
		}
	}
	ec.mu.RUnlock()
	
	// Detect environment using multiple methods
	env, err := ec.detectEnvironmentForRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
	// Cache the result
	ec.mu.Lock()
	ec.runEnvs[cacheKey] = env
	ec.runEnvsTTL[cacheKey] = time.Now().Add(ec.runEnvCacheTTL)
	ec.mu.Unlock()
	
	return env, nil
}

// detectEnvironmentForRun detects environment using optimized API calls
func (ec *EnvironmentCache) detectEnvironmentForRun(ctx context.Context, repo string, runID int64) (string, error) {
	// Method 1: Check recent deployments for this repository
	env, err := ec.detectFromRecentDeployments(ctx, repo, runID)
	if err == nil && env != "" {
		return env, nil
	}
	
	// Method 2: Check workflow run details and jobs
	env, err = ec.detectFromWorkflowRun(ctx, repo, runID)
	if err == nil && env != "" {
		return env, nil
	}
	
	return "", nil
}

// detectFromRecentDeployments detects environment from recent deployments
func (ec *EnvironmentCache) detectFromRecentDeployments(ctx context.Context, repo string, runID int64) (string, error) {
	// Get workflow run details first with rate limiting
	if err := ec.rateLimiter.AcquireWorkflowRuns(ctx); err != nil {
		return "", err
	}
	defer ec.rateLimiter.ReleaseWorkflowRuns()
	
	run, _, err := ec.client.GetWorkflowRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
	// Get recent deployments with caching
	deployments, err := ec.getRecentDeployments(ctx, repo)
	if err != nil {
		return "", err
	}
	
	// Match deployment SHA with workflow run SHA
	runSHA := run.GetHeadSHA()
	for _, deployment := range deployments {
		if deployment.SHA == runSHA {
			return deployment.Environment, nil
		}
	}
	
	return "", nil
}

// detectFromWorkflowRun detects environment from workflow run and jobs
func (ec *EnvironmentCache) detectFromWorkflowRun(ctx context.Context, repo string, runID int64) (string, error) {
	// Get repository environments
	repoEnvs, err := ec.GetRepositoryEnvironments(ctx, repo)
	if err != nil {
		return "", err
	}
	
	// Get workflow run details with rate limiting
	if err := ec.rateLimiter.AcquireWorkflowRuns(ctx); err != nil {
		return "", err
	}
	defer ec.rateLimiter.ReleaseWorkflowRuns()
	
	run, _, err := ec.client.GetWorkflowRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
	// Get jobs for this run (only if we have waiting jobs) with rate limiting
	if err := ec.rateLimiter.AcquireWorkflowJobs(ctx); err != nil {
		return "", err
	}
	defer ec.rateLimiter.ReleaseWorkflowJobs()
	
	jobs, _, err := ec.client.ListWorkflowJobs(ctx, repo, runID, &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 50},
	})
	if err != nil {
		return "", err
	}
	
	// Check if any job is waiting and try to match with repository environments
	for _, job := range jobs.Jobs {
		if job.GetStatus() == "waiting" {
			jobName := job.GetName()
			
			// Try to match job name with repository environments
			for _, env := range repoEnvs {
				if strings.Contains(strings.ToLower(jobName), strings.ToLower(env)) {
					return env, nil
				}
			}
		}
	}
	
	// Fallback: try to match workflow name with environments
	workflowName := run.GetName()
	for _, env := range repoEnvs {
		if strings.Contains(strings.ToLower(workflowName), strings.ToLower(env)) {
			return env, nil
		}
	}
	
	return "", nil
}

// getRecentDeployments gets recent deployments with caching
func (ec *EnvironmentCache) getRecentDeployments(ctx context.Context, repo string) ([]DeploymentEnv, error) {
	ec.mu.RLock()
	if deployments, exists := ec.deploymentEnvs[repo]; exists {
		if time.Now().Before(ec.deploymentEnvsTTL[repo]) {
			ec.mu.RUnlock()
			return deployments, nil
		}
	}
	ec.mu.RUnlock()
	
	// Fetch recent deployments from API with rate limiting
	if err := ec.rateLimiter.AcquireDeployments(ctx); err != nil {
		return nil, err
	}
	defer ec.rateLimiter.ReleaseDeployments()
	
	opts := &github.DeploymentsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 20, // Limit to reduce API load
		},
	}
	
	deployments, _, err := ec.client.ListDeployments(ctx, repo, opts)
	if err != nil {
		return nil, err
	}
	
	var deploymentEnvs []DeploymentEnv
	for _, deployment := range deployments {
		deploymentEnvs = append(deploymentEnvs, DeploymentEnv{
			Environment: deployment.GetEnvironment(),
			SHA:         deployment.GetSHA(),
			Ref:         deployment.GetRef(),
			State:       "pending", // Default state for pending deployments
			CreatedAt:   deployment.GetCreatedAt().Time,
		})
	}
	
	// Cache the result
	ec.mu.Lock()
	ec.deploymentEnvs[repo] = deploymentEnvs
	ec.deploymentEnvsTTL[repo] = time.Now().Add(ec.deploymentCacheTTL)
	ec.mu.Unlock()
	
	return deploymentEnvs, nil
}

// CleanupExpiredCache removes expired cache entries
func (ec *EnvironmentCache) CleanupExpiredCache() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	now := time.Now()
	
	// Cleanup repository environments cache
	for repo, ttl := range ec.repoEnvsTTL {
		if now.After(ttl) {
			delete(ec.repoEnvs, repo)
			delete(ec.repoEnvsTTL, repo)
		}
	}
	
	// Cleanup workflow run environments cache
	for key, ttl := range ec.runEnvsTTL {
		if now.After(ttl) {
			delete(ec.runEnvs, key)
			delete(ec.runEnvsTTL, key)
		}
	}
	
	// Cleanup deployment environments cache
	for repo, ttl := range ec.deploymentEnvsTTL {
		if now.After(ttl) {
			delete(ec.deploymentEnvs, repo)
			delete(ec.deploymentEnvsTTL, repo)
		}
	}
}

// GetCacheStats returns cache statistics
func (ec *EnvironmentCache) GetCacheStats() map[string]int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	return map[string]int{
		"repo_envs":        len(ec.repoEnvs),
		"run_envs":         len(ec.runEnvs),
		"deployment_envs":  len(ec.deploymentEnvs),
	}
}