package scanner

import (
	"context"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/internal/github"
)

// RecentJobsScanner scans for recent jobs
type RecentJobsScanner struct {
	client *ghclient.Client
}

// NewRecentJobsScanner creates a new recent jobs scanner
func NewRecentJobsScanner(client *ghclient.Client) *RecentJobsScanner {
	return &RecentJobsScanner{
		client: client,
	}
}

// ScanRepository scans a single repository for recent jobs
func (s *RecentJobsScanner) ScanRepository(ctx context.Context, repo *github.Repository) ([]JobStatus, error) {
	if repo.GetArchived() || repo.GetDisabled() {
		return nil, nil
	}

	var recentJobs []JobStatus

	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}

	runs, _, err := s.client.ListWorkflowRuns(ctx, repo.GetName(), opts)
	if err != nil {
		return nil, nil
	}

	for _, run := range runs.WorkflowRuns {
		status := run.GetStatus()
		conclusion := run.GetConclusion()
		
		displayStatus := status
		if status == "completed" && conclusion != "" {
			displayStatus = conclusion
		}
		
		recentJobs = append(recentJobs, JobStatus{
			ID:           run.GetID(),
			Name:         run.GetName(),
			RunID:        run.GetID(),
			RunNumber:    run.GetRunNumber(),
			Status:       displayStatus,
			Conclusion:   conclusion,
			StartedAt:    run.CreatedAt.GetTime(),
			CompletedAt:  run.UpdatedAt.GetTime(),
			Environment:  "",
			WorkflowName: run.GetName(),
			Branch:       run.GetHeadBranch(),
			Event:        run.GetEvent(),
			Actor:        run.GetActor().GetLogin(),
			Repository:   repo.GetName(),
		})
	}

	return recentJobs, nil
}