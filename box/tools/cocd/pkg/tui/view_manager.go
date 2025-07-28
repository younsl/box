package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/younsl/cocd/internal/scanner"
)

// ViewType represents different view types
type ViewType string

const (
	ViewPending ViewType = "pending"
	ViewRecent  ViewType = "recent"
)

// ViewManager handles view-specific logic
type ViewManager struct {
	currentView ViewType
	cursor      int
	
	// Pagination for recent jobs
	recentJobsPage    int
	recentJobsPerPage int
	
	// Completed jobs tracking
	completedJobs map[string]scanner.JobStatus
	
	// Loading state tracking
	isLoading bool
	
	// Cancel confirmation popup state
	showCancelConfirm bool
	cancelTargetJob   *scanner.JobStatus
	cancelSelection   int // 0 = No, 1 = Yes
}

// NewViewManager creates a new view manager
func NewViewManager() *ViewManager {
	return &ViewManager{
		currentView:       ViewPending,
		cursor:            0,
		recentJobsPage:    0,
		recentJobsPerPage: 50,
		completedJobs:     make(map[string]scanner.JobStatus),
	}
}

// SwitchToView switches to the specified view
func (vm *ViewManager) SwitchToView(viewType ViewType) {
	vm.currentView = viewType
	vm.cursor = 0
	if viewType == ViewRecent {
		vm.recentJobsPage = 0
	}
}

// GetCurrentView returns the current view type
func (vm *ViewManager) GetCurrentView() ViewType {
	return vm.currentView
}

// GetCursor returns the current cursor position
func (vm *ViewManager) GetCursor() int {
	return vm.cursor
}

// SetCursor sets the cursor position
func (vm *ViewManager) SetCursor(cursor int) {
	vm.cursor = cursor
}

// MoveCursor moves the cursor up or down
func (vm *ViewManager) MoveCursor(direction int, maxItems int) {
	newCursor := vm.cursor + direction
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= maxItems {
		newCursor = maxItems - 1
	}
	vm.cursor = newCursor
}

// GetPageInfo returns pagination information for recent jobs
func (vm *ViewManager) GetPageInfo() (page int, perPage int) {
	return vm.recentJobsPage, vm.recentJobsPerPage
}

// ChangePage changes the page for recent jobs
func (vm *ViewManager) ChangePage(direction int, totalItems int) {
	totalPages := vm.getTotalPages(totalItems)
	newPage := vm.recentJobsPage + direction
	
	if newPage < 0 {
		newPage = 0
	}
	if newPage >= totalPages {
		newPage = totalPages - 1
	}
	
	vm.recentJobsPage = newPage
	vm.cursor = 0
}

// getTotalPages calculates total pages for recent jobs
func (vm *ViewManager) getTotalPages(totalItems int) int {
	if totalItems == 0 {
		return 1
	}
	return (totalItems + vm.recentJobsPerPage - 1) / vm.recentJobsPerPage
}

// GetPaginatedJobs returns paginated jobs for recent view
func (vm *ViewManager) GetPaginatedJobs(jobs []scanner.JobStatus) []scanner.JobStatus {
	if vm.currentView != ViewRecent {
		return jobs
	}
	
	start := vm.recentJobsPage * vm.recentJobsPerPage
	end := start + vm.recentJobsPerPage
	
	if start >= len(jobs) {
		return []scanner.JobStatus{}
	}
	
	if end > len(jobs) {
		end = len(jobs)
	}
	
	return jobs[start:end]
}

// TrackCompletedJobs tracks jobs that have moved from pending to completed
func (vm *ViewManager) TrackCompletedJobs(currentJobs, newJobs []scanner.JobStatus) {
	// Track completed jobs (jobs that were pending but are now completed)
	for _, currentJob := range currentJobs {
		currentKey := fmt.Sprintf("%s:%d:%d", currentJob.Repository, currentJob.RunID, currentJob.ID)
		
		// Check if this job is still pending in the new jobs
		stillPending := false
		for _, newJob := range newJobs {
			newKey := fmt.Sprintf("%s:%d:%d", newJob.Repository, newJob.RunID, newJob.ID)
			if currentKey == newKey {
				stillPending = true
				break
			}
		}
		
		// If job is no longer pending, add it to completed jobs
		if !stillPending && (currentJob.Status == "waiting" || currentJob.Status == "queued" || currentJob.Status == "in_progress") {
			// Mark as completed and add timestamp
			completedJob := currentJob
			completedJob.Status = "completed"
			completedJob.CompletedAt = &time.Time{}
			*completedJob.CompletedAt = time.Now()
			vm.completedJobs[currentKey] = completedJob
		}
	}
}

// GetCombinedPendingJobs returns combined pending and completed jobs
func (vm *ViewManager) GetCombinedPendingJobs(jobs []scanner.JobStatus) []scanner.JobStatus {
	// Combine current pending jobs with completed jobs
	combinedJobs := make([]scanner.JobStatus, len(jobs))
	copy(combinedJobs, jobs)
	
	// Track existing job keys to avoid duplicates
	existingKeys := make(map[string]bool)
	for _, job := range jobs {
		key := fmt.Sprintf("%s:%d:%d", job.Repository, job.RunID, job.ID)
		existingKeys[key] = true
	}
	
	// Add completed jobs to the list (only if not already present)
	for _, completedJob := range vm.completedJobs {
		key := fmt.Sprintf("%s:%d:%d", completedJob.Repository, completedJob.RunID, completedJob.ID)
		if !existingKeys[key] {
			combinedJobs = append(combinedJobs, completedJob)
		}
	}
	
	// Sort jobs: active jobs first (by detection time), then completed jobs
	sort.Slice(combinedJobs, func(i, j int) bool {
		// First sort by completion status (active jobs first)
		iCompleted := vm.isJobCompleted(combinedJobs[i])
		jCompleted := vm.isJobCompleted(combinedJobs[j])
		
		if iCompleted != jCompleted {
			return !iCompleted // Active jobs first (not completed)
		}
		
		// For active jobs, sort by detection time (started time)
		// Earlier detected jobs come first
		if !iCompleted && !jCompleted {
			// Both are active jobs - sort by started time
			if combinedJobs[i].StartedAt != nil && combinedJobs[j].StartedAt != nil {
				return combinedJobs[i].StartedAt.Before(*combinedJobs[j].StartedAt)
			}
			// If one doesn't have started time, put it at the end
			if combinedJobs[i].StartedAt == nil {
				return false
			}
			if combinedJobs[j].StartedAt == nil {
				return true
			}
		}
		
		// For completed jobs, sort by completion time (most recent first)
		if iCompleted && jCompleted {
			if combinedJobs[i].CompletedAt != nil && combinedJobs[j].CompletedAt != nil {
				return combinedJobs[i].CompletedAt.After(*combinedJobs[j].CompletedAt)
			}
		}
		
		return false
	})
	
	return combinedJobs
}

// IsJobCompleted checks if a job is in the completed jobs map
func (vm *ViewManager) IsJobCompleted(job scanner.JobStatus) bool {
	return vm.isJobCompleted(job)
}

// isJobCompleted checks if a job is in the completed jobs map
func (vm *ViewManager) isJobCompleted(job scanner.JobStatus) bool {
	key := fmt.Sprintf("%s:%d:%d", job.Repository, job.RunID, job.ID)
	_, exists := vm.completedJobs[key]
	return exists
}

// GetViewTitle returns the title for the current view
func (vm *ViewManager) GetViewTitle(jobCount int) string {
	switch vm.currentView {
	case ViewPending:
		return fmt.Sprintf("Approval Waiting Jobs [%d]", jobCount)
	case ViewRecent:
		return fmt.Sprintf("Recent Jobs [%d]", jobCount)
	default:
		return "Unknown View"
	}
}

// GetMaxCursorPosition returns the maximum cursor position for current view
func (vm *ViewManager) GetMaxCursorPosition(pendingJobs, recentJobs []scanner.JobStatus) int {
	switch vm.currentView {
	case ViewPending:
		// Include both pending and completed jobs
		return len(pendingJobs) + len(vm.completedJobs)
	case ViewRecent:
		return len(vm.GetPaginatedJobs(recentJobs))
	default:
		return 0
	}
}

// ShowCancelConfirm shows the cancel confirmation popup
func (vm *ViewManager) ShowCancelConfirm(job scanner.JobStatus) {
	vm.showCancelConfirm = true
	vm.cancelTargetJob = &job
	vm.cancelSelection = 0 // Default to "No"
}

// HideCancelConfirm hides the cancel confirmation popup
func (vm *ViewManager) HideCancelConfirm() {
	vm.showCancelConfirm = false
	vm.cancelTargetJob = nil
	vm.cancelSelection = 0
}

// IsShowingCancelConfirm returns whether cancel confirmation is showing
func (vm *ViewManager) IsShowingCancelConfirm() bool {
	return vm.showCancelConfirm
}

// GetCancelTargetJob returns the job to be cancelled
func (vm *ViewManager) GetCancelTargetJob() *scanner.JobStatus {
	return vm.cancelTargetJob
}

// ToggleCancelSelection toggles between Yes/No selection
func (vm *ViewManager) ToggleCancelSelection() {
	if vm.cancelSelection == 0 {
		vm.cancelSelection = 1
	} else {
		vm.cancelSelection = 0
	}
}

// SetCancelSelection sets the cancel selection (0 = No, 1 = Yes)
func (vm *ViewManager) SetCancelSelection(selection int) {
	if selection == 0 || selection == 1 {
		vm.cancelSelection = selection
	}
}

// GetCancelSelection returns the current selection (0 = No, 1 = Yes)
func (vm *ViewManager) GetCancelSelection() int {
	return vm.cancelSelection
}

// IsCancelConfirmed returns true if "Yes" is selected
func (vm *ViewManager) IsCancelConfirmed() bool {
	return vm.cancelSelection == 1
}