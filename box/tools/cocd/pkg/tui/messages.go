package tui

import (
	"time"
	
	"github.com/younsl/cocd/internal/monitor"
	"github.com/younsl/cocd/internal/scanner"
)

// Messages for Bubble Tea
type (
	jobsMsg         []scanner.JobStatus
	recentJobsMsg   []scanner.JobStatus
	errorMsg        string
	tickMsg         time.Time
	scanProgressMsg struct{}
	updateUIMsg        struct{} // For forcing UI updates
	cancelSuccessMsg struct{} // For successful cancellation
	approvalSuccessMsg struct{} // For successful approval
	jobUpdateMsg    monitor.JobUpdate // For streaming job updates
)

