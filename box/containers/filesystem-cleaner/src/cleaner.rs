use anyhow::Result;
use bytesize::ByteSize;
use std::fs;
use std::path::Path;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Duration;
use sysinfo::Disks;
use tokio::time;
use tracing::{error, info, warn};

use crate::config::{Args, CleanupMode};
use crate::matcher::PatternMatcher;
use crate::scanner::FileScanner;

/// Filesystem cleaner orchestrator
///
/// Responsible for:
/// - Monitoring disk usage
/// - Scheduling cleanup operations (once or interval)
/// - Coordinating file scanning and deletion
/// - Logging cleanup results
pub struct Cleaner {
    config: Args,
    matcher: PatternMatcher,
    stopped: Arc<AtomicBool>,
}

impl Cleaner {
    /// Create a new cleaner with the given configuration
    pub fn new(config: Args) -> Result<Self> {
        let matcher = PatternMatcher::new(&config.include_patterns, &config.exclude_patterns)?;

        Ok(Self {
            config,
            matcher,
            stopped: Arc::new(AtomicBool::new(false)),
        })
    }

    /// Run the cleaner based on configured mode (once or interval)
    pub async fn run(&self) -> Result<()> {
        match self.config.cleanup_mode {
            CleanupMode::Once => {
                info!("Running in 'once' mode - single cleanup execution");
                self.perform_cleanup().await;
                info!("Cleanup completed, exiting");
                Ok(())
            }
            CleanupMode::Interval => {
                info!(
                    interval_minutes = self.config.check_interval_minutes,
                    "Running in 'interval' mode - periodic cleanup"
                );

                // Run initial cleanup
                self.perform_cleanup().await;

                // Run periodic cleanup
                let mut interval =
                    time::interval(Duration::from_secs(self.config.check_interval_minutes * 60));

                loop {
                    interval.tick().await;

                    if self.stopped.load(Ordering::Relaxed) {
                        info!("Cleaner stopped");
                        break;
                    }

                    self.perform_cleanup().await;
                }

                Ok(())
            }
        }
    }

    /// Stop the cleaner (for graceful shutdown)
    pub async fn stop(&self) {
        self.stopped.store(true, Ordering::Relaxed);
    }

    /// Perform a single cleanup cycle
    async fn perform_cleanup(&self) {
        info!("Starting cleanup cycle");
        let start_time = std::time::Instant::now();

        for path in &self.config.target_paths {
            let usage = self.get_disk_usage_percent(path);

            if usage > self.config.usage_threshold_percent as f64 {
                warn!(
                    path = %path.display(),
                    usage = usage,
                    threshold = self.config.usage_threshold_percent,
                    cleanup_mode = %self.config.cleanup_mode,
                    dry_run = self.config.dry_run,
                    "Disk usage exceeds threshold, starting cleanup"
                );
                self.clean_path(path).await;
            } else {
                info!(
                    path = %path.display(),
                    usage = usage,
                    threshold = self.config.usage_threshold_percent,
                    cleanup_mode = %self.config.cleanup_mode,
                    "Disk usage is below threshold, skipping cleanup"
                );
            }
        }

        info!(
            duration_secs = start_time.elapsed().as_secs(),
            "Cleanup cycle completed"
        );
    }

    /// Get disk usage percentage for a given path
    fn get_disk_usage_percent(&self, path: &Path) -> f64 {
        let disks = Disks::new_with_refreshed_list();

        // Find the disk that contains this path
        let mut best_match: Option<&sysinfo::Disk> = None;
        let mut best_match_len = 0;

        for disk in &disks {
            let mount_point = disk.mount_point();
            if path.starts_with(mount_point) {
                let mount_len = mount_point.as_os_str().len();
                if mount_len > best_match_len {
                    best_match = Some(disk);
                    best_match_len = mount_len;
                }
            }
        }

        if let Some(disk) = best_match {
            let total = disk.total_space();
            let available = disk.available_space();

            if total == 0 {
                return 0.0;
            }

            let used = total - available;
            (used as f64 / total as f64) * 100.0
        } else {
            error!(path = %path.display(), "Failed to get disk usage - no matching disk found");
            0.0
        }
    }

    /// Clean files in the given path
    async fn clean_path(&self, base_path: &Path) {
        if !base_path.exists() {
            error!(path = %base_path.display(), "Path does not exist");
            return;
        }

        let initial_usage = self.get_disk_usage_percent(base_path);

        // Use FileScanner to collect files
        let scanner = FileScanner::new(&self.matcher);
        let files = scanner.scan(base_path);

        if files.is_empty() {
            info!(
                path = %base_path.display(),
                initial_usage_percent = initial_usage,
                "No files to clean"
            );
            return;
        }

        let total_size: u64 = files.iter().map(|f| f.size).sum();

        info!(
            path = %base_path.display(),
            initial_usage_percent = initial_usage,
            file_count = files.len(),
            total_size = %ByteSize::b(total_size),
            "Starting cleanup operation"
        );

        let mut deleted_count = 0;
        let mut freed_space = 0u64;
        let file_count = files.len();

        for file in &files {
            if self.stopped.load(Ordering::Relaxed) {
                info!("Cleanup interrupted by shutdown");
                break;
            }

            if self.config.dry_run {
                info!(
                    file = %file.path.display(),
                    size = %ByteSize::b(file.size),
                    "[DRY-RUN] Would delete file"
                );
            } else {
                match fs::remove_file(&file.path) {
                    Ok(_) => {
                        info!(
                            file = %file.path.display(),
                            size = %ByteSize::b(file.size),
                            "File deleted successfully"
                        );
                        deleted_count += 1;
                        freed_space += file.size;
                    }
                    Err(e) => {
                        error!(
                            file = %file.path.display(),
                            error = %e,
                            "Failed to delete file"
                        );
                    }
                }
            }
        }

        let final_usage = self.get_disk_usage_percent(base_path);
        let usage_reduction = initial_usage - final_usage;

        if self.config.dry_run {
            info!(
                path = %base_path.display(),
                initial_usage_percent = initial_usage,
                final_usage_percent = final_usage,
                usage_reduction = usage_reduction,
                would_delete = file_count,
                "Cleanup completed (DRY-RUN)"
            );
        } else {
            info!(
                path = %base_path.display(),
                initial_usage_percent = initial_usage,
                final_usage_percent = final_usage,
                usage_reduction = usage_reduction,
                deleted_count = deleted_count,
                freed_space = %ByteSize::b(freed_space),
                "Cleanup completed successfully"
            );
        }
    }
}
