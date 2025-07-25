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

type EnvironmentCache struct {
	client      *ghclient.Client
	rateLimiter *RateLimiter
	mu          sync.RWMutex
	
	repoEnvs map[string][]string
	repoEnvsTTL map[string]time.Time
	
	runEnvs map[string]string
	runEnvsTTL map[string]time.Time
	
	deploymentEnvs map[string][]DeploymentEnv
	deploymentEnvsTTL map[string]time.Time
	
	repoEnvCacheTTL   time.Duration
	runEnvCacheTTL    time.Duration
	deploymentCacheTTL time.Duration
}

type DeploymentEnv struct {
	Environment string
	SHA         string
	Ref         string
	State       string
	CreatedAt   time.Time
}

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
		
		repoEnvCacheTTL:   10 * time.Minute,
		runEnvCacheTTL:    2 * time.Minute,
		deploymentCacheTTL: 1 * time.Minute,
	}
}

func (ec *EnvironmentCache) GetRepositoryEnvironments(ctx context.Context, repo string) ([]string, error) {
	ec.mu.RLock()
	if envs, exists := ec.repoEnvs[repo]; exists {
		if time.Now().Before(ec.repoEnvsTTL[repo]) {
			ec.mu.RUnlock()
			return envs, nil
		}
	}
	ec.mu.RUnlock()
	
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
	
	ec.mu.Lock()
	ec.repoEnvs[repo] = envs
	ec.repoEnvsTTL[repo] = time.Now().Add(ec.repoEnvCacheTTL)
	ec.mu.Unlock()
	
	return envs, nil
}

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
	
	env, err := ec.detectEnvironmentForRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
	ec.mu.Lock()
	ec.runEnvs[cacheKey] = env
	ec.runEnvsTTL[cacheKey] = time.Now().Add(ec.runEnvCacheTTL)
	ec.mu.Unlock()
	
	return env, nil
}

func (ec *EnvironmentCache) detectEnvironmentForRun(ctx context.Context, repo string, runID int64) (string, error) {
	env, err := ec.detectFromRecentDeployments(ctx, repo, runID)
	if err == nil && env != "" {
		return env, nil
	}
	
	env, err = ec.detectFromWorkflowRun(ctx, repo, runID)
	if err == nil && env != "" {
		return env, nil
	}
	
	return "", nil
}

func (ec *EnvironmentCache) detectFromRecentDeployments(ctx context.Context, repo string, runID int64) (string, error) {
	if err := ec.rateLimiter.AcquireWorkflowRuns(ctx); err != nil {
		return "", err
	}
	defer ec.rateLimiter.ReleaseWorkflowRuns()
	
	run, _, err := ec.client.GetWorkflowRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
	deployments, err := ec.getRecentDeployments(ctx, repo)
	if err != nil {
		return "", err
	}
	
	runSHA := run.GetHeadSHA()
	for _, deployment := range deployments {
		if deployment.SHA == runSHA {
			return deployment.Environment, nil
		}
	}
	
	return "", nil
}

func (ec *EnvironmentCache) detectFromWorkflowRun(ctx context.Context, repo string, runID int64) (string, error) {
	repoEnvs, err := ec.GetRepositoryEnvironments(ctx, repo)
	if err != nil {
		return "", err
	}
	
	if err := ec.rateLimiter.AcquireWorkflowRuns(ctx); err != nil {
		return "", err
	}
	defer ec.rateLimiter.ReleaseWorkflowRuns()
	
	run, _, err := ec.client.GetWorkflowRun(ctx, repo, runID)
	if err != nil {
		return "", err
	}
	
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
	
	for _, job := range jobs.Jobs {
		if job.GetStatus() == "waiting" {
			jobName := job.GetName()
			
			for _, env := range repoEnvs {
				if strings.Contains(strings.ToLower(jobName), strings.ToLower(env)) {
					return env, nil
				}
			}
		}
	}
	
	workflowName := run.GetName()
	for _, env := range repoEnvs {
		if strings.Contains(strings.ToLower(workflowName), strings.ToLower(env)) {
			return env, nil
		}
	}
	
	return "", nil
}

func (ec *EnvironmentCache) getRecentDeployments(ctx context.Context, repo string) ([]DeploymentEnv, error) {
	ec.mu.RLock()
	if deployments, exists := ec.deploymentEnvs[repo]; exists {
		if time.Now().Before(ec.deploymentEnvsTTL[repo]) {
			ec.mu.RUnlock()
			return deployments, nil
		}
	}
	ec.mu.RUnlock()
	
	if err := ec.rateLimiter.AcquireDeployments(ctx); err != nil {
		return nil, err
	}
	defer ec.rateLimiter.ReleaseDeployments()
	
	opts := &github.DeploymentsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 20,
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
			State:       "pending",
			CreatedAt:   deployment.GetCreatedAt().Time,
		})
	}
	
	ec.mu.Lock()
	ec.deploymentEnvs[repo] = deploymentEnvs
	ec.deploymentEnvsTTL[repo] = time.Now().Add(ec.deploymentCacheTTL)
	ec.mu.Unlock()
	
	return deploymentEnvs, nil
}

func (ec *EnvironmentCache) CleanupExpiredCache() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	now := time.Now()
	
	for repo, ttl := range ec.repoEnvsTTL {
		if now.After(ttl) {
			delete(ec.repoEnvs, repo)
			delete(ec.repoEnvsTTL, repo)
		}
	}
	
	for key, ttl := range ec.runEnvsTTL {
		if now.After(ttl) {
			delete(ec.runEnvs, key)
			delete(ec.runEnvsTTL, key)
		}
	}
	
	for repo, ttl := range ec.deploymentEnvsTTL {
		if now.After(ttl) {
			delete(ec.deploymentEnvs, repo)
			delete(ec.deploymentEnvsTTL, repo)
		}
	}
}

func (ec *EnvironmentCache) GetCacheStats() map[string]int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	return map[string]int{
		"repo_envs":        len(ec.repoEnvs),
		"run_envs":         len(ec.runEnvs),
		"deployment_envs":  len(ec.deploymentEnvs),
	}
}