use anyhow::Result;
use globset::{Glob, GlobSet, GlobSetBuilder};
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::Duration;
use sysinfo::Disks;
use tokio::time;
use tracing::{debug, error, info, warn};

use crate::config::{Args, CleanupMode};

#[derive(Debug)]
pub struct FileInfo {
    path: PathBuf,
    size: u64,
}

pub struct Cleaner {
    config: Args,
    include_matcher: GlobSet,
    exclude_matcher: GlobSet,
    stopped: Arc<AtomicBool>,
}

impl Cleaner {
    pub fn new(config: Args) -> Result<Self> {
        let include_matcher = Self::build_matcher(&config.include_patterns)?;
        let exclude_matcher = Self::build_matcher(&config.exclude_patterns)?;

        Ok(Self {
            config,
            include_matcher,
            exclude_matcher,
            stopped: Arc::new(AtomicBool::new(false)),
        })
    }

    fn build_matcher(patterns: &[String]) -> Result<GlobSet> {
        let mut builder = GlobSetBuilder::new();
        for pattern in patterns {
            builder.add(Glob::new(pattern)?);
        }
        Ok(builder.build()?)
    }

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

    pub async fn stop(&self) {
        self.stopped.store(true, Ordering::Relaxed);
    }

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

    async fn clean_path(&self, base_path: &Path) {
        if !base_path.exists() {
            error!(path = %base_path.display(), "Path does not exist");
            return;
        }

        let initial_usage = self.get_disk_usage_percent(base_path);

        let files = self.collect_files(base_path);
        if files.is_empty() {
            info!(
                path = %base_path.display(),
                initial_usage_percent = initial_usage,
                "No files to clean"
            );
            return;
        }

        let total_size: u64 = files.iter().map(|f| f.size).sum();
        let total_size_mb = total_size / (1024 * 1024);

        info!(
            path = %base_path.display(),
            initial_usage_percent = initial_usage,
            file_count = files.len(),
            total_size_mb = total_size_mb,
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

            let file_size_kb = file.size / 1024;

            if self.config.dry_run {
                info!(
                    file = %file.path.display(),
                    size_kb = file_size_kb,
                    "[DRY-RUN] Would delete file"
                );
            } else {
                match fs::remove_file(&file.path) {
                    Ok(_) => {
                        info!(
                            file = %file.path.display(),
                            size_kb = file_size_kb,
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
        let freed_mb = freed_space / (1024 * 1024);

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
                freed_mb = freed_mb,
                "Cleanup completed successfully"
            );
        }
    }

    fn collect_files(&self, base_path: &Path) -> Vec<FileInfo> {
        let mut files = Vec::new();

        match self.walk_directory(base_path, &mut files) {
            Ok(_) => {}
            Err(e) => {
                error!(
                    path = %base_path.display(),
                    error = %e,
                    "Error walking directory"
                );
            }
        }

        files
    }

    fn walk_directory(&self, dir: &Path, files: &mut Vec<FileInfo>) -> Result<()> {
        if !dir.exists() {
            return Ok(());
        }

        let entries = match fs::read_dir(dir) {
            Ok(entries) => entries,
            Err(e) => {
                warn!(path = %dir.display(), error = %e, "Error reading directory");
                return Ok(());
            }
        };

        for entry in entries {
            let entry = match entry {
                Ok(e) => e,
                Err(e) => {
                    warn!(error = %e, "Error reading directory entry");
                    continue;
                }
            };

            let path = entry.path();
            let file_name = match path.file_name() {
                Some(name) => name.to_string_lossy().to_string(),
                None => continue,
            };

            let metadata = match entry.metadata() {
                Ok(m) => m,
                Err(e) => {
                    warn!(path = %path.display(), error = %e, "Error reading metadata");
                    continue;
                }
            };

            if metadata.is_dir() {
                if self.should_exclude(&file_name) {
                    debug!(dir = %path.display(), "Skipping excluded directory");
                    continue;
                }
                // Recursively walk subdirectory
                let _ = self.walk_directory(&path, files);
            } else {
                // Process file
                if self.should_exclude(&file_name) {
                    continue;
                }

                if !self.matches_include_pattern(&file_name) {
                    continue;
                }

                files.push(FileInfo {
                    path,
                    size: metadata.len(),
                });
            }
        }

        Ok(())
    }

    fn should_exclude(&self, name: &str) -> bool {
        self.exclude_matcher.is_match(name)
    }

    fn matches_include_pattern(&self, name: &str) -> bool {
        self.include_matcher.is_match(name)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs::File;
    use std::io::Write;
    use tempfile::TempDir;

    fn create_test_args() -> Args {
        Args {
            target_paths: vec![PathBuf::from("/tmp")],
            usage_threshold_percent: 80,
            check_interval_minutes: 10,
            include_patterns: vec!["*".to_string()],
            exclude_patterns: vec![".git".to_string(), "node_modules".to_string()],
            cleanup_mode: CleanupMode::Once,
            dry_run: true,
            log_level: "info".to_string(),
        }
    }

    #[test]
    fn test_cleaner_creation() {
        let args = create_test_args();
        let cleaner = Cleaner::new(args);
        assert!(cleaner.is_ok());
    }

    #[test]
    fn test_should_exclude() {
        let args = create_test_args();
        let cleaner = Cleaner::new(args).unwrap();

        assert!(cleaner.should_exclude(".git"));
        assert!(cleaner.should_exclude("node_modules"));
        assert!(!cleaner.should_exclude("test.txt"));
    }

    #[test]
    fn test_matches_include_pattern() {
        let args = create_test_args();
        let cleaner = Cleaner::new(args).unwrap();

        assert!(cleaner.matches_include_pattern("test.txt"));
        assert!(cleaner.matches_include_pattern("file.log"));
    }

    #[tokio::test]
    async fn test_collect_files() {
        let temp_dir = TempDir::new().unwrap();
        let temp_path = temp_dir.path();

        // Create test files
        File::create(temp_path.join("test1.txt"))
            .unwrap()
            .write_all(b"test")
            .unwrap();
        File::create(temp_path.join("test2.log"))
            .unwrap()
            .write_all(b"log")
            .unwrap();

        // Create excluded directory
        fs::create_dir(temp_path.join(".git")).unwrap();
        File::create(temp_path.join(".git/config"))
            .unwrap()
            .write_all(b"config")
            .unwrap();

        let mut args = create_test_args();
        args.target_paths = vec![temp_path.to_path_buf()];

        let cleaner = Cleaner::new(args).unwrap();
        let files = cleaner.collect_files(temp_path);

        // Should find test1.txt and test2.log, but not .git/config
        assert_eq!(files.len(), 2);
        assert!(files.iter().any(|f| f.path.ends_with("test1.txt")));
        assert!(files.iter().any(|f| f.path.ends_with("test2.log")));
        assert!(!files
            .iter()
            .any(|f| f.path.to_string_lossy().contains(".git")));
    }
}
