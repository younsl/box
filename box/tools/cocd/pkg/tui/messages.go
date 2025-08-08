package tui

import (
	"time"
	
	"github.com/younsl/cocd/internal/monitor"
	"github.com/younsl/cocd/internal/scanner"
)

// Messages for Bubble Tea
type (
	jobsMsg               []scanner.JobStatus
	pendingJobsMsg        []scanner.JobStatus
	recentJobsMsg         []scanner.JobStatus
	errorMsg              string
	tickMsg               time.Time
	scanProgressMsg       struct{}
	updateUIMsg           struct{}
	cancelSuccessMsg      struct{}
	approvalSuccessMsg    struct{}
	recentJobUpdateMsg      monitor.JobUpdate
	jobUpdateMsg            monitor.JobUpdate
	startRecentStreamingMsg struct{}
)

