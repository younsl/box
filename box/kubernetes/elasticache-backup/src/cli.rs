use clap::Parser;

/// ElastiCache snapshot backup to S3 automation
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Args {
    /// ElastiCache cluster ID (read replica node)
    #[arg(long, env = "CACHE_CLUSTER_ID")]
    pub cache_cluster_id: String,

    /// S3 bucket name for storing RDB files
    #[arg(long, env = "S3_BUCKET_NAME")]
    pub s3_bucket_name: String,

    /// AWS region
    #[arg(long, env = "AWS_REGION", default_value = "ap-northeast-2")]
    pub region: String,

    /// Maximum wait time for snapshot completion in seconds
    #[arg(long, default_value = "1800")]
    pub snapshot_timeout: u64,

    /// Maximum wait time for S3 export completion in seconds
    #[arg(long, default_value = "300")]
    pub export_timeout: u64,

    /// Snapshot status check interval in seconds
    #[arg(long, default_value = "30")]
    pub check_interval: u64,

    /// Number of snapshots to retain in S3 (0 = unlimited)
    #[arg(long, env = "RETENTION_COUNT", default_value = "0")]
    pub retention_count: u32,
}
