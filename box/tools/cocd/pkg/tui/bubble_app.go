package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/internal/monitor"
	"github.com/younsl/cocd/internal/scanner"
)

// BubbleApp is the main Bubble Tea application model
type BubbleApp struct {
	// Core dependencies
	monitor *monitor.Monitor
	config  *AppConfig
	ctx     context.Context
	cancel  context.CancelFunc
	
	// Component managers
	viewManager    *ViewManager
	uiComponents   *UIComponents
	commandHandler *CommandHandler
	
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
func NewBubbleApp(m *monitor.Monitor, config *AppConfig) *BubbleApp {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize component managers
	viewManager := NewViewManager()
	uiComponents := NewUIComponents(config)
	commandHandler := NewCommandHandler(m, config)
	
	app := &BubbleApp{
		monitor:        m,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		viewManager:    viewManager,
		uiComponents:   uiComponents,
		commandHandler: commandHandler,
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
		return app.handleKeyPress(msg)
		
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
		
	default:
		return app, nil
	}
}

// View renders the UI
func (app *BubbleApp) View() string {
	if app.showHelp {
		return app.uiComponents.RenderHelp(app.monitor)
	}
	
	if app.viewManager.IsShowingCancelConfirm() {
		if job := app.viewManager.GetCancelTargetJob(); job != nil {
			selection := app.viewManager.GetCancelSelection()
			return app.uiComponents.RenderCancelConfirm(*job, selection)
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
	
	// Hide cancel confirmation popup if it's showing (in case of cancel workflow error)
	if app.viewManager.IsShowingCancelConfirm() {
		app.viewManager.HideCancelConfirm()
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

// Key press handlers

func (app *BubbleApp) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle cancel confirmation popup keys first
	if app.viewManager.IsShowingCancelConfirm() {
		switch msg.String() {
		case "left":
			// Select "No"
			app.viewManager.SetCancelSelection(0)
			return app, nil
		case "right":
			// Select "Yes"
			app.viewManager.SetCancelSelection(1)
			return app, nil
		case "enter":
			// Confirm current selection
			if app.viewManager.IsCancelConfirmed() {
				return app, app.commandHandler.CancelWorkflow(app.ctx, app.viewManager)
			} else {
				app.viewManager.HideCancelConfirm()
				return app, nil
			}
		case "esc":
			// Cancel the cancellation
			app.viewManager.HideCancelConfirm()
			return app, nil
		case "y", "Y":
			// Legacy support - direct confirm
			return app, app.commandHandler.CancelWorkflow(app.ctx, app.viewManager)
		case "n", "N":
			// Legacy support - direct cancel
			app.viewManager.HideCancelConfirm()
			return app, nil
		default:
			return app, nil
		}
	}
	
	// If help is showing, any key closes it (except quit keys)
	if app.showHelp {
		switch msg.String() {
		case "ctrl+c", "q":
			app.cancel()
			return app, tea.Quit
		default:
			app.showHelp = false
			return app, nil
		}
	}
	
	switch msg.String() {
	case "ctrl+c", "q":
		app.cancel()
		return app, tea.Quit
		
	case "h", "?":
		app.showHelp = !app.showHelp
		return app, nil
		
	case "esc":
		return app, nil
		
	case "p":
		return app.switchToPendingView()
		
	case "l":
		return app.switchToRecentView()
		
	case "r":
		return app.refreshCurrentView()
		
	case "c":
		// Show cancel confirmation for selected job
		return app.showCancelConfirmation()
		
	case "up", "k":
		return app.moveCursorUp()
		
	case "down", "j":
		return app.moveCursorDown()
		
	case "enter":
		// Enter key - no action for now
		return app, nil
		
	case "o":
		// Open GitHub Actions page in browser
		return app, app.commandHandler.JumpToActions(app.viewManager, app.jobs, app.recentJobs)
		
	case "left":
		return app.navigatePageLeft()
		
	case "right":
		return app.navigatePageRight()
		
	default:
		return app, nil
	}
}

// View switching methods

func (app *BubbleApp) switchToPendingView() (tea.Model, tea.Cmd) {
	app.viewManager.SwitchToView(ViewPending)
	return app, nil
}

func (app *BubbleApp) switchToRecentView() (tea.Model, tea.Cmd) {
	app.loading = true
	app.viewManager.SwitchToView(ViewRecent)
	// Initialize timer for recent jobs view
	app.commandHandler.UpdateTimerForView(ViewRecent)
	return app, app.commandHandler.LoadRecentJobs(app.ctx)
}

func (app *BubbleApp) refreshCurrentView() (tea.Model, tea.Cmd) {
	app.loading = true
	if app.viewManager.GetCurrentView() == ViewPending {
		return app, app.commandHandler.LoadPendingJobs(app.ctx)
	} else {
		return app, app.commandHandler.LoadRecentJobs(app.ctx)
	}
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
	content.WriteString(app.uiComponents.RenderHeader(app.monitor))
	content.WriteString("\n")
	
	// View selector
	content.WriteString(app.uiComponents.RenderViewSelector(
		app.viewManager.GetCurrentView(),
		len(app.jobs),
		len(app.recentJobs),
		app.viewManager,
	))
	content.WriteString("\n")
	
	// Job table
	jobs := app.getJobsForCurrentView()
	content.WriteString(app.uiComponents.RenderJobTable(jobs, app.viewManager.GetCursor(), app.viewManager))
	content.WriteString("\n")
	
	// Pagination (only for Recent Jobs view)
	if app.viewManager.GetCurrentView() == ViewRecent {
		pagination := app.uiComponents.RenderPagination(app.viewManager.GetCurrentView(), app.viewManager, len(app.recentJobs), jobs)
		if pagination != "" {
			content.WriteString(pagination)
		}
	}
	
	// Status/Info
	content.WriteString(app.uiComponents.RenderStatus(app.errorMsg))
	
	return content.String()
}

// Helper methods

func (app *BubbleApp) getJobsForCurrentView() []scanner.JobStatus {
	switch app.viewManager.GetCurrentView() {
	case ViewPending:
		return app.viewManager.GetCombinedPendingJobs(app.jobs)
	case ViewRecent:
		return app.viewManager.GetPaginatedJobs(app.recentJobs)
	default:
		return []scanner.JobStatus{}
	}
}

func (app *BubbleApp) getMaxCursorPosition() int {
	return app.viewManager.GetMaxCursorPosition(app.jobs, app.recentJobs)
}

// RunBubbleApp runs the Bubble Tea application
func RunBubbleApp(m *monitor.Monitor, config *AppConfig) error {
	app := NewBubbleApp(m, config)
	
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	
	return err
}