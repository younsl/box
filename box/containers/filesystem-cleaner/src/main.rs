use anyhow::Result;
use clap::Parser;
use std::sync::Arc;
use tokio::signal;
use tracing::{error, info, warn};

mod cleaner;
mod config;
mod matcher;
mod scanner;

use cleaner::Cleaner;
use config::Args;

const VERSION: &str = env!("CARGO_PKG_VERSION");

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Setup logging
    setup_logging(&args.log_level)?;

    // Log startup with version info
    info!(
        version = VERSION,
        commit = option_env!("VERGEN_GIT_SHA").unwrap_or("unknown"),
        date = option_env!("VERGEN_BUILD_TIMESTAMP").unwrap_or("unknown"),
        "Starting filesystem-cleaner"
    );

    // Log configuration
    info!(
        target_paths = ?args.target_paths,
        usage_threshold_percent = args.usage_threshold_percent,
        cleanup_mode = ?args.cleanup_mode,
        include_patterns = ?args.include_patterns,
        exclude_patterns = ?args.exclude_patterns,
        dry_run = args.dry_run,
        log_level = args.log_level,
        check_interval_minutes = args.check_interval_minutes,
        "Configuration loaded"
    );

    if args.dry_run {
        warn!("Running in DRY-RUN mode - no files will be deleted");
    }

    let cleaner = Arc::new(Cleaner::new(args)?);
    let cleaner_clone = Arc::clone(&cleaner);

    // Setup signal handler
    tokio::spawn(async move {
        if let Err(e) = signal::ctrl_c().await {
            error!(error = %e, "Failed to listen for shutdown signal");
            return;
        }
        info!("Received shutdown signal, stopping cleaner...");
        cleaner_clone.stop().await;
    });

    // Run cleaner
    if let Err(e) = cleaner.run().await {
        error!(error = %e, "Failed to run cleaner");
        return Err(e);
    }

    Ok(())
}

fn setup_logging(level: &str) -> Result<()> {
    use tracing_subscriber::{EnvFilter, fmt};

    let filter = EnvFilter::try_new(level)
        .or_else(|_| EnvFilter::try_new("info"))
        .unwrap();

    // Human-readable compact format
    fmt()
        .with_env_filter(filter)
        .with_target(false)
        .with_thread_ids(false)
        .with_file(false)
        .with_line_number(false)
        .with_level(true)
        .with_ansi(true)
        .compact()
        .init();

    Ok(())
}
