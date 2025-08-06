package scanner

import (
	"context"

	"github.com/google/go-github/v60/github"
)

// Scanner interface for different scanning strategies
type Scanner interface {
	ScanRepository(ctx context.Context, repo *github.Repository) ([]JobStatus, error)
}