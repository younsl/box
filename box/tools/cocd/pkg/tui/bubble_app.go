package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/internal/scanner"
)

// BubbleApp is the main Bubble Tea application model
type BubbleApp struct {
	// Core dependencies
	monitor Monitor
	config  *AppConfig
	ctx     context.Context
	cancel  context.CancelFunc
	
	// Component managers
	viewManager    ViewManagerInterface
	uiRenderer     UIRenderer
	commandHandler CommandHandlerInterface
	keyHandler     KeyHandler
	jobService     JobService
	
	// Application state
	jobs       []scanner.JobStatus
	recentJobs []scanner.JobStatus
	
	// UI state
	showHelp     bool
	loading      bool
	errorMsg     string
	lastUpdate   time.Time
	lastCountdown int // Track last countdown value for change detection
	width        int // Terminal width for pagination alignment
	height       int // Terminal height
	
	// Channels
	jobsChan chan []scanner.JobStatus
}

// NewBubbleApp creates a new Bubble Tea application
func NewBubbleApp(m Monitor, config *AppConfig) *BubbleApp {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize component managers
	viewManager := NewViewManager()
	uiRenderer := NewUIComponents(config)
	commandHandler := NewCommandHandler(m, config)
	keyHandler := NewKeyHandler(commandHandler)
	jobService := NewJobService(commandHandler)
	
	app := &BubbleApp{
		monitor:        m,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		viewManager:    viewManager,
		uiRenderer:     uiRenderer,
		commandHandler: commandHandler,
		keyHandler:     keyHandler,
		jobService:     jobService,
		jobsChan:       make(chan []scanner.JobStatus, 100),
		loading:        true,
	}
	
	// Initialize timer immediately to prevent "Loading" state
	commandHandler.InitializeTimer()
	
	return app
}

// Init initializes the Bubble Tea application
func (app *BubbleApp) Init() tea.Cmd {
	return tea.Batch(
		app.commandHandler.StartMonitoring(app.ctx, app.jobsChan),
		app.commandHandler.TickCmd(),
	)
}

// Update handles messages and updates the model
func (app *BubbleApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		return app, nil
		
	case tea.KeyMsg:
		return app.keyHandler.HandleKeyPress(msg, app)
		
	case jobsMsg:
		return app.handleJobsMessage(msg)
		
	case recentJobsMsg:
		return app.handleRecentJobsMessage(msg)
		
	case errorMsg:
		return app.handleErrorMessage(msg)
		
	case tickMsg:
		return app.handleTickMessage(msg)
		
	case scanProgressMsg:
		return app, nil
		
	case updateUIMsg:
		// Force UI update without changing data
		return app, nil
		
	case cancelSuccessMsg:
		app.viewManager.HideCancelConfirm()
		// Refresh the current view to see updated status
		return app.refreshCurrentView()
		
	case approvalSuccessMsg:
		app.viewManager.HideApprovalConfirm()
		// Refresh the current view to see updated status
		return app.refreshCurrentView()
		
	default:
		return app, nil
	}
}

// View renders the UI
func (app *BubbleApp) View() string {
	if app.showHelp {
		return app.uiRenderer.RenderHelp(app.monitor)
	}
	
	if app.viewManager.IsShowingCancelConfirm() {
		if job := app.viewManager.GetCancelTargetJob(); job != nil {
			selection := app.viewManager.GetCancelSelection()
			return app.uiRenderer.RenderCancelConfirm(*job, selection)
		}
	}
	
	if app.viewManager.IsShowingApprovalConfirm() {
		if job := app.viewManager.GetApprovalTargetJob(); job != nil {
			selection := app.viewManager.GetApprovalSelection()
			return app.uiRenderer.RenderApprovalConfirm(*job, selection)
		}
	}
	
	return app.renderMain()
}

// Message handlers

func (app *BubbleApp) handleJobsMessage(msg jobsMsg) (tea.Model, tea.Cmd) {
	newJobs := []scanner.JobStatus(msg)
	
	// Track completed jobs
	app.viewManager.TrackCompletedJobs(app.jobs, newJobs)
	
	app.jobs = newJobs
	app.loading = false
	app.lastUpdate = time.Now()
	app.errorMsg = ""
	
	return app, nil
}

func (app *BubbleApp) handleRecentJobsMessage(msg recentJobsMsg) (tea.Model, tea.Cmd) {
	app.recentJobs = []scanner.JobStatus(msg)
	app.loading = false
	app.lastUpdate = time.Now()
	app.errorMsg = ""
	
	// Ensure timer is set for recent jobs view since it's a one-time load
	app.commandHandler.UpdateTimerForView(ViewRecent)
	
	return app, nil
}

func (app *BubbleApp) handleErrorMessage(msg errorMsg) (tea.Model, tea.Cmd) {
	app.errorMsg = string(msg)
	app.loading = false
	
	// Hide cancel/approval confirmation popup if it's showing (in case of workflow error)
	if app.viewManager.IsShowingCancelConfirm() {
		app.viewManager.HideCancelConfirm()
	}
	if app.viewManager.IsShowingApprovalConfirm() {
		app.viewManager.HideApprovalConfirm()
	}
	
	return app, nil
}

func (app *BubbleApp) handleTickMessage(msg tickMsg) (tea.Model, tea.Cmd) {
	// Update the view for real-time AGE updates and scan countdown
	app.monitor.GetProgressTracker().UpdateScanCountdown()
	
	// Auto-refresh pending jobs every 10 seconds
	if app.viewManager.GetCurrentView() == ViewPending && time.Since(app.lastUpdate) > 10*time.Second {
		return app, tea.Batch(app.commandHandler.TickCmd(), app.commandHandler.LoadPendingJobs(app.ctx))
	}
	
	// Check if countdown value has changed for efficient UI updates
	currentCountdown := app.monitor.GetScanProgress().ScanCountdown
	if currentCountdown != app.lastCountdown {
		app.lastCountdown = currentCountdown
		// Force UI update only when countdown changes
		return app, tea.Batch(app.commandHandler.TickCmd(), func() tea.Msg { return updateUIMsg{} })
	}
	
	return app, app.commandHandler.TickCmd()
}


// View switching methods

func (app *BubbleApp) toggleView() (tea.Model, tea.Cmd) {
	currentView := app.viewManager.GetCurrentView()
	
	if currentView == ViewPending {
		// Switch to Recent view
		app.loading = true
		app.viewManager.SwitchToView(ViewRecent)
		// Initialize timer for recent jobs view
		app.commandHandler.UpdateTimerForView(ViewRecent)
		return app, app.jobService.RefreshJobs(app.ctx, ViewRecent)
	} else {
		// Switch to Pending view
		app.viewManager.SwitchToView(ViewPending)
		return app, nil
	}
}

func (app *BubbleApp) refreshCurrentView() (tea.Model, tea.Cmd) {
	app.loading = true
	return app, app.jobService.RefreshJobs(app.ctx, app.viewManager.GetCurrentView())
}

func (app *BubbleApp) showCancelConfirmation() (tea.Model, tea.Cmd) {
	jobs := app.getJobsForCurrentView()
	if len(jobs) == 0 {
		return app, nil
	}
	
	cursor := app.viewManager.GetCursor()
	if cursor >= len(jobs) {
		return app, nil
	}
	
	selectedJob := jobs[cursor]
	
	// Only allow cancellation for pending/in_progress jobs
	if selectedJob.Status != "waiting" && selectedJob.Status != "queued" && selectedJob.Status != "in_progress" {
		return app, nil
	}
	
	app.viewManager.ShowCancelConfirm(selectedJob)
	return app, nil
}

func (app *BubbleApp) showApprovalConfirmation() (tea.Model, tea.Cmd) {
	jobs := app.getJobsForCurrentView()
	if len(jobs) == 0 {
		return app, nil
	}
	
	cursor := app.viewManager.GetCursor()
	if cursor >= len(jobs) {
		return app, nil
	}
	
	selectedJob := jobs[cursor]
	
	// Only allow approval for waiting jobs that need approval
	if selectedJob.Status != "waiting" {
		return app, nil
	}
	
	app.viewManager.ShowApprovalConfirm(selectedJob)
	return app, nil
}

// Navigation methods

func (app *BubbleApp) moveCursorUp() (tea.Model, tea.Cmd) {
	app.viewManager.MoveCursor(-1, app.getMaxCursorPosition())
	return app, nil
}

func (app *BubbleApp) moveCursorDown() (tea.Model, tea.Cmd) {
	app.viewManager.MoveCursor(1, app.getMaxCursorPosition())
	return app, nil
}

func (app *BubbleApp) navigatePageLeft() (tea.Model, tea.Cmd) {
	if app.viewManager.GetCurrentView() == ViewRecent {
		app.viewManager.ChangePage(-1, len(app.recentJobs))
	}
	return app, nil
}

func (app *BubbleApp) navigatePageRight() (tea.Model, tea.Cmd) {
	if app.viewManager.GetCurrentView() == ViewRecent {
		app.viewManager.ChangePage(1, len(app.recentJobs))
	}
	return app, nil
}

// UI rendering methods

func (app *BubbleApp) renderMain() string {
	var content strings.Builder
	
	// Header
	content.WriteString(app.uiRenderer.RenderHeader(app.monitor))
	content.WriteString("\n")
	
	// View selector
	content.WriteString(app.uiRenderer.RenderViewSelector(
		app.viewManager.GetCurrentView(),
		len(app.jobs),
		len(app.recentJobs),
		app.viewManager,
	))
	content.WriteString("\n")
	
	// Job table
	jobs := app.getJobsForCurrentView()
	content.WriteString(app.uiRenderer.RenderJobTable(jobs, app.viewManager.GetCursor(), app.viewManager))
	content.WriteString("\n")
	
	// Pagination (only for Recent Jobs view)
	if app.viewManager.GetCurrentView() == ViewRecent {
		pagination := app.uiRenderer.RenderPagination(app.viewManager.GetCurrentView(), app.viewManager, len(app.recentJobs), jobs)
		if pagination != "" {
			content.WriteString(pagination)
		}
	}
	
	// Status/Info
	content.WriteString(app.uiRenderer.RenderStatus(app.errorMsg))
	
	return content.String()
}

// Helper methods

func (app *BubbleApp) getJobsForCurrentView() []scanner.JobStatus {
	return app.jobService.GetJobsForView(
		app.viewManager.GetCurrentView(),
		app.jobs,
		app.recentJobs,
		app.viewManager,
	)
}

func (app *BubbleApp) getMaxCursorPosition() int {
	return app.viewManager.GetMaxCursorPosition(app.jobs, app.recentJobs)
}

// RunBubbleApp runs the Bubble Tea application
func RunBubbleApp(m Monitor, config *AppConfig) error {
	app := NewBubbleApp(m, config)
	
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	
	return err
}