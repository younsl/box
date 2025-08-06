package github

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	org    string
	repo   string
}

func NewClient(token, baseURL, org string, repo ...string) (*Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	if baseURL != "" && baseURL != "https://api.github.com" {
		// Ensure trailing slash for GitHub API
		if baseURL[len(baseURL)-1] != '/' {
			baseURL += "/"
		}
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		client.BaseURL = parsedURL
	}

	var repoName string
	if len(repo) > 0 {
		repoName = repo[0]
	}

	return &Client{
		client: client,
		org:    org,
		repo:   repoName,
	}, nil
}

func (c *Client) ListRepositories(ctx context.Context, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
	return c.client.Repositories.ListByOrg(ctx, c.org, opts)
}

func (c *Client) ListWorkflowRuns(ctx context.Context, repo string, opts *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	return c.client.Actions.ListRepositoryWorkflowRuns(ctx, c.org, repo, opts)
}

func (c *Client) ListWorkflowJobs(ctx context.Context, repo string, runID int64, opts *github.ListWorkflowJobsOptions) (*github.Jobs, *github.Response, error) {
	return c.client.Actions.ListWorkflowJobs(ctx, c.org, repo, runID, opts)
}

func (c *Client) GetWorkflowRun(ctx context.Context, repo string, runID int64) (*github.WorkflowRun, *github.Response, error) {
	return c.client.Actions.GetWorkflowRunByID(ctx, c.org, repo, runID)
}

func (c *Client) ListEnvironments(ctx context.Context, repo string) (*github.EnvResponse, *github.Response, error) {
	return c.client.Repositories.ListEnvironments(ctx, c.org, repo, &github.EnvironmentListOptions{})
}

func (c *Client) CancelWorkflowRun(ctx context.Context, repo string, runID int64) (*github.Response, error) {
	return c.client.Actions.CancelWorkflowRunByID(ctx, c.org, repo, runID)
}

// ApprovePendingDeployment approves a pending deployment for a workflow run
func (c *Client) ApprovePendingDeployment(ctx context.Context, repo string, runID int64, environmentIDs []int64, comment string) (*github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/actions/runs/%v/pending_deployments", c.org, repo, runID)
	
	type approvalRequest struct {
		EnvironmentIDs []int64 `json:"environment_ids"`
		State          string  `json:"state"`
		Comment        string  `json:"comment"`
	}
	
	req := &approvalRequest{
		EnvironmentIDs: environmentIDs,
		State:          "approved",
		Comment:        comment,
	}
	
	request, err := c.client.NewRequest("POST", u, req)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.client.Do(ctx, request, nil)
	if err != nil {
		return resp, err
	}
	
	return resp, nil
}

// PendingDeployment represents a pending deployment
type PendingDeployment struct {
	Environment struct {
		ID   *int64  `json:"id,omitempty"`
		Name *string `json:"name,omitempty"`
	} `json:"environment"`
	WaitTimer            int    `json:"wait_timer"`
	WaitTimerStartedAt   string `json:"wait_timer_started_at"`
	CurrentUserCanApprove bool   `json:"current_user_can_approve"`
	Reviewers            []struct {
		Type string `json:"type"`
		ID   int64  `json:"id"`
	} `json:"reviewers"`
}

// GetPendingDeployments gets pending deployments for a workflow run
func (c *Client) GetPendingDeployments(ctx context.Context, repo string, runID int64) ([]*PendingDeployment, *github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/actions/runs/%v/pending_deployments", c.org, repo, runID)
	
	request, err := c.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	
	var pendingDeployments []*PendingDeployment
	resp, err := c.client.Do(ctx, request, &pendingDeployments)
	if err != nil {
		return nil, resp, err
	}
	
	return pendingDeployments, resp, nil
}

// ListDeployments lists deployments for a repository
func (c *Client) ListDeployments(ctx context.Context, repo string, opts *github.DeploymentsListOptions) ([]*github.Deployment, *github.Response, error) {
	return c.client.Repositories.ListDeployments(ctx, c.org, repo, opts)
}

// GetWorkflowJob gets a specific workflow job
func (c *Client) GetWorkflowJob(ctx context.Context, repo string, jobID int64) (*github.WorkflowJob, *github.Response, error) {
	return c.client.Actions.GetWorkflowJobByID(ctx, c.org, repo, jobID)
}

// GetContents gets the contents of a file or directory
func (c *Client) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	return c.client.Repositories.GetContents(ctx, owner, repo, path, opts)
}

