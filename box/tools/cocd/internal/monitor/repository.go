package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/internal/github"
)

// RepositoryManager handles repository caching and filtering
type RepositoryManager struct {
	client          *ghclient.Client
	cachedRepos     []*github.Repository
	lastRepoFetch   time.Time
	repoCacheExpiry time.Duration
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(client *ghclient.Client) *RepositoryManager {
	return &RepositoryManager{
		client:          client,
		repoCacheExpiry: 60 * time.Minute, // Extended cache for GHES performance
	}
}

// GetRepositoriesWithCache returns repositories using cache when possible
func (rm *RepositoryManager) GetRepositoriesWithCache(ctx context.Context) ([]*github.Repository, error) {
	// Check if cache is still valid
	if len(rm.cachedRepos) > 0 && time.Since(rm.lastRepoFetch) < rm.repoCacheExpiry {
		return rm.cachedRepos, nil
	}

	// Cache expired or empty, fetch fresh data
	var allRepos []*github.Repository
	page := 1
	for {
		repoOpts := &github.RepositoryListByOrgOptions{
			Type: "all",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 30, // Further reduced for GHES load
			},
		}

		repos, resp, err := rm.client.ListRepositories(ctx, repoOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories - check your token and organization name: %w", err)
		}

		allRepos = append(allRepos, repos...)

		// Check if we've reached the last page
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Update cache
	rm.cachedRepos = allRepos
	rm.lastRepoFetch = time.Now()

	return allRepos, nil
}

// FilterRepositories filters repositories based on criteria
func (rm *RepositoryManager) FilterRepositories(repos []*github.Repository, filter RepoFilter) []*github.Repository {
	var filtered []*github.Repository
	
	for _, repo := range repos {
		// Skip archived repos unless explicitly included
		if repo.GetArchived() && !filter.IncludeArchived {
			continue
		}
		
		// Skip disabled repos unless explicitly included
		if repo.GetDisabled() && !filter.IncludeDisabled {
			continue
		}
		
		// For fast scanning with MaxAge, check recent activity
		if filter.MaxAge > 0 {
			if repo.PushedAt != nil && time.Since(repo.PushedAt.Time) < filter.MaxAge {
				filtered = append(filtered, repo)
			}
		} else {
			filtered = append(filtered, repo)
		}
	}
	
	return filtered
}

// GetActiveRepositories returns repositories with recent activity for fast scanning
func (rm *RepositoryManager) GetActiveRepositories(ctx context.Context, maxRepos int) ([]*github.Repository, error) {
	allRepos, err := rm.GetRepositoriesWithCache(ctx)
	if err != nil {
		return nil, err
	}

	// Always limit to 100 repositories maximum for balanced performance
	if maxRepos > 100 {
		maxRepos = 100
	}

	// Filter for active repositories (recent pushes within 7 days for better targeting)
	filter := RepoFilter{
		IncludeArchived: false,
		IncludeDisabled: false,
		MaxAge:          7 * 24 * time.Hour,
	}
	
	activeRepos := rm.FilterRepositories(allRepos, filter)
	
	// If we have active repos, sort by most recent push activity
	if len(activeRepos) > 0 {
		sort.Slice(activeRepos, func(i, j int) bool {
			if activeRepos[i].PushedAt == nil || activeRepos[j].PushedAt == nil {
				return false
			}
			return activeRepos[i].PushedAt.Time.After(activeRepos[j].PushedAt.Time)
		})
	} else {
		// Fallback to recently updated repositories
		filter.MaxAge = 0 // Remove age filter
		filter.IncludeArchived = false
		filter.IncludeDisabled = false
		
		activeRepos = rm.FilterRepositories(allRepos, filter)
		sort.Slice(activeRepos, func(i, j int) bool {
			if activeRepos[i].UpdatedAt == nil || activeRepos[j].UpdatedAt == nil {
				return false
			}
			return activeRepos[i].UpdatedAt.Time.After(activeRepos[j].UpdatedAt.Time)
		})
	}

	// Limit to specified number of repositories
	if len(activeRepos) > maxRepos {
		activeRepos = activeRepos[:maxRepos]
	}

	return activeRepos, nil
}

// GetSmartRepositories returns top 100 repositories with GitHub Actions and recent activity
func (rm *RepositoryManager) GetSmartRepositories(ctx context.Context, maxRepos int) ([]*github.Repository, error) {
	allRepos, err := rm.GetRepositoriesWithCache(ctx)
	if err != nil {
		return nil, err
	}

	var candidateRepos []*github.Repository
	
	for _, repo := range allRepos {
		// Skip archived and disabled repos
		if repo.GetArchived() || repo.GetDisabled() {
			continue
		}
		
		// Must have recent activity (last 7 days for focused scanning)
		if repo.PushedAt == nil || time.Since(repo.PushedAt.Time) > 7*24*time.Hour {
			continue
		}

		// Skip repos without workflow files (heuristic check)
		if !rm.hasWorkflowFiles(ctx, repo) {
			continue
		}
		
		candidateRepos = append(candidateRepos, repo)
	}
	
	// Sort by most recent activity
	sort.Slice(candidateRepos, func(i, j int) bool {
		if candidateRepos[i].PushedAt == nil || candidateRepos[j].PushedAt == nil {
			return false
		}
		return candidateRepos[i].PushedAt.Time.After(candidateRepos[j].PushedAt.Time)
	})

	// Return top 100 active repos with Actions
	if len(candidateRepos) > maxRepos {
		candidateRepos = candidateRepos[:maxRepos]
	}

	return candidateRepos, nil
}

// hasWorkflowFiles checks if repository has workflow files (quick heuristic)
func (rm *RepositoryManager) hasWorkflowFiles(ctx context.Context, repo *github.Repository) bool {
	// Quick check: .github/workflows directory exists
	opts := &github.RepositoryContentGetOptions{}
	_, _, _, err := rm.client.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), ".github/workflows", opts)
	return err == nil // If no error, directory exists
}

// GetValidRepositories returns all non-archived, non-disabled repositories
func (rm *RepositoryManager) GetValidRepositories(ctx context.Context) ([]*github.Repository, error) {
	allRepos, err := rm.GetRepositoriesWithCache(ctx)
	if err != nil {
		return nil, err
	}

	filter := RepoFilter{
		IncludeArchived: false,
		IncludeDisabled: false,
	}
	
	return rm.FilterRepositories(allRepos, filter), nil
}

// CalculateRepoStats calculates repository statistics
func (rm *RepositoryManager) CalculateRepoStats(repos []*github.Repository) (archived, disabled, valid int) {
	for _, repo := range repos {
		if repo.GetArchived() {
			archived++
		} else if repo.GetDisabled() {
			disabled++
		} else {
			valid++
		}
	}
	return
}

// GetCacheStatus returns cache status information
func (rm *RepositoryManager) GetCacheStatus() string {
	if len(rm.cachedRepos) == 0 {
		return "Empty"
	}
	
	timeSince := time.Since(rm.lastRepoFetch)
	remaining := rm.repoCacheExpiry - timeSince
	
	if remaining <= 0 {
		return "Expired"
	}
	
	if remaining > time.Minute {
		return fmt.Sprintf("ttl %dm", int(remaining.Minutes()))
	}
	return fmt.Sprintf("ttl %ds", int(remaining.Seconds()))
}

// GetMemoryUsage returns current memory usage information
func (rm *RepositoryManager) GetMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Convert bytes to MB
	allocMB := m.Alloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024
	
	return fmt.Sprintf("%dMB/%dMB", allocMB, sysMB)
}