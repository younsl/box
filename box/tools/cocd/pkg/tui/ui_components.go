package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/scanner"
)

// UIComponents handles UI rendering
type UIComponents struct {
	config *AppConfig
}

// NewUIComponents creates new UI components
func NewUIComponents(config *AppConfig) UIRenderer {
	return &UIComponents{
		config: config,
	}
}

// RenderHeader renders the header section
func (ui *UIComponents) RenderHeader(monitor Monitor) string {
	serverName := ui.config.ServerURL
	if serverName == "" || serverName == "https://api.github.com" {
		serverName = "GitHub.com"
	}
	
	org := ui.config.Org
	if ui.config.Repo != "" {
		org = fmt.Sprintf("%s/%s", ui.config.Org, ui.config.Repo)
	}
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Bold(true).
		Padding(0, 1)
	
	progress := monitor.GetScanProgress()
	titleText := "CoCD"
	if ui.config.Version != "" && ui.config.Version != "dev" {
		titleText = fmt.Sprintf("CoCD v%s", ui.config.Version)
	} else if ui.config.Version == "dev" {
		titleText = "CoCD dev"
	}
	title := headerStyle.Render(titleText)
	memory := fmt.Sprintf("Mem: %s", progress.MemoryUsage)
	server := fmt.Sprintf("Server: %s", serverName)
	
	// Get username
	username := "unknown"
	if user, err := monitor.GetAuthenticatedUser(context.Background()); err == nil && user != "" {
		username = user
	}
	
	userInfo := fmt.Sprintf("User: %s", username)
	organization := fmt.Sprintf("Org: %s", org)
	
	status := ui.getConnectionStatus(false, "")
	
	scanInfo := ui.getScanInfo(progress)
	timerInfo := ui.getTimerInfo(progress)
	keyBindings := ui.getKeyBindings()
	
	return fmt.Sprintf("%s  %s  %s  %s  %s  Status: %s\n%s\n%s\n%s", 
		title, memory, server, organization, userInfo, status, scanInfo, timerInfo, keyBindings)
}

// RenderViewSelector renders the view selector
func (ui *UIComponents) RenderViewSelector(currentView ViewType, pendingCount, recentCount int, vm ViewManagerInterface) string {
	pendingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	recentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	
	if currentView == ViewPending {
		pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	} else {
		recentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	}
	
	pending := pendingStyle.Render(fmt.Sprintf("Approval Waiting Jobs [%d]", pendingCount))
	recent := recentStyle.Render(fmt.Sprintf("Recent Jobs [%d]", recentCount))
	
	return fmt.Sprintf("%s  %s", pending, recent)
}

// RenderJobTable renders the job table
func (ui *UIComponents) RenderJobTable(jobs []scanner.JobStatus, cursor int, vm ViewManagerInterface) string {
	var b strings.Builder
	
	// Calculate dynamic column widths based on content
	var columnWidths []int
	var ageWidth int
	
	if len(jobs) == 0 {
		// Use default column widths when no jobs
		columnWidths = []int{25, 20, 8, 12, 20, 12} // repo, job, id, status, branch, actor
		ageWidth = 8
	} else {
		columnWidths = ui.calculateColumnWidths(jobs)
		ageWidth = ui.calculateAgeColumnWidth(jobs)
	}
	
	repoWidth := columnWidths[0]
	jobWidth := columnWidths[1]
	idWidth := columnWidths[2]
	statusWidth := columnWidths[3]
	branchWidth := columnWidths[4]
	actorWidth := columnWidths[5]
	
	// Always render table header
	ui.renderTableHeader(&b, repoWidth, jobWidth, idWidth, statusWidth, branchWidth, actorWidth, ageWidth)
	
	if len(jobs) == 0 {
		// Show "No jobs found" message after header
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Padding(1, 0)
		b.WriteString(emptyStyle.Render("No jobs found"))
	} else {
		// Render table rows
		for i, job := range jobs {
			ui.renderTableRow(&b, job, i, cursor, vm, repoWidth, jobWidth, idWidth, statusWidth, branchWidth, actorWidth, ageWidth)
		}
	}
	
	return b.String()
}

// RenderStatus renders the status information
func (ui *UIComponents) RenderStatus(errorMsg string) string {
	var b strings.Builder
	
	if errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Error: " + errorMsg))
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	
	return b.String()
}

// RenderPagination renders pagination dots positioned under the AGE column
func (ui *UIComponents) RenderPagination(currentView ViewType, vm ViewManagerInterface, totalJobs int, jobs []scanner.JobStatus) string {
	if currentView != ViewRecent || totalJobs == 0 {
		return ""
	}
	
	currentPage, perPage := vm.GetPageInfo()
	totalPages := (totalJobs + perPage - 1) / perPage
	
	if totalPages <= 1 {
		return ""
	}
	
	var dots strings.Builder
	
	// Create pagination dots
	for i := 0; i < totalPages; i++ {
		if i == currentPage {
			// Current page - filled dot with highlight color
			dots.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("●"))
		} else {
			// Other pages - empty dot with muted color
			dots.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("●"))
		}
	}
	
	paginationText := dots.String()
	
	// Calculate column widths to position pagination under AGE column
	columnWidths := ui.calculateColumnWidths(jobs)
	ageWidth := ui.calculateAgeColumnWidth(jobs)
	
	// Calculate total width up to AGE column (all columns + spaces between them)
	totalWidthUpToAge := 0
	for i := 0; i < len(columnWidths); i++ {
		totalWidthUpToAge += columnWidths[i]
		if i < len(columnWidths)-1 {
			totalWidthUpToAge += 1 // Space between columns
		}
	}
	totalWidthUpToAge += 1 // Space before AGE column
	
	// Center pagination dots within the AGE column
	ageColumnCenter := totalWidthUpToAge + (ageWidth / 2)
	paginationStart := ageColumnCenter - (totalPages / 2)
	
	if paginationStart < 0 {
		paginationStart = 0
	}
	
	return strings.Repeat(" ", paginationStart) + paginationText
}

// RenderHelp renders the help screen
func (ui *UIComponents) RenderHelp(monitor Monitor) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4"))
	
	interval := monitor.GetUpdateInterval()
	intervalStr := fmt.Sprintf("%d sec (auto)", interval)
	
	help := fmt.Sprintf(`CoCD - GitHub Actions Monitor

KEY BINDINGS:
  q, Ctrl+C    Quit
  t            Toggle between Approval Waiting Jobs and Recent Jobs
  r            Refresh current view
  a            Approve selected deployment (with confirmation)
  c            Cancel selected workflow (with confirmation)
  h, ?         Toggle this help
  ↑/↓, k/j     Navigate jobs (k=up, j=down)
  ←/→          Navigate pages (Recent Jobs only)
  o            Open GitHub Actions page in browser

SCAN SETTINGS:
SETTING              SMART SCAN             RECENT JOBS
Interval             %-22s Manual only
Target Repos         Top 200 (7d)           Top 100 active
Workers              2 concurrent           2 concurrent
Timeout              60 seconds             90 seconds
API Filter           status="waiting"       All runs
Cache                Repo/Env (60m/5m)      Repo list (60m)
Result Limit         All waiting            Last 200 jobs

Press any key to continue...`, intervalStr)
	
	return helpStyle.Render(help)
}


// RenderApprovalConfirm renders the approval confirmation popup
func (ui *UIComponents) RenderApprovalConfirm(job scanner.JobStatus, selection int) string {
	// Create a centered popup with a more professional design
	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("2")).
		Width(60).
		Align(lipgloss.Center)
	
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true).
		Align(lipgloss.Center).
		Render("⚠️  Confirm Deployment Approval")
	
	jobInfo := fmt.Sprintf(
		"Repository: %s\nWorkflow: %s\nStatus: %s",
		job.Repository,
		job.WorkflowName,
		job.Status,
	)
	
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Align(lipgloss.Center).
		Render("This will approve the deployment to production!")
	
	// Add approval message preview
	ch := NewCommandHandler(nil, ui.config)
	approvalMessage := ch.(*CommandHandler).generateApprovalMessage()
	messagePreview := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Background(lipgloss.Color("8")).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("Message: %s", approvalMessage))
	
	// Create interactive Yes/No buttons with consistent width
	buttonWidth := 8
	
	// Simple design with color only on selected button
	var noButton, yesButton string
	if selection == 0 { // No selected - green background
		noButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("2")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Render("No")
		yesButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Render("Yes")
	} else { // Yes selected - red background for approval
		noButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Render("No")
		yesButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("1")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Render("Yes")
	}
	
	// Create button container with proper spacing
	buttonContainer := lipgloss.JoinHorizontal(
		lipgloss.Center,
		noButton,
		strings.Repeat(" ", 4), // Space between buttons
		yesButton,
	)
	
	buttons := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(buttonContainer)
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center).
		Render("Use ←/→ to select, Enter to confirm, Esc to cancel")
	
	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s", title, jobInfo, warning, messagePreview, buttons, instructions)
	
	return confirmStyle.Render(content)
}

// RenderCancelConfirm renders the cancel confirmation popup with interactive selection
func (ui *UIComponents) RenderCancelConfirm(job scanner.JobStatus, selection int) string {
	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("0")).
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Align(lipgloss.Center)
	
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Align(lipgloss.Center).
		Render("⚠️  Confirm Cancel Workflow")
	
	jobInfo := fmt.Sprintf("Repository: %s\nWorkflow: %s\nRun #%d\nActor: %s", 
		job.Repository, job.WorkflowName, job.RunNumber, job.Actor)
	
	warning := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Italic(true).
		Align(lipgloss.Center).
		Render("This action cannot be undone!")
	
	// Create interactive Yes/No buttons with consistent width
	buttonWidth := 8
	
	// Simple design with color only on selected button
	var noButton, yesButton string
	if selection == 0 { // No selected - green background
		noButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("2")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Render("No")
		yesButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Render("Yes")
	} else { // Yes selected - red background
		noButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Render("No")
		yesButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("1")).
			Padding(0, 1).
			Width(buttonWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Render("Yes")
	}
	
	// Create button container with proper spacing
	buttonContainer := lipgloss.JoinHorizontal(
		lipgloss.Center,
		noButton,
		strings.Repeat(" ", 4), // Space between buttons
		yesButton,
	)
	
	buttons := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(buttonContainer)
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center).
		Render("Use ←/→ to select, Enter to confirm, Esc to cancel")
	
	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s", title, jobInfo, warning, buttons, instructions)
	
	return confirmStyle.Render(content)
}

// Helper functions

func (ui *UIComponents) getConnectionStatus(loading bool, errorMsg string) string {
	if loading {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("Connecting")
	} else if errorMsg != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Error")
	} else {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("Connected")
	}
}

func (ui *UIComponents) getScanInfo(progress monitor.ScanProgress) string {
	var scanInfo string
	if progress.TotalRepos > 0 {
		// Simplified repository display with focus on key metrics
		var repoInfo string
		
		// Consistent repository display format
		// Show real-time scan progress: completed/total
		targetRepos := progress.LimitedRepos
		if targetRepos == 0 || targetRepos >= progress.ValidRepos {
			targetRepos = progress.ActiveRepos
		}
		
		if progress.ArchivedRepos > 0 {
			repoInfo = fmt.Sprintf("Repos: %d/%d (%d archived)", 
				progress.CompletedRepos, targetRepos, progress.ArchivedRepos)
		} else {
			repoInfo = fmt.Sprintf("Repos: %d/%d", 
				progress.CompletedRepos, targetRepos)
		}
		
		scanInfo = fmt.Sprintf("Mode: %s | %s | Cache: %s", 
			progress.ScanMode, repoInfo, progress.CacheStatus)
	} else {
		scanInfo = fmt.Sprintf("Mode: Idle | Cache: %s", progress.CacheStatus)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(scanInfo)
}

func (ui *UIComponents) getTimerInfo(progress monitor.ScanProgress) string {
	var timerInfo string
	if progress.NextScanAt != nil {
		nextType := "Fast"
		if progress.IsNextScanFull {
			nextType = "Full"
		}
		countdown := progress.ScanCountdown
		if countdown < 0 {
			countdown = 0
		}
		
		
		timerInfo = fmt.Sprintf("Next %s scan in %ds", nextType, countdown)
	} else {
		// Show loading state with duration
		stateDuration := progress.StateDuration
		if stateDuration < 0 {
			stateDuration = 0
		}
		timerInfo = fmt.Sprintf("Scanning... (%ds)", stateDuration)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(timerInfo)
}

func (ui *UIComponents) getKeyBindings() string {
	keyBindings := "Keys: [t]oggle view [r]efresh [a]pprove [c]ancel [o]pen browser [h]elp [q]uit [↑↓] navigate"
	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(keyBindings)
}

func (ui *UIComponents) renderTableHeader(b *strings.Builder, repoWidth, jobWidth, idWidth, statusWidth, branchWidth, actorWidth, ageWidth int) {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)
	
	configs := ui.getColumnConfigs()
	widths := []int{repoWidth, jobWidth, idWidth, statusWidth, branchWidth, actorWidth}
	
	// Create properly padded headers
	var headers []string
	for i, config := range configs {
		headers = append(headers, ui.padString(config.Header, widths[i]))
	}
	ageHeader := ui.padString("AGE", ageWidth)
	headers = append(headers, ageHeader)
	
	// Apply styling to entire header row to maintain alignment
	headerRow := strings.Join(headers, " ")
	
	b.WriteString(headerStyle.Render(headerRow))
	b.WriteString("\n")
}

func (ui *UIComponents) renderTableRow(b *strings.Builder, job scanner.JobStatus, i, cursor int, vm ViewManagerInterface, repoWidth, jobWidth, idWidth, statusWidth, branchWidth, actorWidth, ageWidth int) {
	// Truncate and pad columns
	repo := ui.padString(ui.truncate(job.Repository, repoWidth), repoWidth)
	jobName := ui.padString(ui.truncate(job.Name, jobWidth), jobWidth)
	jobID := ui.padString(fmt.Sprintf("#%d", job.RunNumber), idWidth)
	status := ui.padString(job.Status, statusWidth)
	branch := ui.padString(ui.truncate(job.Branch, branchWidth), branchWidth)
	actor := ui.padString(ui.truncate(job.Actor, actorWidth), actorWidth)
	age := ui.padString(ui.formatAge(job.StartedAt), ageWidth)
	
	// Build row string
	rowString := fmt.Sprintf("%s %s %s %s %s %s %s",
		repo, jobName, jobID, status, branch, actor, age)
	
	// Apply styles based on priority: cursor > newly highlighted > completed > normal
	if i == cursor {
		// Cursor selection - highest priority (blue background)
		rowStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
		b.WriteString(rowStyle.Render(rowString))
	} else if vm.IsJobHighlighted(job) {
		// Newly scanned job - second priority (green background with fade effect)
		// Use a subtle green background to indicate newly discovered job
		rowStyle := lipgloss.NewStyle().Background(lipgloss.Color("2")).Foreground(lipgloss.Color("0"))
		b.WriteString(rowStyle.Render(rowString))
	} else {
		// Check if this job is completed (from our tracking)
		isCompleted := vm.IsJobCompleted(job)
		
		if isCompleted {
			// Completed jobs: gray out everything
			rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			b.WriteString(rowStyle.Render(rowString))
		} else {
			// Active jobs: normal coloring with status color only for STATUS column
			var statusColored string
			switch job.Status {
			case "waiting":
				statusColored = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(status)
			case "in_progress":
				statusColored = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(status)
			case "completed", "success":
				statusColored = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(status)
			case "failure":
				statusColored = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(status)
			case "cancelled":
				statusColored = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(status)
			default:
				statusColored = status
			}
			
			// Build row with only status column colored
			b.WriteString(fmt.Sprintf("%s %s %s %s %s %s %s",
				repo, jobName, jobID, statusColored, branch, actor, age))
		}
	}
	
	b.WriteString("\n")
}

// Text formatting utilities

func (ui *UIComponents) truncate(s string, width int) string {
	if runewidth.StringWidth(s) <= width {
		return s
	}
	
	// If width is too small to even fit "..", return empty string
	if width < 2 {
		return ""
	}
	
	// Find the position to truncate while considering display width
	var truncated strings.Builder
	currentWidth := 0
	
	for _, r := range s {
		runeWidth := runewidth.RuneWidth(r)
		// Reserve space for ".." suffix
		if currentWidth+runeWidth+2 > width {
			break
		}
		truncated.WriteRune(r)
		currentWidth += runeWidth
	}
	
	result := truncated.String() + ".."
	
	// Double-check the final width and adjust if necessary
	if runewidth.StringWidth(result) > width {
		// Fallback: trim more characters if still too wide
		text := truncated.String()
		for runewidth.StringWidth(text)+2 > width && len(text) > 0 {
			runes := []rune(text)
			if len(runes) > 0 {
				text = string(runes[:len(runes)-1])
			} else {
				break
			}
		}
		return text + ".."
	}
	
	return result
}

func (ui *UIComponents) padString(s string, width int) string {
	currentWidth := runewidth.StringWidth(s)
	if currentWidth >= width {
		return s
	}
	padding := width - currentWidth
	if padding < 0 {
		padding = 0
	}
	return s + strings.Repeat(" ", padding)
}

func (ui *UIComponents) formatAge(startedAt *time.Time) string {
	if startedAt == nil {
		return "N/A"
	}
	
	now := time.Now()
	diff := now.Sub(*startedAt)
	
	if diff < time.Minute {
		return fmt.Sprintf("%ds", int(diff.Seconds()))
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(diff.Hours()/24))
	}
}

// Column configuration structure
type ColumnConfig struct {
	Header       string
	MaxWidth     int
	MinimumWidth int
}

// Column width calculations

func (ui *UIComponents) getColumnConfigs() []ColumnConfig {
	return []ColumnConfig{
		{Header: "REPOSITORY", MaxWidth: 30, MinimumWidth: 12},
		{Header: "JOB NAME", MaxWidth: 50, MinimumWidth: 20},
		{Header: "RNO", MaxWidth: 10, MinimumWidth: 4},
		{Header: "STATUS", MaxWidth: 15, MinimumWidth: 10},
		{Header: "BRANCH", MaxWidth: 25, MinimumWidth: 10},
		{Header: "ACTOR", MaxWidth: 20, MinimumWidth: 10},
	}
}

func (ui *UIComponents) calculateColumnWidths(jobs []scanner.JobStatus) []int {
	configs := ui.getColumnConfigs()
	widths := make([]int, len(configs))
	
	// Initialize with header widths
	for i, config := range configs {
		widths[i] = runewidth.StringWidth(config.Header)
	}
	
	// Calculate max content width for each column
	for _, job := range jobs {
		contents := []string{
			job.Repository,
			job.Name,
			fmt.Sprintf("#%d", job.RunNumber),
			job.Status,
			job.Branch,
			job.Actor,
		}
		
		for i, content := range contents {
			width := runewidth.StringWidth(content)
			if width > widths[i] {
				widths[i] = width
			}
		}
	}
	
	// Apply constraints
	for i, config := range configs {
		// Apply maximum width constraints
		if widths[i] > config.MaxWidth {
			widths[i] = config.MaxWidth
		}
		
		// Ensure minimum widths
		if widths[i] < config.MinimumWidth {
			widths[i] = config.MinimumWidth
		}
	}
	
	return widths
}

func (ui *UIComponents) calculateAgeColumnWidth(jobs []scanner.JobStatus) int {
	// Minimum width for "AGE" header
	minWidth := runewidth.StringWidth("AGE")
	
	// Calculate max content width for age column
	for _, job := range jobs {
		ageText := ui.formatAge(job.StartedAt)
		width := runewidth.StringWidth(ageText)
		if width > minWidth {
			minWidth = width
		}
	}
	
	// Cap at maximum reasonable width
	if minWidth > 10 {
		minWidth = 10
	}
	
	return minWidth
}