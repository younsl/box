use clap::Parser;
use std::path::PathBuf;
use std::time::Duration;

const VERSION: &str = concat!(
    env!("CARGO_PKG_VERSION"),
    " (commit: ",
    env!("GIT_HASH"),
    ")"
);

#[derive(Parser, Debug, Clone)]
#[command(
    version = VERSION,
    about = "Scan Kubernetes pods for Java and Node.js versions",
    long_about = None,
    after_help = "Examples:
  # Scan default namespace:
  jvs -n default

  # Scan multiple namespaces:
  jvs -n production,staging,development

  # Export to CSV:
  jvs -n production -o results.csv

  # High concurrency scan:
  jvs -n production --max-concurrent 50

  # Filter pods with Java < 15 and Node < 20.3.0:
  jvs -n production --min-java-version 15 --min-node-version 20.3.0"
)]
pub struct Config {
    /// Comma-separated list of namespaces to scan (required)
    #[arg(short, long, required = true, value_delimiter = ',')]
    pub namespaces: Vec<String>,

    /// Maximum number of concurrent tasks
    #[arg(short = 'c', long, default_value_t = 20)]
    pub max_concurrent: usize,

    /// Timeout for kubectl commands in seconds
    #[arg(short, long, default_value_t = 30)]
    pub timeout: u64,

    /// Skip DaemonSet pods
    #[arg(short, long, default_value_t = true)]
    pub skip_daemonset: bool,

    /// Enable verbose logging
    #[arg(short, long)]
    pub verbose: bool,

    /// Output results to CSV file
    #[arg(short, long, value_name = "FILE")]
    pub output: Option<PathBuf>,

    /// Minimum Java version to filter (e.g., "15" or "11.0.16")
    /// Only show pods with Java version below this threshold
    #[arg(long, value_name = "VERSION")]
    pub min_java_version: Option<String>,

    /// Minimum Node.js version to filter (e.g., "20.3.0" or "18")
    /// Only show pods with Node.js version below this threshold
    #[arg(long, value_name = "VERSION")]
    pub min_node_version: Option<String>,
}

impl Config {
    pub fn timeout_duration(&self) -> Duration {
        Duration::from_secs(self.timeout)
    }
}
