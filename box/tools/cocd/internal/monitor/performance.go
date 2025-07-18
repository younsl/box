package monitor

import (
	"context"
	"time"
)

// PerformanceOptimizer handles adaptive performance optimizations for GHES
type PerformanceOptimizer struct {
	responseTimeHistory []time.Duration
	avgResponseTime     time.Duration
	maxHistorySize      int
	slowResponseThreshold time.Duration
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		responseTimeHistory:   make([]time.Duration, 0, 10),
		maxHistorySize:        10,
		slowResponseThreshold: 2 * time.Second,
	}
}

// RecordResponseTime records API response time for adaptive optimization
func (po *PerformanceOptimizer) RecordResponseTime(duration time.Duration) {
	po.responseTimeHistory = append(po.responseTimeHistory, duration)
	
	// Keep only recent history
	if len(po.responseTimeHistory) > po.maxHistorySize {
		po.responseTimeHistory = po.responseTimeHistory[1:]
	}
	
	// Calculate average
	var total time.Duration
	for _, d := range po.responseTimeHistory {
		total += d
	}
	po.avgResponseTime = total / time.Duration(len(po.responseTimeHistory))
}

// GetOptimalDelay returns optimal delay based on server response times
func (po *PerformanceOptimizer) GetOptimalDelay(baseDelay time.Duration) time.Duration {
	if po.avgResponseTime > po.slowResponseThreshold {
		// Server is slow, increase delay significantly
		return baseDelay * 3
	} else if po.avgResponseTime > time.Second {
		// Server is moderately slow, increase delay
		return baseDelay * 2
	}
	return baseDelay
}

// IsServerUnderLoad checks if server appears to be under load
func (po *PerformanceOptimizer) IsServerUnderLoad() bool {
	return po.avgResponseTime > po.slowResponseThreshold
}

// WithPerformanceTracking wraps API calls with performance tracking
func (po *PerformanceOptimizer) WithPerformanceTracking(ctx context.Context, fn func(context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)
	
	po.RecordResponseTime(duration)
	
	// If server is under load, add adaptive delay
	if po.IsServerUnderLoad() {
		adaptiveDelay := po.GetOptimalDelay(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(adaptiveDelay):
		}
	}
	
	return err
}