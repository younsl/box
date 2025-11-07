use anyhow::{Context, Result};
use aws_sdk_elasticache::Client as ElastiCacheClient;
use aws_sdk_elasticache::types::Snapshot;
use chrono::{FixedOffset, Utc};
use std::time::{Duration, Instant};
use tracing::{debug, info, warn};

use crate::error::BackupError;

/// Create an ElastiCache snapshot
pub async fn create_snapshot(client: &ElastiCacheClient, cache_cluster_id: &str) -> Result<String> {
    let snapshot_start_time = Instant::now();

    // Generate snapshot name with cluster ID and date
    // Use TZ environment variable to determine timezone offset (default: UTC+9 for Asia/Seoul)
    let tz_offset = std::env::var("TZ_OFFSET_HOURS")
        .ok()
        .and_then(|v| v.parse::<i32>().ok())
        .unwrap_or(9); // Default to UTC+9 (Asia/Seoul)

    let timezone =
        FixedOffset::east_opt(tz_offset * 3600).expect("Failed to create timezone offset");
    let date_str = Utc::now()
        .with_timezone(&timezone)
        .format("%Y%m%d")
        .to_string();
    let snapshot_name = format!("{}-{}", cache_cluster_id, date_str);

    info!(
        cache_cluster_id = %cache_cluster_id,
        snapshot_name = %snapshot_name,
        "Creating ElastiCache snapshot"
    );

    let response = client
        .create_snapshot()
        .cache_cluster_id(cache_cluster_id)
        .snapshot_name(&snapshot_name)
        .send()
        .await
        .context("Failed to create ElastiCache snapshot")?;

    if let Some(snapshot_info) = response.snapshot() {
        let creation_time = snapshot_start_time.elapsed().as_secs_f64();
        info!(
            duration_seconds = creation_time,
            snapshot_name = %snapshot_name,
            snapshot_arn = snapshot_info.arn().unwrap_or("N/A"),
            snapshot_status = ?snapshot_info.snapshot_status(),
            "Snapshot creation initiated successfully"
        );
    }

    Ok(snapshot_name)
}

/// Wait for snapshot to become available
pub async fn wait_for_completion(
    client: &ElastiCacheClient,
    snapshot_name: &str,
    max_wait_time: u64,
    check_interval: u64,
) -> Result<Snapshot> {
    let wait_start_time = Instant::now();
    let mut checks_performed = 0;

    info!(
        snapshot_name = %snapshot_name,
        max_wait_time_seconds = max_wait_time,
        check_interval_seconds = check_interval,
        "Waiting for snapshot completion"
    );

    loop {
        if wait_start_time.elapsed().as_secs() >= max_wait_time {
            return Err(BackupError::Timeout(format!(
                "Snapshot completion timeout after {:.1}s",
                wait_start_time.elapsed().as_secs_f64()
            ))
            .into());
        }

        let response = client
            .describe_snapshots()
            .snapshot_name(snapshot_name)
            .send()
            .await
            .context("Failed to describe snapshots")?;

        let snapshots = response.snapshots();
        if snapshots.is_empty() {
            return Err(
                BackupError::NotFound(format!("Snapshot {} not found", snapshot_name)).into(),
            );
        }

        let snapshot = &snapshots[0];
        let status = snapshot.snapshot_status().unwrap_or("Unknown");
        checks_performed += 1;
        let elapsed_time = wait_start_time.elapsed().as_secs_f64();

        debug!(
            check_number = checks_performed,
            status = %status,
            elapsed_seconds = elapsed_time,
            "Snapshot status check"
        );

        if status == "available" {
            let total_wait_time = wait_start_time.elapsed().as_secs_f64();
            info!(
                checks_performed,
                duration_seconds = total_wait_time,
                cache_node_type = snapshot.cache_node_type().unwrap_or("Unknown"),
                engine = snapshot.engine().unwrap_or("Unknown"),
                engine_version = snapshot.engine_version().unwrap_or("Unknown"),
                "Snapshot completed successfully"
            );

            return Ok(snapshot.clone());
        } else if status == "failed" {
            return Err(BackupError::SnapshotFailed(format!(
                "Snapshot creation failed with status: {} after {} checks",
                status, checks_performed
            ))
            .into());
        } else if status == "creating" {
            debug!("Snapshot creation in progress");
        }

        // Additional progress logging every 10 checks (~5 minutes)
        if checks_performed % 10 == 0 {
            info!(
                check_number = checks_performed,
                status = %status,
                elapsed_seconds = elapsed_time,
                "Long-running snapshot detected"
            );
        }

        tokio::time::sleep(Duration::from_secs(check_interval)).await;
    }
}

/// Delete a snapshot
pub async fn cleanup(client: &ElastiCacheClient, snapshot_name: &str) {
    // Skip cleanup if this is an export snapshot (has s3-export suffix)
    if snapshot_name.contains("-s3-export") {
        info!(
            snapshot_name = %snapshot_name,
            reason = "export snapshot",
            "Skipping cleanup"
        );
        return;
    }

    let cleanup_start_time = Instant::now();
    info!(
        snapshot_name = %snapshot_name,
        "Cleaning up source snapshot"
    );

    // Verify snapshot state before deletion
    match client
        .describe_snapshots()
        .snapshot_name(snapshot_name)
        .send()
        .await
    {
        Ok(response) => {
            if let Some(snapshot) = response.snapshots().first() {
                let status = snapshot.snapshot_status().unwrap_or("Unknown");

                if status != "available" && status != "failed" {
                    warn!(
                        snapshot_name = %snapshot_name,
                        status = %status,
                        "Snapshot is not in deletable state, skipping cleanup"
                    );
                    return;
                }

                debug!(
                    snapshot_name = %snapshot_name,
                    status = %status,
                    "Snapshot is in deletable state, proceeding with deletion"
                );
            } else {
                warn!(
                    snapshot_name = %snapshot_name,
                    "Snapshot not found for cleanup"
                );
                return;
            }
        }
        Err(e) => {
            warn!(
                snapshot_name = %snapshot_name,
                error = %e,
                "Could not verify snapshot state before cleanup"
            );
        }
    }

    match client
        .delete_snapshot()
        .snapshot_name(snapshot_name)
        .send()
        .await
    {
        Ok(_) => {
            let cleanup_time = cleanup_start_time.elapsed().as_secs_f64();
            info!(
                snapshot_name = %snapshot_name,
                duration_seconds = cleanup_time,
                "Source snapshot cleanup completed"
            );
        }
        Err(e) => {
            let cleanup_error_time = cleanup_start_time.elapsed().as_secs_f64();
            warn!(
                snapshot_name = %snapshot_name,
                duration_seconds = cleanup_error_time,
                error = %e,
                "Snapshot cleanup failed"
            );
        }
    }
}
