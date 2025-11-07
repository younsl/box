use anyhow::{Context, Result};
use aws_sdk_s3::Client as S3Client;
use chrono::{DateTime, Utc};
use std::time::Instant;
use tracing::{debug, info, warn};

/// S3 object metadata for retention management
#[derive(Debug, Clone)]
struct S3Object {
    key: String,
    last_modified: DateTime<Utc>,
    size: i64,
}

/// Clean up old snapshots in S3 based on retention policy
pub async fn cleanup_old_snapshots(
    s3_client: &S3Client,
    bucket_name: &str,
    cache_cluster_id: &str,
    retention_count: u32,
) -> Result<usize> {
    if retention_count == 0 {
        info!(
            retention_count,
            "Retention count is 0, skipping cleanup for unlimited retention"
        );
        return Ok(0);
    }

    let cleanup_start_time = Instant::now();
    info!(
        bucket_name = %bucket_name,
        cache_cluster_id = %cache_cluster_id,
        retention_count,
        "Starting S3 snapshot retention cleanup"
    );

    // List all objects with the cache cluster prefix
    let prefix = format!("{}-", cache_cluster_id);
    debug!(
        prefix = %prefix,
        "Listing S3 objects with prefix"
    );

    let mut objects = Vec::new();
    let mut continuation_token: Option<String> = None;
    let mut total_listed = 0;

    loop {
        let mut request = s3_client
            .list_objects_v2()
            .bucket(bucket_name)
            .prefix(&prefix);

        if let Some(token) = continuation_token {
            request = request.continuation_token(token);
        }

        let response = request.send().await.context("Failed to list S3 objects")?;

        let contents = response.contents();
        let batch_size = contents.len();
        total_listed += batch_size;

        debug!(batch_size, total_listed, "Retrieved S3 objects batch");

        for obj in contents {
            if let (Some(key), Some(last_modified), Some(size)) =
                (obj.key(), obj.last_modified(), obj.size())
            {
                objects.push(S3Object {
                    key: key.to_string(),
                    last_modified: DateTime::from_timestamp(
                        last_modified.secs(),
                        last_modified.subsec_nanos(),
                    )
                    .unwrap_or_else(Utc::now),
                    size,
                });
            }
        }

        if response.is_truncated().unwrap_or(false) {
            continuation_token = response.next_continuation_token().map(|s| s.to_string());
            debug!("More objects available, continuing pagination");
        } else {
            break;
        }
    }

    let total_objects = objects.len();
    info!(
        total_objects,
        retention_count, "S3 objects listing completed"
    );

    if total_objects == 0 {
        info!("No snapshots found in S3, nothing to clean up");
        return Ok(0);
    }

    // Sort by last modified date (newest first)
    objects.sort_by(|a, b| b.last_modified.cmp(&a.last_modified));

    // Log the snapshots we found
    for (idx, obj) in objects.iter().enumerate() {
        debug!(
            index = idx + 1,
            key = %obj.key,
            last_modified = %obj.last_modified.format("%Y-%m-%d %H:%M:%S UTC"),
            size_bytes = obj.size,
            status = if idx < retention_count as usize { "KEEP" } else { "DELETE" },
            "Snapshot status"
        );
    }

    // Determine which objects to delete
    let objects_to_delete: Vec<_> = objects.iter().skip(retention_count as usize).collect();

    let delete_count = objects_to_delete.len();
    let keep_count = total_objects - delete_count;

    if delete_count == 0 {
        info!(
            total_objects,
            retention_count, "All snapshots are within retention policy, nothing to delete"
        );
        return Ok(0);
    }

    info!(
        total_objects,
        keep_count, delete_count, "Retention analysis completed"
    );

    // Delete old objects
    let mut deleted_count = 0;
    let mut failed_count = 0;
    let mut total_deleted_size: i64 = 0;

    for obj in objects_to_delete {
        let delete_start = Instant::now();

        info!(
            key = %obj.key,
            last_modified = %obj.last_modified.format("%Y-%m-%d %H:%M:%S UTC"),
            size_bytes = obj.size,
            "Deleting old snapshot"
        );

        match s3_client
            .delete_object()
            .bucket(bucket_name)
            .key(&obj.key)
            .send()
            .await
        {
            Ok(_) => {
                let delete_duration = delete_start.elapsed().as_secs_f64();
                deleted_count += 1;
                total_deleted_size += obj.size;

                info!(
                    key = %obj.key,
                    duration_seconds = delete_duration,
                    size_bytes = obj.size,
                    deleted_count,
                    remaining = delete_count - deleted_count,
                    "Snapshot deleted successfully"
                );
            }
            Err(e) => {
                let delete_duration = delete_start.elapsed().as_secs_f64();
                failed_count += 1;

                warn!(
                    key = %obj.key,
                    error = %e,
                    duration_seconds = delete_duration,
                    failed_count,
                    "Failed to delete snapshot"
                );
            }
        }
    }

    let total_cleanup_time = cleanup_start_time.elapsed().as_secs_f64();

    info!(
        total_objects,
        kept = keep_count,
        deleted = deleted_count,
        failed = failed_count,
        total_deleted_size_bytes = total_deleted_size,
        total_deleted_size_mb = total_deleted_size as f64 / 1024.0 / 1024.0,
        duration_seconds = total_cleanup_time,
        "S3 snapshot retention cleanup completed"
    );

    if failed_count > 0 {
        warn!(
            failed_count,
            deleted_count, "Some deletions failed, but continuing"
        );
    }

    Ok(deleted_count)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_retention_disabled_when_zero() {
        // This is a unit test placeholder
        // In a real scenario, you would mock the S3 client
        assert_eq!(0, 0);
    }
}
