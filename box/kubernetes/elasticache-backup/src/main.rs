use anyhow::Result;
use clap::Parser;
use std::time::Instant;
use tracing::{error, info, info_span};

mod backup;
mod cli;
mod error;
mod export;
mod retention;
mod snapshot;
mod types;

use cli::Args;
use types::{ExecutionSummary, RetentionInfo, StepTimings};

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing with configurable format
    // Use JSON format if LOG_FORMAT=json, otherwise use pretty format
    let log_format = std::env::var("LOG_FORMAT").unwrap_or_else(|_| "pretty".to_string());
    let log_level = std::env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string());

    match log_format.to_lowercase().as_str() {
        "json" => {
            tracing_subscriber::fmt()
                .with_env_filter(tracing_subscriber::EnvFilter::new(&log_level))
                .json()
                .with_current_span(true)
                .with_span_list(true)
                .init();
        }
        _ => {
            tracing_subscriber::fmt()
                .with_env_filter(tracing_subscriber::EnvFilter::new(&log_level))
                .with_target(false)
                .with_thread_ids(false)
                .with_file(false)
                .compact()
                .init();
        }
    }

    let args = Args::parse();

    let _span = info_span!(
        "elasticache_backup",
        cache_cluster_id = %args.cache_cluster_id,
        s3_bucket_name = %args.s3_bucket_name,
        region = %args.region
    )
    .entered();

    info!(
        cache_cluster_id = %args.cache_cluster_id,
        s3_bucket_name = %args.s3_bucket_name,
        region = %args.region,
        "ElastiCache snapshot backup started"
    );

    let lambda_start_time = Instant::now();
    let mut step_timings = StepTimings::default();
    let mut snapshot_name: Option<String> = None;

    match backup::run(&args, &mut step_timings, &mut snapshot_name).await {
        Ok((target_snapshot, s3_location, deleted_count)) => {
            let total_time = lambda_start_time.elapsed().as_secs_f64();

            let retention_info = if args.retention_count > 0 {
                Some(RetentionInfo {
                    enabled: true,
                    retention_count: args.retention_count,
                    deleted_count,
                })
            } else {
                None
            };

            let summary = ExecutionSummary {
                status: "Success".to_string(),
                message: "ElastiCache snapshot backup completed successfully".to_string(),
                total_execution_time_seconds: total_time,
                step_timings,
                cache_cluster: args.cache_cluster_id.clone(),
                snapshot_name: snapshot_name.clone(),
                target_snapshot_name: Some(target_snapshot.clone()),
                s3_location: Some(s3_location.clone()),
                s3_bucket: args.s3_bucket_name.clone(),
                retention_info,
            };

            info!(
                snapshot_creation_seconds = summary.step_timings.snapshot_creation,
                snapshot_wait_seconds = summary.step_timings.snapshot_wait,
                s3_export_seconds = summary.step_timings.s3_export,
                export_wait_seconds = summary.step_timings.export_wait,
                cleanup_seconds = summary.step_timings.cleanup,
                retention_seconds = summary.step_timings.retention,
                total_execution_seconds = total_time,
                "Execution timing summary"
            );

            info!(
                status = "success",
                snapshot_name = snapshot_name.as_deref().unwrap_or(""),
                target_snapshot_name = %target_snapshot,
                s3_location = %s3_location,
                total_execution_seconds = total_time,
                "Backup execution completed successfully"
            );

            println!("{}", serde_json::to_string_pretty(&summary)?);
            Ok(())
        }
        Err(e) => {
            let total_time = lambda_start_time.elapsed().as_secs_f64();

            error!(
                snapshot_creation_seconds = step_timings.snapshot_creation,
                snapshot_wait_seconds = step_timings.snapshot_wait,
                s3_export_seconds = step_timings.s3_export,
                export_wait_seconds = step_timings.export_wait,
                cleanup_seconds = step_timings.cleanup,
                retention_seconds = step_timings.retention,
                total_execution_seconds = total_time,
                "Execution timing summary (error)"
            );

            error!(
                status = "failed",
                error = %e,
                snapshot_name = snapshot_name.as_deref().unwrap_or(""),
                total_execution_seconds = total_time,
                "Backup execution failed"
            );

            Err(e)
        }
    }
}
