mod config;
mod scanner;
mod types;

use anyhow::Result;
use clap::Parser;
use config::Config;
use scanner::{export_to_csv, Scanner};
use std::time::Instant;
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() -> Result<()> {
    let config = Config::parse();

    // Setup logging
    let level = if config.verbose {
        Level::DEBUG
    } else {
        Level::INFO
    };

    let subscriber = FmtSubscriber::builder()
        .with_max_level(level)
        .with_target(false)
        .with_writer(std::io::stderr) // Send logs to stderr
        .finish();

    tracing::subscriber::set_global_default(subscriber)?;

    info!("Starting Java and Node.js version scanner");
    let start_time = Instant::now();

    // Extract output path before moving config
    let output_path = config.output.clone();

    // Setup signal handler for graceful shutdown
    let (shutdown_tx, mut shutdown_rx) = tokio::sync::watch::channel(false);

    tokio::spawn(async move {
        tokio::signal::ctrl_c()
            .await
            .expect("Failed to listen for ctrl-c");
        info!("Received shutdown signal, canceling operations...");
        let _ = shutdown_tx.send(true);
    });

    // Create scanner and run
    let scanner = Scanner::new(config, start_time);

    // Run scan with timeout
    let scan_timeout = tokio::time::Duration::from_secs(10 * 60);
    let result = tokio::select! {
        _ = shutdown_rx.changed() => {
            anyhow::bail!("Scan cancelled by user");
        }
        res = tokio::time::timeout(scan_timeout, scanner.scan_pods()) => {
            res??
        }
    };

    let elapsed = start_time.elapsed();

    // Print results to console
    scanner.print_results(&result, elapsed);

    // Export to CSV if output file is specified
    if let Some(path) = output_path {
        export_to_csv(&result, &path, elapsed)?;
    }

    Ok(())
}
