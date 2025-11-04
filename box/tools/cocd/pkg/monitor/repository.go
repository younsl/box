package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/google/go-github/v60/github"
	ghclient "github.com/younsl/cocd/pkg/github"
)

const (
	DefaultRepoCacheExpiry = 60 * time.Minute
	DefaultPerPage         = 30
	
	DefaultMaxAge = 7 * 24 * time.Hour
	
	BytesToMB = 1024 * 1024
)

type RepositoryManager struct {
	client          *ghclient.Client
	cachedRepos     []*github.Repository
	lastRepoFetch   time.Time
	repoCacheExpiry time.Duration
}

func NewRepositoryManager(client *ghclient.Client) *RepositoryManager {
	return &RepositoryManager{
		client:          client,
		repoCacheExpiry: DefaultRepoCacheExpiry,
	}
}

func (rm *RepositoryManager) GetRepositoriesWithCache(ctx context.Context) ([]*github.Repository, error) {
	if len(rm.cachedRepos) > 0 && time.Since(rm.lastRepoFetch) < rm.repoCacheExpiry {
		return rm.cachedRepos, nil
	}

	var allRepos []*github.Repository
	page := 1
	for {
		repoOpts := &github.RepositoryListByOrgOptions{
			Type:      "sources",
			Sort:      "pushed",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		}

		repos, resp, err := rm.client.ListRepositories(ctx, repoOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories - check your token and organization name: %w", err)
		}

		for _, repo := range repos {
			if !repo.GetArchived() && !repo.GetDisabled() {
				allRepos = append(allRepos, repo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	rm.cachedRepos = allRepos
	rm.lastRepoFetch = time.Now()

	return allRepos, nil
}

func (rm *RepositoryManager) FilterRepositories(repos []*github.Repository, filter RepoFilter) []*github.Repository {
	var filtered []*github.Repository
	
	for _, repo := range repos {
		if repo.GetArchived() && !filter.IncludeArchived {
			continue
		}
		
		if repo.GetDisabled() && !filter.IncludeDisabled {
			continue
		}
		
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

func (rm *RepositoryManager) GetActiveRepositories(ctx context.Context, maxRepos int) ([]*github.Repository, error) {
	allRepos, err := rm.GetRepositoriesWithCache(ctx)
	if err != nil {
		return nil, err
	}

	if maxRepos > MaxActiveRepositories {
		maxRepos = MaxActiveRepositories
	}

	filter := RepoFilter{
		IncludeArchived: false,
		IncludeDisabled: false,
		MaxAge:          DefaultMaxAge,
	}
	
	activeRepos := rm.FilterRepositories(allRepos, filter)
	
	if len(activeRepos) > 0 {
		sort.Slice(activeRepos, func(i, j int) bool {
			if activeRepos[i].PushedAt == nil || activeRepos[j].PushedAt == nil {
				return false
			}
			return activeRepos[i].PushedAt.Time.After(activeRepos[j].PushedAt.Time)
		})
	} else {
		filter.MaxAge = 0
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

	if len(activeRepos) > maxRepos {
		activeRepos = activeRepos[:maxRepos]
	}

	return activeRepos, nil
}

func (rm *RepositoryManager) GetSmartRepositories(ctx context.Context, maxRepos int) ([]*github.Repository, error) {
	allRepos, err := rm.GetRepositoriesWithCache(ctx)
	if err != nil {
		return nil, err
	}

	var candidateRepos []*github.Repository
	
	for _, repo := range allRepos {
		if repo.PushedAt == nil || time.Since(repo.PushedAt.Time) > DefaultMaxAge {
			continue
		}
		
		candidateRepos = append(candidateRepos, repo)
	}
	
	sort.Slice(candidateRepos, func(i, j int) bool {
		if candidateRepos[i].PushedAt == nil || candidateRepos[j].PushedAt == nil {
			return false
		}
		return candidateRepos[i].PushedAt.Time.After(candidateRepos[j].PushedAt.Time)
	})

	if len(candidateRepos) > maxRepos {
		candidateRepos = candidateRepos[:maxRepos]
	}

	return candidateRepos, nil
}

func (rm *RepositoryManager) hasWorkflowFiles(ctx context.Context, repo *github.Repository) bool {
	opts := &github.RepositoryContentGetOptions{}
	_, _, _, err := rm.client.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), ".github/workflows", opts)
	return err == nil
}

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

func (rm *RepositoryManager) GetMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	allocMB := m.Alloc / BytesToMB
	sysMB := m.Sys / BytesToMB
	
	return fmt.Sprintf("%dMB/%dMB", allocMB, sysMB)
}
