package scanner

import (
	"context"
	"time"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/internal/github"
)

type EnvironmentCache interface {
	GetWorkflowRunEnvironment(ctx context.Context, repo string, runID int64) (string, error)
}

type SmartScanner struct {
	client      *ghclient.Client
	environment string
	envCache    EnvironmentCache
}

func NewSmartScanner(client *ghclient.Client, environment string, envCache EnvironmentCache) *SmartScanner {
	return &SmartScanner{
		client:      client,
		environment: environment,
		envCache:    envCache,
	}
}

func (s *SmartScanner) ScanRepository(ctx context.Context, repo *github.Repository) ([]JobStatus, error) {
	if repo.GetArchived() || repo.GetDisabled() {
		return nil, nil
	}

	if repo.PushedAt == nil || time.Since(repo.PushedAt.Time) > 3*24*time.Hour {
		return nil, nil
	}

	var waitingJobs []JobStatus

	opts := &github.ListWorkflowRunsOptions{
		Status: "waiting",
		ListOptions: github.ListOptions{
			PerPage: 5,
		},
	}

	runs, _, err := s.client.ListWorkflowRuns(ctx, repo.GetName(), opts)
	if err != nil {
		return nil, nil
	}

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