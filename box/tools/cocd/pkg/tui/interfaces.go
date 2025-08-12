package tui

import (
	"context"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/scanner"
)

// Monitor defines the interface for monitoring jobs
type Monitor interface {
	StartMonitoring(ctx context.Context, jobsChan chan []scanner.JobStatus)
	GetPendingJobs(ctx context.Context) ([]scanner.JobStatus, error)
	GetRecentJobs(ctx context.Context) ([]scanner.JobStatus, error)
	GetClient() interface{} // Returns GitHub client
	GetProgressTracker() ProgressTracker
	GetScanProgress() monitor.ScanProgress
	GetUpdateInterval() int
	GetRecentJobsWithStreaming(ctx context.Context, jobUpdateChan chan<- monitor.JobUpdate) error
	GetAuthenticatedUser(ctx context.Context) (string, error)
}

// ProgressTracker defines the interface for tracking progress
type ProgressTracker interface {
	UpdateScanCountdown()
	SetNextScanTimer(nextScanAt time.Time, scanCount int, isFull bool)
}

// ViewManager defines the interface for managing views
type ViewManagerInterface interface {
	// View management
	SwitchToView(viewType ViewType)
	GetCurrentView() ViewType
	
	// Cursor management
	GetCursor() int
	MoveCursor(direction int, maxItems int)
	
	// Pagination
	GetPageInfo() (page int, perPage int)
	ChangePage(direction int, totalItems int)
	GetPaginatedJobs(jobs []scanner.JobStatus) []scanner.JobStatus
	
	// Job tracking
	TrackCompletedJobs(currentJobs, newJobs []scanner.JobStatus)
	GetCombinedPendingJobs(jobs []scanner.JobStatus) []scanner.JobStatus
	IsJobCompleted(job scanner.JobStatus) bool
	GetMaxCursorPosition(pendingJobs, recentJobs []scanner.JobStatus) int
	
	// Cancel confirmation
	ShowCancelConfirm(job scanner.JobStatus)
	HideCancelConfirm()
	IsShowingCancelConfirm() bool
	GetCancelTargetJob() *scanner.JobStatus
	SetCancelSelection(selection int)
	GetCancelSelection() int
	IsCancelConfirmed() bool
	
	// Approval confirmation
	ShowApprovalConfirm(job scanner.JobStatus)
	HideApprovalConfirm()
	IsShowingApprovalConfirm() bool
	GetApprovalTargetJob() *scanner.JobStatus
	SetApprovalSelection(selection int)
	GetApprovalSelection() int
	IsApprovalConfirmed() bool
	
	// Job highlighting for newly scanned jobs
	MarkNewlyScannedJobs(jobs []scanner.JobStatus) []scanner.JobStatus
	IsJobHighlighted(job scanner.JobStatus) bool
}

// CommandHandler defines the interface for handling commands
type CommandHandlerInterface interface {
	StartMonitoring(ctx context.Context, jobsChan chan []scanner.JobStatus) tea.Cmd
	LoadPendingJobs(ctx context.Context) tea.Cmd
	LoadRecentJobs(ctx context.Context) tea.Cmd
	LoadRecentJobsStreaming(ctx context.Context, updateChan chan<- tea.Msg) tea.Cmd
	TickCmd() tea.Cmd
	JumpToActions(vm ViewManagerInterface, jobs, recentJobs []scanner.JobStatus) tea.Cmd
	InitializeTimer()
	UpdateTimerForView(viewType ViewType)
	CancelWorkflow(ctx context.Context, vm ViewManagerInterface) tea.Cmd
	ApproveDeployment(ctx context.Context, vm ViewManagerInterface) tea.Cmd
	DelayedRefresh(delay time.Duration) tea.Cmd
}

// UIRenderer defines the interface for rendering UI components
type UIRenderer interface {
	RenderHeader(monitor Monitor) string
	RenderViewSelector(currentView ViewType, pendingCount, recentCount int, vm ViewManagerInterface) string
	RenderJobTable(jobs []scanner.JobStatus, cursor int, vm ViewManagerInterface) string
	RenderStatus(errorMsg string) string
	RenderPagination(currentView ViewType, vm ViewManagerInterface, totalJobs int, jobs []scanner.JobStatus) string
	RenderHelp(monitor Monitor) string
	RenderCancelConfirm(job scanner.JobStatus, selection int) string
	RenderApprovalConfirm(job scanner.JobStatus, selection int) string
}

// KeyHandler defines the interface for handling keyboard input
type KeyHandler interface {
	HandleKeyPress(msg tea.KeyMsg, app *BubbleApp) (tea.Model, tea.Cmd)
}

// JobService defines the interface for job-related operations
type JobService interface {
	GetJobsForView(view ViewType, pendingJobs, recentJobs []scanner.JobStatus, vm ViewManagerInterface) []scanner.JobStatus
	RefreshJobs(ctx context.Context, view ViewType) tea.Cmd
	RefreshJobsWithStreaming(ctx context.Context, view ViewType, updateChan chan<- tea.Msg) tea.Cmd
}