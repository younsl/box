package tui

import (
	"context"
	
	"github.com/google/go-github/v60/github"
	githubclient "github.com/younsl/cocd/pkg/github"
)

// GitHubClient defines the interface for GitHub operations
type GitHubClient interface {
	CancelWorkflowRun(ctx context.Context, repo string, runID int64) (*github.Response, error)
	GetPendingDeployments(ctx context.Context, repo string, runID int64) ([]*githubclient.PendingDeployment, *github.Response, error)
	ApprovePendingDeployment(ctx context.Context, repo string, runID int64, environmentIDs []int64, comment string) (*github.Response, error)
	GetWorkflowRun(ctx context.Context, repo string, runID int64) (*github.WorkflowRun, *github.Response, error)
}

// GitHubClientAdapter adapts the internal GitHub client to our interface
type GitHubClientAdapter struct {
	client *githubclient.Client
}

// NewGitHubClientAdapter creates a new GitHub client adapter
func NewGitHubClientAdapter(client interface{}) GitHubClient {
	if gc, ok := client.(*githubclient.Client); ok {
		return &GitHubClientAdapter{client: gc}
	}
	return nil
}

func (gca *GitHubClientAdapter) CancelWorkflowRun(ctx context.Context, repo string, runID int64) (*github.Response, error) {
	return gca.client.CancelWorkflowRun(ctx, repo, runID)
}

func (gca *GitHubClientAdapter) GetPendingDeployments(ctx context.Context, repo string, runID int64) ([]*githubclient.PendingDeployment, *github.Response, error) {
	return gca.client.GetPendingDeployments(ctx, repo, runID)
}

func (gca *GitHubClientAdapter) ApprovePendingDeployment(ctx context.Context, repo string, runID int64, environmentIDs []int64, comment string) (*github.Response, error) {
	return gca.client.ApprovePendingDeployment(ctx, repo, runID, environmentIDs, comment)
}

func (gca *GitHubClientAdapter) GetWorkflowRun(ctx context.Context, repo string, runID int64) (*github.WorkflowRun, *github.Response, error) {
	return gca.client.GetWorkflowRun(ctx, repo, runID)
}