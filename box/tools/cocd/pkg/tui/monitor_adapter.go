package tui

import (
	"context"
	"time"
	
	"github.com/younsl/cocd/internal/monitor"
	"github.com/younsl/cocd/internal/scanner"
)

// MonitorAdapter adapts the monitor.Monitor to the Monitor interface
type MonitorAdapter struct {
	monitor *monitor.Monitor
}

// NewMonitorAdapter creates a new monitor adapter
func NewMonitorAdapter(m *monitor.Monitor) Monitor {
	return &MonitorAdapter{
		monitor: m,
	}
}

// StartMonitoring starts the monitoring process
func (ma *MonitorAdapter) StartMonitoring(ctx context.Context, jobsChan chan []scanner.JobStatus) {
	ma.monitor.StartMonitoring(ctx, jobsChan)
}

// GetPendingJobs returns pending jobs
func (ma *MonitorAdapter) GetPendingJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return ma.monitor.GetPendingJobs(ctx)
}

// GetRecentJobs returns recent jobs
func (ma *MonitorAdapter) GetRecentJobs(ctx context.Context) ([]scanner.JobStatus, error) {
	return ma.monitor.GetRecentJobs(ctx)
}

// GetClient returns the GitHub client
func (ma *MonitorAdapter) GetClient() interface{} {
	return ma.monitor.GetClient()
}

// GetProgressTracker returns the progress tracker
func (ma *MonitorAdapter) GetProgressTracker() ProgressTracker {
	return &progressTrackerAdapter{
		tracker: ma.monitor.GetProgressTracker(),
	}
}

// GetScanProgress returns the scan progress
func (ma *MonitorAdapter) GetScanProgress() monitor.ScanProgress {
	return ma.monitor.GetScanProgress()
}

// GetUpdateInterval returns the update interval
func (ma *MonitorAdapter) GetUpdateInterval() int {
	return ma.monitor.GetUpdateInterval()
}

// GetPendingJobsWithStreaming gets pending jobs with real-time streaming
func (ma *MonitorAdapter) GetPendingJobsWithStreaming(ctx context.Context, jobUpdateChan chan<- monitor.JobUpdate) error {
	return ma.monitor.GetPendingJobsWithStreaming(ctx, jobUpdateChan)
}

// GetRecentJobsWithStreaming gets recent jobs with real-time streaming
func (ma *MonitorAdapter) GetRecentJobsWithStreaming(ctx context.Context, jobUpdateChan chan<- monitor.JobUpdate) error {
	return ma.monitor.GetRecentJobsWithStreaming(ctx, jobUpdateChan)
}

// progressTrackerAdapter adapts monitor.ProgressTracker to ProgressTracker interface
type progressTrackerAdapter struct {
	tracker *monitor.ProgressTracker
}

func (pta *progressTrackerAdapter) UpdateScanCountdown() {
	pta.tracker.UpdateScanCountdown()
}

func (pta *progressTrackerAdapter) SetNextScanTimer(nextScanAt time.Time, scanCount int, isFull bool) {
	pta.tracker.SetNextScanTimer(nextScanAt, scanCount, isFull)
}

// Ensure MonitorAdapter implements Monitor interface
var _ Monitor = (*MonitorAdapter)(nil)