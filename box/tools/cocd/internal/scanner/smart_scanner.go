package scanner

import (
	"context"
	"time"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/internal/github"
)

// EnvironmentCache interface to avoid circular dependency
type EnvironmentCache interface {
	GetWorkflowRunEnvironment(ctx context.Context, repo string, runID int64) (string, error)
}

// SmartScanner scans repositories with intelligent filtering
type SmartScanner struct {
	client      *ghclient.Client
	environment string
	envCache    EnvironmentCache
}

// NewSmartScanner creates a new smart scanner
func NewSmartScanner(client *ghclient.Client, environment string, envCache EnvironmentCache) *SmartScanner {
	return &SmartScanner{
		client:      client,
		environment: environment,
		envCache:    envCache,
	}
}

// ScanRepository scans a single repository for waiting jobs only (highly optimized)
func (s *SmartScanner) ScanRepository(ctx context.Context, repo *github.Repository) ([]JobStatus, error) {
	if repo.GetArchived() || repo.GetDisabled() {
		return nil, nil
	}

	// Skip repositories without recent activity (last 7 days)
	if repo.PushedAt == nil || time.Since(repo.PushedAt.Time) > 7*24*time.Hour {
		return nil, nil
	}

	var waitingJobs []JobStatus

	// API filter: only get "waiting" status runs
	opts := &github.ListWorkflowRunsOptions{
		Status: "waiting",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	runs, _, err := s.client.ListWorkflowRuns(ctx, repo.GetName(), opts)
	if err != nil {
		return nil, nil
	}

	// Process only waiting runs (already filtered by API)
	for _, run := range runs.WorkflowRuns {
		env := ""
		if s.envCache != nil {
			env, _ = s.envCache.GetWorkflowRunEnvironment(ctx, repo.GetName(), run.GetID())
		}
		
		if s.environment == "" || env == s.environment {
			waitingJobs = append(waitingJobs, JobStatus{
				ID:           run.GetID(),
				Name:         run.GetName(),
				RunID:        run.GetID(),
				RunNumber:    run.GetRunNumber(),
				Status:       run.GetStatus(),
				Conclusion:   run.GetConclusion(),
				StartedAt:    run.CreatedAt.GetTime(),
				CompletedAt:  run.UpdatedAt.GetTime(),
				Environment:  env,
				WorkflowName: run.GetName(),
				Branch:       run.GetHeadBranch(),
				Event:        run.GetEvent(),
				Actor:        run.GetActor().GetLogin(),
				Repository:   repo.GetName(),
			})
		}
	}

	return waitingJobs, nil
}