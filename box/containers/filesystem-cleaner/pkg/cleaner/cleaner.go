package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/younsl/box/containers/filesystem-cleaner/pkg/config"
)

type Cleaner struct {
	config config.Config
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type fileInfo struct {
	path    string
	size    int64
	modTime time.Time
}

func New(cfg config.Config) *Cleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Cleaner{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Cleaner) Run() error {
	// Run cleanup based on mode
	if c.config.CleanupMode == "once" {
		logrus.Info("Running in 'once' mode - single cleanup execution")
		c.performCleanup()
		c.wg.Wait()
		logrus.Info("Cleanup completed, exiting")
		return nil
	}

	// Interval mode - run periodically
	logrus.WithField("interval", c.config.CheckInterval).Info("Running in 'interval' mode - periodic cleanup")
	
	ticker := time.NewTicker(c.config.CheckInterval)
	defer ticker.Stop()

	// Run initial cleanup
	c.performCleanup()

	for {
		select {
		case <-c.ctx.Done():
			logrus.Info("Cleaner stopped")
			c.wg.Wait()
			return nil
		case <-ticker.C:
			c.performCleanup()
		}
	}
}

func (c *Cleaner) Stop() {
	c.cancel()
}

func (c *Cleaner) performCleanup() {
	c.wg.Add(1)
	defer c.wg.Done()

	logrus.Info("Starting cleanup cycle")
	startTime := time.Now()

	for _, path := range c.config.CleanPaths {
		usage := c.getDiskUsagePercent(path)
		
		logger := logrus.WithFields(logrus.Fields{
			"path":  path,
			"usage": usage,
		})

		if usage > float64(c.config.ThresholdPercent) {
			logger.WithFields(logrus.Fields{
				"threshold":     c.config.ThresholdPercent,
				"cleanup_mode":  c.config.CleanupMode,
				"dry_run":       c.config.DryRun,
			}).Warn("Disk usage exceeds threshold, starting cleanup")
			c.cleanPath(path)
		} else {
			logger.WithFields(logrus.Fields{
				"threshold":     c.config.ThresholdPercent,
				"cleanup_mode":  c.config.CleanupMode,
			}).Info("Disk usage is below threshold, skipping cleanup")
		}
	}

	logrus.WithField("duration", time.Since(startTime)).Info("Cleanup cycle completed")
}

func (c *Cleaner) getDiskUsagePercent(path string) float64 {
	var stat syscall.Statfs_t
	
	if err := syscall.Statfs(path, &stat); err != nil {
		logrus.WithError(err).WithField("path", path).
			Error("Failed to get filesystem stats")
		return 0
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	if total == 0 {
		return 0
	}

	return float64(used) / float64(total) * 100
}

func (c *Cleaner) cleanPath(basePath string) {
	logger := logrus.WithField("path", basePath)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		logger.Error("Path does not exist")
		return
	}

	// Get initial usage before cleanup
	initialUsage := c.getDiskUsagePercent(basePath)

	files := c.collectFiles(basePath)
	if len(files) == 0 {
		logger.WithField("initial_usage_percent", initialUsage).
			Info("No files to clean")
		return
	}

	totalSize := int64(0)
	for _, f := range files {
		totalSize += f.size
	}

	logger.WithFields(logrus.Fields{
		"initial_usage_percent": initialUsage,
		"file_count":           len(files),
		"total_size_mb":        totalSize / (1024 * 1024),
	}).Info("Starting cleanup operation")

	deletedCount := 0
	freedSpace := int64(0)

	for _, file := range files {
		if c.ctx.Err() != nil {
			logger.Info("Cleanup interrupted by shutdown")
			break
		}

		fileLogger := logger.WithFields(logrus.Fields{
			"file":    file.path,
			"size_kb": file.size / 1024,
		})

		if c.config.DryRun {
			fileLogger.Info("[DRY-RUN] Would delete file")
		} else {
			if err := os.Remove(file.path); err != nil {
				fileLogger.WithError(err).Error("Failed to delete file")
			} else {
				fileLogger.Info("File deleted successfully")
				deletedCount++
				freedSpace += file.size
			}
		}

	}

	finalUsage := c.getDiskUsagePercent(basePath)
	usageReduction := initialUsage - finalUsage
	
	resultLogger := logger.WithFields(logrus.Fields{
		"initial_usage_percent": initialUsage,
		"final_usage_percent":   finalUsage,
		"usage_reduction":       usageReduction,
		"deleted_count":         deletedCount,
		"freed_mb":              freedSpace / (1024 * 1024),
	})

	if c.config.DryRun {
		resultLogger.WithField("would_delete", len(files)).
			Info("Cleanup completed (DRY-RUN)")
	} else {
		resultLogger.Info("Cleanup completed successfully")
	}
}

func (c *Cleaner) collectFiles(basePath string) []fileInfo {
	var files []fileInfo

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).WithField("path", path).
				Warn("Error accessing path")
			return nil
		}

		if info.IsDir() {
			if c.shouldExclude(info.Name()) {
				logrus.WithField("dir", path).Debug("Skipping excluded directory")
				return filepath.SkipDir
			}
			return nil
		}

		if c.shouldExclude(info.Name()) {
			return nil
		}

		if !c.matchesPattern(info.Name()) {
			return nil
		}

		// Collect all matching files (no age filter)
		files = append(files, fileInfo{
			path:    path,
			size:    info.Size(),
			modTime: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		logrus.WithError(err).WithField("path", basePath).
			Error("Error walking path")
	}

	return files
}

func (c *Cleaner) shouldExclude(name string) bool {
	for _, pattern := range c.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func (c *Cleaner) matchesPattern(name string) bool {
	for _, pattern := range c.config.FilePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}