package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/scanner"
)

// BubbleApp is the main Bubble Tea application model
type BubbleApp struct {
	monitor Monitor
	config  *AppConfig
	ctx     context.Context
	cancel  context.CancelFunc
	
	viewManager    ViewManagerInterface
	uiRenderer     UIRenderer
	commandHandler CommandHandlerInterface
	keyHandler     KeyHandler
	jobService     JobService
	
	jobs       []scanner.JobStatus
	recentJobs []scanner.JobStatus
	
	showHelp     bool
	loading      bool
	errorMsg     string
	lastUpdate   time.Time
	lastCountdown int
	width        int
	height       int
	showWaitingOnly bool
	scanning     bool
	
	jobsChan chan []scanner.JobStatus
	updateChan chan tea.Msg
}

// NewBubbleApp creates a new Bubble Tea application
func NewBubbleApp(m Monitor, config *AppConfig) *BubbleApp {
	ctx, cancel := context.WithCancel(context.Background())
	
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
		updateChan:     make(chan tea.Msg, 100),
		loading:        true,
	}
	
	commandHandler.UpdateTimerForView(ViewRecent)
	
	return app
}

// Init initializes the Bubble Tea application
func (app *BubbleApp) Init() tea.Cmd {
	return tea.Batch(
		app.commandHandler.StartMonitoring(app.ctx, app.jobsChan),
		app.commandHandler.LoadRecentJobsStreaming(app.ctx, app.updateChan),
		app.commandHandler.TickCmd(),
		app.listenForUpdates(),
	)
}

// listenForUpdates creates a command to continuously listen for streaming updates
func (app *BubbleApp) listenForUpdates() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		select {
		case msg := <-app.updateChan:
			return msg
		case <-app.ctx.Done():
			return nil
		}
	})
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
		
	case pendingJobsMsg:
		return app.handlePendingJobsMessage(msg)
		
	case recentJobsMsg:
		return app.handleRecentJobsMessage(msg)
		
	case errorMsg:
		return app.handleErrorMessage(msg)
		
	case tickMsg:
		return app.handleTickMessage(msg)
		
	case jobUpdateMsg:
		return app.handleJobUpdateMessage(msg)
		
	case recentJobUpdateMsg:
		return app.handleRecentJobUpdateMessage(msg)
	
	case scanProgressMsg:
		return app, nil
		
	case updateUIMsg:
		// Force UI update without changing data
		return app, nil
		
	case cancelSuccessMsg:
		app.viewManager.HideCancelConfirm()
		// Refresh the current view to see updated status
		return app.refreshCurrentView()
		
	case cancelProcessingMsg:
		app.viewManager.HideCancelConfirm()
		// Silently wait and then refresh to sync with GitHub
		return app, app.commandHandler.DelayedRefresh(3 * time.Second)
		
	case approvalSuccessMsg:
		app.viewManager.HideApprovalConfirm()
		// Refresh the current view to see updated status
		return app.refreshCurrentView()
		
	case approvalProcessingMsg:
		app.viewManager.HideApprovalConfirm()
		// Silently wait and then refresh to sync with GitHub
		return app, app.commandHandler.DelayedRefresh(3 * time.Second)
		
	case delayedRefreshMsg:
		// Perform the delayed refresh without showing loading indicator
		return app.silentRefreshCurrentView()
		
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
	// Deprecated: redirect to pending jobs handler for backwards compatibility
	return app.handlePendingJobsMessage(pendingJobsMsg(msg))
}

func (app *BubbleApp) handlePendingJobsMessage(msg pendingJobsMsg) (tea.Model, tea.Cmd) {
	newJobs := []scanner.JobStatus(msg)
	
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
	
	app.commandHandler.UpdateTimerForView(ViewRecent)
	
	return app, nil
}

func (app *BubbleApp) handleErrorMessage(msg errorMsg) (tea.Model, tea.Cmd) {
	app.errorMsg = string(msg)
	app.loading = false
	
	if app.viewManager.IsShowingCancelConfirm() {
		app.viewManager.HideCancelConfirm()
	}
	if app.viewManager.IsShowingApprovalConfirm() {
		app.viewManager.HideApprovalConfirm()
	}
	
	return app, nil
}

func (app *BubbleApp) handleTickMessage(msg tickMsg) (tea.Model, tea.Cmd) {
	app.monitor.GetProgressTracker().UpdateScanCountdown()
	
	progress := app.monitor.GetScanProgress()
	isScanning := progress.ScanMode != "Idle" && progress.ScanMode != "Completed"
	
	// Only auto-refresh Recent Jobs if we're on Recent Jobs view
	currentView := app.viewManager.GetCurrentView()
	if time.Since(app.lastUpdate) > 30*time.Second && !isScanning && currentView == ViewRecent {
		app.loading = true
		return app, tea.Batch(
			app.commandHandler.TickCmd(),
			app.commandHandler.LoadRecentJobsStreaming(app.ctx, app.updateChan),
		)
	}
	
	currentCountdown := app.monitor.GetScanProgress().ScanCountdown
	if currentCountdown != app.lastCountdown {
		app.lastCountdown = currentCountdown
		return app, tea.Batch(app.commandHandler.TickCmd(), func() tea.Msg { return updateUIMsg{} })
	}
	
	return app, app.commandHandler.TickCmd()
}



func (app *BubbleApp) toggleView() (tea.Model, tea.Cmd) {
	currentView := app.viewManager.GetCurrentView()
	
	if currentView == ViewPending {
		app.viewManager.SwitchToView(ViewRecent)
		app.commandHandler.UpdateTimerForView(ViewRecent)
		
		app.loading = true
		return app, app.commandHandler.LoadRecentJobsStreaming(app.ctx, app.updateChan)
	} else {
		app.viewManager.SwitchToView(ViewPending)
		
		return app, nil
	}
}

func (app *BubbleApp) refreshCurrentView() (tea.Model, tea.Cmd) {
	currentView := app.viewManager.GetCurrentView()
	app.loading = true
	
	if currentView == ViewRecent {
			return app, app.commandHandler.LoadRecentJobsStreaming(app.ctx, app.updateChan)
	}
	
	return app, app.jobService.RefreshJobs(app.ctx, currentView)
}

func (app *BubbleApp) silentRefreshCurrentView() (tea.Model, tea.Cmd) {
	currentView := app.viewManager.GetCurrentView()
	// Don't show loading indicator for silent refresh
	
	if currentView == ViewRecent {
		return app, app.commandHandler.LoadRecentJobsStreaming(app.ctx, app.updateChan)
	}
	
	return app, app.jobService.RefreshJobs(app.ctx, currentView)
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
	
	if selectedJob.Status != "waiting" {
		return app, nil
	}
	
	app.viewManager.ShowApprovalConfirm(selectedJob)
	return app, nil
}


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


func (app *BubbleApp) renderMain() string {
	var content strings.Builder
	
	content.WriteString(app.uiRenderer.RenderHeader(app.monitor))
	content.WriteString("\n")
	
	content.WriteString(app.uiRenderer.RenderViewSelector(
		app.viewManager.GetCurrentView(),
		len(app.jobs),
		len(app.recentJobs),
		app.viewManager,
	))
	content.WriteString("\n")
	
	jobs := app.getJobsForCurrentView()
	content.WriteString(app.uiRenderer.RenderJobTable(jobs, app.viewManager.GetCursor(), app.viewManager))
	content.WriteString("\n")
	
	if app.viewManager.GetCurrentView() == ViewRecent {
		pagination := app.uiRenderer.RenderPagination(app.viewManager.GetCurrentView(), app.viewManager, len(app.recentJobs), jobs)
		if pagination != "" {
			content.WriteString(pagination)
		}
	}
	
	content.WriteString(app.uiRenderer.RenderStatus(app.errorMsg))
	
	return content.String()
}

func (app *BubbleApp) handleJobUpdateMessage(msg jobUpdateMsg) (tea.Model, tea.Cmd) {
	update := monitor.JobUpdate(msg)
	
	if update.Error != nil {
		app.loading = false
		app.errorMsg = update.Error.Error()
		return app, app.listenForUpdates() // Continue listening
	}
	
	if len(update.Jobs) > 0 {
		currentView := app.viewManager.GetCurrentView()
		
		for _, job := range update.Jobs {
			existsInJobs := false
			for _, existingJob := range app.jobs {
				if existingJob.RunID == job.RunID && existingJob.ID == job.ID {
					existsInJobs = true
					break
				}
			}
			if !existsInJobs {
				app.jobs = append(app.jobs, job)
			}
			
			existsInRecent := false
			for _, existingJob := range app.recentJobs {
				if existingJob.RunID == job.RunID && existingJob.ID == job.ID {
					existsInRecent = true
					break
				}
			}
			if !existsInRecent {
				app.recentJobs = append(app.recentJobs, job)
			}
		}
		
		if currentView == ViewPending {
			monitor.SortJobsByTime(app.jobs, false)
		} else {
			monitor.SortJobsByTime(app.recentJobs, true)
		}
		
		app.loading = false
		app.lastUpdate = time.Now()
		app.errorMsg = ""
	}
	
	return app, app.listenForUpdates()
}


func (app *BubbleApp) handleRecentJobUpdateMessage(msg recentJobUpdateMsg) (tea.Model, tea.Cmd) {
	update := monitor.JobUpdate(msg)
	
	if update.Error != nil {
		app.loading = false
		app.errorMsg = update.Error.Error()
		return app, app.listenForUpdates()
	}
	
	// Handle repository completion for real-time updates
	if update.CompletedRepo != "" {
		// Add or update jobs from the completed repository
		if len(update.Jobs) > 0 {
			// Remove old jobs from the same repository first
			var filteredRecentJobs []scanner.JobStatus
			for _, job := range app.recentJobs {
				if job.Repository != update.CompletedRepo {
					filteredRecentJobs = append(filteredRecentJobs, job)
				}
			}
			app.recentJobs = filteredRecentJobs
			
			// Add new jobs from this repository
			for _, job := range update.Jobs {
				app.recentJobs = append(app.recentJobs, job)
				
				// Also update pending jobs if status is waiting
				if job.Status == "waiting" {
					existsInPending := false
					for i, existingJob := range app.jobs {
						if existingJob.RunID == job.RunID && existingJob.ID == job.ID {
							app.jobs[i] = job
							existsInPending = true
							break
						}
					}
					
					if !existsInPending {
						app.jobs = append(app.jobs, job)
					}
				}
			}
			
			// Sort jobs after each update
			monitor.SortJobsByTime(app.recentJobs, true)
			monitor.SortJobsByTime(app.jobs, false)
		}
	}
	
	// Check if scan is completed
	if update.Progress.ScanMode == "Completed" {
		app.loading = false
		app.lastUpdate = time.Now()
		app.errorMsg = ""
	}
	
	return app, app.listenForUpdates()
}


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