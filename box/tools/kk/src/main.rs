mod checker;
mod config;

use anyhow::Result;
use clap::Parser;
use log::error;
use std::path::PathBuf;

const VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Parser)]
#[command(
    name = "kk",
    version = VERSION,
    about = "kk checks domain configurations",
    long_about = "kk validates domain configurations based on a provided YAML file. \
                  It checks various aspects of the domain setup to ensure correctness and adherence to standards."
)]
struct Cli {
    /// Path to the YAML configuration file (e.g., configs/domains.yaml)
    #[arg(short, long, value_name = "FILE")]
    config: PathBuf,

    /// Enable verbose logging
    #[arg(short, long)]
    verbose: bool,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    // Initialize logger
    let log_level = if cli.verbose { "debug" } else { "info" };
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    // Load configuration
    let config = match config::Config::load(&cli.config) {
        Ok(cfg) => cfg,
        Err(e) => {
            error!("Failed to load config '{}': {}", cli.config.display(), e);
            return Err(e);
        }
    };

    println!("Loaded domain list from '{}'.", cli.config.display());

    // Run checks
    checker::run_checks(config.domains).await?;

    Ok(())
}
