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

