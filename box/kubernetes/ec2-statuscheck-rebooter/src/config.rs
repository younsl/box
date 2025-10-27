use clap::Parser;

#[derive(Parser, Debug, Clone)]
#[command(
    name = "ec2-statuscheck-rebooter",
    version,
    about = "Automated EC2 instance reboot on status check failures"
)]
pub struct Config {
    /// Check interval in seconds
    #[arg(long, env = "CHECK_INTERVAL_SECONDS", default_value = "300")]
    pub check_interval_seconds: u64,

    /// Instance status check failure threshold before reboot
    #[arg(long, env = "FAILURE_THRESHOLD", default_value = "2")]
    pub failure_threshold: u32,

    /// AWS region
    #[arg(long, env = "AWS_REGION")]
    pub region: Option<String>,

    /// Comma-separated tag filters for EC2 instances (format: Key=Value,Key=Value)
    #[arg(long, env = "TAG_FILTERS", value_delimiter = ',')]
    pub tag_filters: Vec<String>,

    /// Dry run mode (no actual reboot)
    #[arg(long, env = "DRY_RUN", default_value = "false")]
    pub dry_run: bool,

    /// Log format: json or pretty
    #[arg(long, env = "LOG_FORMAT", default_value = "json")]
    pub log_format: String,

    /// Log level
    #[arg(long, env = "LOG_LEVEL", default_value = "info")]
    pub log_level: String,
}

impl Config {
    pub fn from_args() -> Self {
        Self::parse()
    }

    pub fn display(&self, actual_region: &str) {
        let region_info = if let Some(region) = &self.region {
            region.clone()
        } else {
            format!("auto-detect ({})", actual_region)
        };

        let tag_filter_info = if self.tag_filters.is_empty() {
            "NONE (monitoring ALL instances)".to_string()
        } else {
            format!("{:?}", self.tag_filters)
        };

        tracing::info!(
            check_interval_seconds = self.check_interval_seconds,
            failure_threshold = self.failure_threshold,
            dry_run = self.dry_run,
            region = %region_info,
            tag_filters = %tag_filter_info,
            log_format = %self.log_format,
            log_level = %self.log_level,
            "Configuration initialized"
        );

        if self.dry_run {
            tracing::warn!("DRY RUN MODE ENABLED - No instances will be rebooted, only logged");
        }

        if self.tag_filters.is_empty() {
            tracing::warn!(
                region = actual_region,
                "WARNING: No tag filters configured - monitoring ALL running instances in this region"
            );
        }
    }
}
