package tui

import (
	"context"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/pkg/scanner"
)

// DefaultJobService implements the JobService interface
type DefaultJobService struct {
	commands CommandHandlerInterface
}

// NewJobService creates a new job service
func NewJobService(commands CommandHandlerInterface) JobService {
	return &DefaultJobService{
		commands: commands,
	}
}

// GetJobsForView returns jobs for the current view
func (js *DefaultJobService) GetJobsForView(view ViewType, pendingJobs, recentJobs []scanner.JobStatus, vm ViewManagerInterface) []scanner.JobStatus {
	switch view {
	case ViewPending:
		highlightedPendingJobs := vm.MarkNewlyScannedJobs(pendingJobs)
		return vm.GetCombinedPendingJobs(highlightedPendingJobs)
	case ViewRecent:
		highlightedRecentJobs := vm.MarkNewlyScannedJobs(recentJobs)
		return vm.GetPaginatedJobs(highlightedRecentJobs)
	default:
		return []scanner.JobStatus{}
	}
}

// RefreshJobs refreshes jobs for the current view
func (js *DefaultJobService) RefreshJobs(ctx context.Context, view ViewType) tea.Cmd {
	if view == ViewPending {
		return js.commands.LoadPendingJobs(ctx)
	}
	return js.commands.LoadRecentJobs(ctx)
}

// RefreshJobsWithStreaming refreshes jobs with streaming for recent jobs
func (js *DefaultJobService) RefreshJobsWithStreaming(ctx context.Context, view ViewType, updateChan chan<- tea.Msg) tea.Cmd {
	if view == ViewPending {
		return js.commands.LoadPendingJobs(ctx)
	}
	return js.commands.LoadRecentJobsStreaming(ctx, updateChan)
}