package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/internal/monitor"
	"github.com/younsl/cocd/internal/scanner"
)

// CommandHandler handles all command operations
type CommandHandler struct {
	monitor *monitor.Monitor
	config  *AppConfig
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(monitor *monitor.Monitor, config *AppConfig) *CommandHandler {
	return &CommandHandler{
		monitor: monitor,
		config:  config,
	}
}

// StartMonitoring starts the background monitoring process
func (ch *CommandHandler) StartMonitoring(ctx context.Context, jobsChan chan []scanner.JobStatus) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Start the monitoring in the background
		go ch.monitor.StartMonitoring(ctx, jobsChan)
		
		// Start a goroutine to listen for job updates and forward them to the UI
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case jobs := <-jobsChan:
					// We can't directly send tea.Msg from here
					// The monitor will handle periodic updates
					_ = jobs
				}
			}
		}()
		
		// Return immediate command to load pending jobs
		return ch.LoadPendingJobs(ctx)()
	})
}

// LoadPendingJobs loads pending jobs
func (ch *CommandHandler) LoadPendingJobs(ctx context.Context) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		jobs, err := ch.monitor.GetPendingJobs(ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return jobsMsg(jobs)
	})
}

// LoadRecentJobs loads recent jobs
func (ch *CommandHandler) LoadRecentJobs(ctx context.Context) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		jobs, err := ch.monitor.GetRecentJobs(ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		// Set timer after loading recent jobs since this is a one-time operation
		nextScanAt := time.Now().Add(30 * time.Second)
		ch.monitor.GetProgressTracker().SetNextScanTimer(nextScanAt, 1, false)
		return recentJobsMsg(jobs)
	})
}

// TickCmd creates a tick command for periodic updates
func (ch *CommandHandler) TickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// JumpToActions opens the selected job's GitHub Actions page in browser
func (ch *CommandHandler) JumpToActions(vm *ViewManager, jobs, recentJobs []scanner.JobStatus) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		var selectedJob *scanner.JobStatus
		
		if vm.GetCurrentView() == ViewPending {
			// Get the combined jobs list (same as rendered in table)
			combinedJobs := vm.GetCombinedPendingJobs(jobs)
			if len(combinedJobs) > 0 && vm.GetCursor() < len(combinedJobs) {
				selectedJob = &combinedJobs[vm.GetCursor()]
			}
		} else if vm.GetCurrentView() == ViewRecent {
			visibleJobs := vm.GetPaginatedJobs(recentJobs)
			if len(visibleJobs) > 0 && vm.GetCursor() < len(visibleJobs) {
				selectedJob = &visibleJobs[vm.GetCursor()]
			}
		}
		
		if selectedJob != nil {
			url := selectedJob.GetActionsURL(ch.config.ServerURL, ch.config.Org)
			if err := OpenURL(url); err != nil {
				return errorMsg(fmt.Sprintf("Failed to open browser: %v", err))
			}
		}
		
		return nil
	})
}

// InitializeTimer initializes the timer to prevent loading state
func (ch *CommandHandler) InitializeTimer() {
	nextScanAt := time.Now().Add(10 * time.Second)
	ch.monitor.GetProgressTracker().SetNextScanTimer(nextScanAt, 1, false)
}

// UpdateTimerForView updates the timer for a specific view
func (ch *CommandHandler) UpdateTimerForView(viewType ViewType) {
	var delay time.Duration
	switch viewType {
	case ViewRecent:
		delay = 30 * time.Second
	default:
		delay = 10 * time.Second
	}
	
	nextScanAt := time.Now().Add(delay)
	ch.monitor.GetProgressTracker().SetNextScanTimer(nextScanAt, 1, false)
}

// CancelWorkflow cancels the selected workflow run
func (ch *CommandHandler) CancelWorkflow(ctx context.Context, vm *ViewManager) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		job := vm.GetCancelTargetJob()
		if job == nil {
			return errorMsg("No job selected for cancellation")
		}
		
		// Get the GitHub client from monitor
		client := ch.monitor.GetClient()
		if client == nil {
			return errorMsg("GitHub client not available")
		}
		
		// Cancel the workflow run
		_, err := client.CancelWorkflowRun(ctx, job.Repository, job.RunID)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to cancel workflow: %v", err))
		}
		
		return cancelSuccessMsg{}
	})
}