use clap::Parser;
use std::path::PathBuf;

/// Build extended version information
fn build_version() -> &'static str {
    const VERSION: &str = env!("CARGO_PKG_VERSION");
    const COMMIT: &str = match option_env!("VERGEN_GIT_SHA") {
        Some(c) => c,
        None => "unknown",
    };
    const BUILD_DATE: &str = match option_env!("VERGEN_BUILD_TIMESTAMP") {
        Some(d) => d,
        None => "unknown",
    };

    // Use Box::leak to create a static string
    Box::leak(
        format!(
            "{}\nCommit: {}\nBuild Date: {}",
            VERSION, COMMIT, BUILD_DATE
        )
        .into_boxed_str(),
    )
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum CleanupMode {
    Once,
    Interval,
}

impl std::fmt::Display for CleanupMode {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CleanupMode::Once => write!(f, "once"),
            CleanupMode::Interval => write!(f, "interval"),
        }
    }
}

impl std::str::FromStr for CleanupMode {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "once" => Ok(CleanupMode::Once),
            "interval" => Ok(CleanupMode::Interval),
            _ => Err(format!("Invalid cleanup mode: {}", s)),
        }
    }
}

#[derive(Parser, Debug)]
#[command(name = "filesystem-cleaner")]
#[command(author, version, about, long_about = None)]
#[command(long_version = build_version())]
pub struct Args {
    /// Target filesystem paths to clean (comma-separated)
    #[arg(
        long = "target-paths",
        env = "TARGET_PATHS",
        default_value = "/home/runner/_work",
        value_delimiter = ',',
        help = "Target filesystem paths to clean (comma-separated)"
    )]
    pub target_paths: Vec<PathBuf>,

    /// Disk usage percentage threshold to trigger cleanup (0-100)
    #[arg(
        long = "usage-threshold-percent",
        env = "USAGE_THRESHOLD_PERCENT",
        default_value = "80",
        help = "Disk usage percentage threshold to trigger cleanup (0-100)"
    )]
    pub usage_threshold_percent: u8,

    /// Interval between cleanup checks in minutes (used with cleanup-mode=interval)
    #[arg(
        long = "check-interval-minutes",
        env = "CHECK_INTERVAL_MINUTES",
        default_value = "10",
        help = "Interval between cleanup checks in minutes"
    )]
    pub check_interval_minutes: u64,

    /// Glob patterns to include for deletion (comma-separated)
    /// Examples: *.tmp, **/cache/**, */build/*
    #[arg(
        long = "include-patterns",
        env = "INCLUDE_PATTERNS",
        default_value = "*",
        value_delimiter = ',',
        help = "Glob patterns to include for deletion (e.g., *.tmp, **/cache/**)"
    )]
    pub include_patterns: Vec<String>,

    /// Glob patterns to exclude from deletion (comma-separated)
    /// Examples: **/.git/**, **/node_modules/**, *.log
    #[arg(
        long = "exclude-patterns",
        env = "EXCLUDE_PATTERNS",
        default_value = "**/.git/**,**/node_modules/**,*.log",
        value_delimiter = ',',
        help = "Glob patterns to exclude from deletion (e.g., **/.git/**, **/node_modules/**)"
    )]
    pub exclude_patterns: Vec<String>,

    /// Cleanup mode: 'once' for single run (initContainer), 'interval' for periodic cleanup
    #[arg(
        long = "cleanup-mode",
        env = "CLEANUP_MODE",
        default_value = "interval",
        help = "Cleanup mode: 'once' or 'interval'"
    )]
    pub cleanup_mode: CleanupMode,

    /// Dry run mode (don't delete files)
    #[arg(
        long = "dry-run",
        env = "DRY_RUN",
        default_value = "false",
        help = "Dry run mode - no files will be deleted"
    )]
    pub dry_run: bool,

    /// Log level (trace, debug, info, warn, error)
    #[arg(
        long = "log-level",
        env = "LOG_LEVEL",
        default_value = "info",
        help = "Log level (trace, debug, info, warn, error)"
    )]
    pub log_level: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cleanup_mode_from_str() {
        assert_eq!("once".parse::<CleanupMode>().unwrap(), CleanupMode::Once);
        assert_eq!(
            "interval".parse::<CleanupMode>().unwrap(),
            CleanupMode::Interval
        );
        assert_eq!("ONCE".parse::<CleanupMode>().unwrap(), CleanupMode::Once);
        assert!("invalid".parse::<CleanupMode>().is_err());
    }

    #[test]
    fn test_cleanup_mode_display() {
        assert_eq!(CleanupMode::Once.to_string(), "once");
        assert_eq!(CleanupMode::Interval.to_string(), "interval");
    }
}
