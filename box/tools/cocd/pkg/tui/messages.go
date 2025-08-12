package tui

import (
	"time"
	
	"github.com/younsl/cocd/pkg/monitor"
	"github.com/younsl/cocd/pkg/scanner"
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

