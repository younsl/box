package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/scanner"
)

// CommandHandler handles all command operations
type CommandHandler struct {
	monitor Monitor
	config  *AppConfig
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(monitor Monitor, config *AppConfig) CommandHandlerInterface {
	return &CommandHandler{
		monitor: monitor,
		config:  config,
	}
}

func (ch *CommandHandler) generateApprovalMessage() string {
	timezone := ch.config.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	
	timestamp := time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
	return fmt.Sprintf("Remote approved by cocd at %s", timestamp)
}

func (ch *CommandHandler) StartMonitoring(ctx context.Context, jobsChan chan []scanner.JobStatus) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		go ch.monitor.StartMonitoring(ctx, jobsChan)
		
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case jobs := <-jobsChan:
					_ = jobs
				}
			}
		}()
		
		return ch.LoadPendingJobs(ctx)()
	})
}

func (ch *CommandHandler) LoadPendingJobs(ctx context.Context) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		jobs, err := ch.monitor.GetPendingJobs(ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return pendingJobsMsg(jobs)
	})
}


func (ch *CommandHandler) LoadRecentJobs(ctx context.Context) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		jobs, err := ch.monitor.GetRecentJobs(ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		nextScanAt := time.Now().Add(30 * time.Second)
		ch.monitor.GetProgressTracker().SetNextScanTimer(nextScanAt, 1, false)
		return recentJobsMsg(jobs)
	})
}

func (ch *CommandHandler) LoadRecentJobsStreaming(ctx context.Context, updateChan chan<- tea.Msg) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		jobUpdateChan := make(chan monitor.JobUpdate, 100)
		
		go func() {
			defer close(jobUpdateChan)
			
			err := ch.monitor.GetRecentJobsWithStreaming(ctx, jobUpdateChan)
			if err != nil {
				select {
				case jobUpdateChan <- monitor.JobUpdate{Error: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
		
		go func() {
			for update := range jobUpdateChan {
				select {
				case updateChan <- recentJobUpdateMsg(update):
				case <-ctx.Done():
					return
				}
			}
		}()
		
		return scanProgressMsg{}
	})
}

func (ch *CommandHandler) TickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (ch *CommandHandler) JumpToActions(vm ViewManagerInterface, jobs, recentJobs []scanner.JobStatus) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		var selectedJob *scanner.JobStatus
		
		if vm.GetCurrentView() == ViewPending {
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

func (ch *CommandHandler) InitializeTimer() {
	nextScanAt := time.Now().Add(10 * time.Second)
	ch.monitor.GetProgressTracker().SetNextScanTimer(nextScanAt, 1, false)
}

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

func (ch *CommandHandler) DelayedRefresh(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return delayedRefreshMsg{}
	})
}

func (ch *CommandHandler) CancelWorkflow(ctx context.Context, vm ViewManagerInterface) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		job := vm.GetCancelTargetJob()
		if job == nil {
			return errorMsg("No job selected for cancellation")
		}
		
		clientInterface := ch.monitor.GetClient()
		if clientInterface == nil {
			return errorMsg("GitHub client not available")
		}
		
		client := NewGitHubClientAdapter(clientInterface)
		if client == nil {
			return errorMsg("Failed to create GitHub client adapter")
		}
		
		_, err := client.CancelWorkflowRun(ctx, job.Repository, job.RunID)
		if err != nil {
			// Silently handle "job scheduled" error - it usually means cancellation is processing
			if strings.Contains(err.Error(), "job scheduled on GitHub side") {
				// Return processing message to trigger delayed refresh
				return cancelProcessingMsg{job: job}
			}
			return errorMsg(fmt.Sprintf("Failed to cancel workflow: %v", err))
		}
		
		return cancelSuccessMsg{}
	})
}

func (ch *CommandHandler) ApproveDeployment(ctx context.Context, vm ViewManagerInterface) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		job := vm.GetApprovalTargetJob()
		if job == nil {
			return errorMsg("No job selected for approval")
		}
		
		clientInterface := ch.monitor.GetClient()
		if clientInterface == nil {
			return errorMsg("GitHub client not available")
		}
		
		client := NewGitHubClientAdapter(clientInterface)
		if client == nil {
			return errorMsg("Failed to create GitHub client adapter")
		}
		
		pendingDeployments, _, err := client.GetPendingDeployments(ctx, job.Repository, job.RunID)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to get pending deployments: %v", err))
		}
		
		if len(pendingDeployments) == 0 {
			return errorMsg("No pending deployments found for this workflow")
		}
		
		var environmentIDs []int64
		for _, pd := range pendingDeployments {
			if pd.Environment.ID != nil {
				environmentIDs = append(environmentIDs, *pd.Environment.ID)
			}
		}
		
		if len(environmentIDs) == 0 {
			return errorMsg("No environment IDs found in pending deployments")
		}
		
		_, err = client.ApprovePendingDeployment(ctx, job.Repository, job.RunID, environmentIDs, ch.generateApprovalMessage())
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to approve deployment: %v", err))
		}
		
		return approvalSuccessMsg{}
	})
}