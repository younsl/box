import boto3
import os
import time
import logging
from datetime import datetime

# Configure logging - Global INFO level
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# ElastiCache Configuration (Cluster Mode Disabled - Use Read Replica)
CACHE_CLUSTER_ID = os.getenv("CACHE_CLUSTER_ID", "<YOUR_ELASTICACHE_NODE_NAME>")  # Read Replica node
REPLICATION_GROUP_ID = None  # Not used for cluster-mode disabled

# S3 Configuration
S3_BUCKET_NAME = os.getenv("S3_BUCKET_NAME", "<S3_BUCKET_NAME_TO_STORE_RDB_FILES")

logger.info(f"Configuration loaded - CacheCluster: {CACHE_CLUSTER_ID}, S3 Bucket: {S3_BUCKET_NAME}")

def create_elasticache_snapshot():
    """ Create ElastiCache snapshot using AWS API """
    try:
        logger.info("Starting ElastiCache snapshot creation")
        snapshot_start_time = time.time()
        
        elasticache_client = boto3.client('elasticache', region_name='ap-northeast-2')
        
        # Generate snapshot name with node name and date
        date_str = datetime.now().strftime('%Y%m%d')
        snapshot_name = f"{CACHE_CLUSTER_ID}-{date_str}"
        
        logger.info(f"Creating snapshot: {snapshot_name}")
        
        # Create snapshot using Cache Cluster ID (Read Replica)
        if CACHE_CLUSTER_ID:
            logger.info(f"Creating snapshot with CacheClusterId (Read Replica): {CACHE_CLUSTER_ID}")
            response = elasticache_client.create_snapshot(
                CacheClusterId=CACHE_CLUSTER_ID,
                SnapshotName=snapshot_name
            )
            snapshot_info = response['Snapshot']
            logger.info("Snapshot created successfully using Read Replica")
        else:
            raise Exception("CACHE_CLUSTER_ID not specified")
        
        creation_time = time.time() - snapshot_start_time
        logger.info(f"Snapshot creation initiated successfully in {creation_time:.2f}s")
        logger.info(f"Snapshot ARN: {snapshot_info.get('SnapshotArn', 'N/A')}")
        logger.info(f"Snapshot Status: {snapshot_info.get('SnapshotStatus', 'Unknown')}")
        
        return snapshot_name, snapshot_info
        
    except Exception as e:
        snapshot_error_time = time.time() - snapshot_start_time if 'snapshot_start_time' in locals() else 0
        logger.error(f"Snapshot creation failed after {snapshot_error_time:.2f}s: {str(e)}")
        raise

def wait_for_snapshot_completion(snapshot_name, max_wait_time=1800):
    """ Wait for snapshot to complete (max 30 minutes) """
    try:
        logger.info(f"Waiting for snapshot completion: {snapshot_name}")
        wait_start_time = time.time()
        
        elasticache_client = boto3.client('elasticache', region_name='ap-northeast-2')
        check_interval = 30  # Check every 30 seconds
        checks_performed = 0
        
        while time.time() - wait_start_time < max_wait_time:
            try:
                response = elasticache_client.describe_snapshots(SnapshotName=snapshot_name)
                snapshots = response.get('Snapshots', [])
                
                if not snapshots:
                    raise Exception(f"Snapshot {snapshot_name} not found")
                
                snapshot = snapshots[0]
                status = snapshot.get('SnapshotStatus', 'Unknown')
                checks_performed += 1
                elapsed_time = time.time() - wait_start_time
                
                # Get additional snapshot progress info
                progress_percent = snapshot.get('PercentProgress', 'Unknown')
                
                # Detailed status logging for each check
                logger.info(f"Check #{checks_performed}: Status='{status}', Progress={progress_percent}%, Elapsed={elapsed_time:.1f}s")
                
                if status == 'available':
                    total_wait_time = time.time() - wait_start_time
                    logger.info(f"Snapshot completed successfully after {checks_performed} checks in {total_wait_time:.1f}s")
                    
                    # Log snapshot details with size information
                    node_type = snapshot.get('NodeType', 'Unknown')
                    engine = snapshot.get('Engine', 'Unknown')
                    engine_version = snapshot.get('EngineVersion', 'Unknown')
                    snapshot_size = snapshot.get('SnapshotSizeInBytes', 0)
                    snapshot_size_mb = snapshot_size / 1024 / 1024 if snapshot_size > 0 else 0
                    
                    logger.info(f"Snapshot Details - Type: {node_type}, Engine: {engine} {engine_version}")
                    logger.info(f"Snapshot Size: {snapshot_size:,} bytes ({snapshot_size_mb:.1f} MB)")
                    logger.info(f"Snapshot ready for S3 export - Size: {snapshot_size_mb:.1f} MB")
                    
                    return snapshot
                    
                elif status == 'failed':
                    raise Exception(f"Snapshot creation failed with status: {status} after {checks_performed} checks")
                elif status == 'creating':
                    logger.info(f"Snapshot creation in progress... ({progress_percent}% complete)")
                else:
                    logger.info(f"Snapshot status: {status}")
                
                # Additional progress logging every 5 minutes
                if checks_performed % 10 == 0:
                    logger.info(f"Long-running snapshot detected - Check #{checks_performed}, Status: {status}, {elapsed_time:.1f}s elapsed")
                
                time.sleep(check_interval)
                
            except Exception as e:
                if "SnapshotNotFoundFault" in str(e):
                    logger.warning(f"Snapshot not yet visible, continuing to wait...")
                    time.sleep(check_interval)
                    continue
                else:
                    raise
        
        # Timeout reached
        total_wait_time = time.time() - wait_start_time
        logger.error(f"Snapshot completion timeout after {total_wait_time:.1f}s")
        raise Exception(f"Snapshot completion timeout after {total_wait_time:.1f}s")
        
    except Exception as e:
        wait_error_time = time.time() - wait_start_time if 'wait_start_time' in locals() else 0
        logger.error(f"Error waiting for snapshot after {wait_error_time:.2f}s: {str(e)}")
        raise

def export_snapshot_to_s3(snapshot_name):
    """ Export snapshot to S3 using copy_snapshot """
    try:
        logger.info(f"Starting S3 export for snapshot: {snapshot_name}")
        export_start_time = time.time()
        
        elasticache_client = boto3.client('elasticache', region_name='ap-northeast-2')
        
        # Get source snapshot details before copy
        try:
            snapshot_response = elasticache_client.describe_snapshots(SnapshotName=snapshot_name)
            snapshots = snapshot_response.get('Snapshots', [])
            if snapshots:
                source_snapshot = snapshots[0]
                source_size = source_snapshot.get('SnapshotSizeInBytes', 0)
                source_size_mb = source_size / 1024 / 1024 if source_size > 0 else 0
                logger.info(f"Source snapshot size before S3 copy: {source_size:,} bytes ({source_size_mb:.1f} MB)")
            else:
                logger.warning(f"Could not retrieve size information for snapshot: {snapshot_name}")
        except Exception as e:
            logger.warning(f"Failed to get snapshot size before copy: {str(e)}")
        
        # Generate target snapshot name for S3 export
        target_snapshot_name = f"{snapshot_name}-s3-export"
        
        logger.info(f"Initiating copy_snapshot to S3 bucket: {S3_BUCKET_NAME}")
        logger.info(f"Target snapshot name: {target_snapshot_name}")
        
        response = elasticache_client.copy_snapshot(
            SourceSnapshotName=snapshot_name,
            TargetSnapshotName=target_snapshot_name,
            TargetBucket=S3_BUCKET_NAME
        )
        
        copied_snapshot = response.get('Snapshot', {})
        snapshot_arn = copied_snapshot.get('SnapshotArn', 'Unknown')
        
        export_initiation_time = time.time() - export_start_time
        logger.info(f"S3 export initiated successfully in {export_initiation_time:.2f}s")
        logger.info(f"Target Snapshot: {target_snapshot_name}")
        logger.info(f"Snapshot ARN: {snapshot_arn}")
        
        return target_snapshot_name, f"s3://{S3_BUCKET_NAME}/{target_snapshot_name}"
        
    except Exception as e:
        export_error_time = time.time() - export_start_time if 'export_start_time' in locals() else 0
        logger.error(f"S3 export failed after {export_error_time:.2f}s: {str(e)}")
        raise

def wait_for_export_completion(source_snapshot_name, max_wait_time=300):
    """ Wait for S3 export to complete by checking source snapshot status (max 5 minutes) """
    try:
        logger.info(f"Waiting for S3 export completion of source snapshot: {source_snapshot_name}")
        wait_start_time = time.time()
        
        elasticache_client = boto3.client('elasticache', region_name='ap-northeast-2')
        check_interval = 30  # Check every 30 seconds
        checks_performed = 0
        
        while time.time() - wait_start_time < max_wait_time:
            try:
                response = elasticache_client.describe_snapshots(SnapshotName=source_snapshot_name)
                snapshots = response.get('Snapshots', [])
                
                if not snapshots:
                    raise Exception(f"Source snapshot {source_snapshot_name} not found")
                
                snapshot = snapshots[0]
                status = snapshot.get('SnapshotStatus', 'Unknown')
                checks_performed += 1
                elapsed_time = time.time() - wait_start_time
                
                logger.info(f"Export Check #{checks_performed}: Source snapshot status='{status}', Elapsed={elapsed_time:.1f}s")
                
                if status == 'available':
                    total_wait_time = time.time() - wait_start_time
                    logger.info(f"S3 export completed successfully after {checks_performed} checks in {total_wait_time:.1f}s")
                    logger.info(f"Source snapshot {source_snapshot_name} is now available for cleanup")
                    return True
                    
                elif status == 'failed':
                    raise Exception(f"S3 export failed with source snapshot status: {status} after {checks_performed} checks")
                elif status in ['copying']:
                    logger.info(f"S3 export in progress... Source snapshot status: {status} (Check #{checks_performed})")
                else:
                    logger.info(f"Source snapshot status: {status} (Check #{checks_performed})")
                
                # Additional progress logging every 5 minutes
                if checks_performed % 10 == 0:
                    logger.info(f"Long-running S3 export - Check #{checks_performed}, Status: {status}, {elapsed_time:.1f}s elapsed")
                
                time.sleep(check_interval)
                
            except Exception as e:
                if "SnapshotNotFoundFault" in str(e):
                    logger.warning(f"Source snapshot not yet visible, continuing to wait...")
                    time.sleep(check_interval)
                    continue
                else:
                    raise
        
        # Timeout reached
        total_wait_time = time.time() - wait_start_time
        logger.error(f"S3 export completion timeout after {total_wait_time:.1f}s")
        raise Exception(f"S3 export completion timeout after {total_wait_time:.1f}s")
        
    except Exception as e:
        wait_error_time = time.time() - wait_start_time if 'wait_start_time' in locals() else 0
        logger.error(f"Error waiting for S3 export after {wait_error_time:.2f}s: {str(e)}")
        raise

def cleanup_snapshot(snapshot_name):
    """ Delete the source snapshot - only snapshots without s3-export suffix """
    try:
        # Skip cleanup if this is an export snapshot (has s3-export suffix)
        if '-s3-export' in snapshot_name:
            logger.info(f"Skipping cleanup of export snapshot: {snapshot_name}")
            return
            
        logger.info(f"Cleaning up source snapshot: {snapshot_name}")
        cleanup_start_time = time.time()
        
        elasticache_client = boto3.client('elasticache', region_name='ap-northeast-2')
        
        # Verify snapshot state before deletion
        try:
            response = elasticache_client.describe_snapshots(SnapshotName=snapshot_name)
            snapshots = response.get('Snapshots', [])
            
            if not snapshots:
                logger.warning(f"Snapshot {snapshot_name} not found for cleanup")
                return
                
            snapshot = snapshots[0]
            status = snapshot.get('SnapshotStatus', 'Unknown')
            
            if status not in ['available', 'failed']:
                logger.warning(f"Snapshot {snapshot_name} is in '{status}' state, cannot delete. Skipping cleanup.")
                return
                
            logger.info(f"Snapshot {snapshot_name} is in '{status}' state, proceeding with deletion")
            
        except Exception as e:
            logger.warning(f"Could not verify snapshot state before cleanup: {str(e)}")
            # Continue with cleanup attempt
        
        elasticache_client.delete_snapshot(SnapshotName=snapshot_name)
        
        cleanup_time = time.time() - cleanup_start_time
        logger.info(f"Source snapshot cleanup completed in {cleanup_time:.2f}s")
        
    except Exception as e:
        cleanup_error_time = time.time() - cleanup_start_time if 'cleanup_start_time' in locals() else 0
        logger.warning(f"Snapshot cleanup failed after {cleanup_error_time:.2f}s: {str(e)}")
        # Don't raise exception for cleanup failures

def lambda_handler(event, context):
    """ AWS Lambda entry point """
    lambda_start_time = time.time()
    logger.info("=== ElastiCache Snapshot Backup Lambda Started ===")
    logger.info(f"Lambda Request ID: {context.aws_request_id if context else 'N/A'}")
    logger.info(f"Remaining time: {context.get_remaining_time_in_millis() if context else 'N/A'}ms")
    
    snapshot_name = None
    step_times = {}
    
    try:
        # Step 1: Create snapshot
        step1_start = time.time()
        logger.info("Step 1 (Snapshot Creation): Creating ElastiCache snapshot...")
        snapshot_name, snapshot_info = create_elasticache_snapshot()
        step_times['snapshot_creation'] = round(time.time() - step1_start, 2)
        elapsed_total = round(time.time() - lambda_start_time, 2)
        logger.info(f"Step 1 (Snapshot Creation) completed in {step_times['snapshot_creation']}s (total elapsed: {elapsed_total}s)")
        
        # Step 2: Wait for completion
        step2_start = time.time()
        logger.info("Step 2 (Snapshot Wait): Waiting for snapshot completion...")
        completed_snapshot = wait_for_snapshot_completion(snapshot_name)
        step_times['snapshot_wait'] = round(time.time() - step2_start, 2)
        elapsed_total = round(time.time() - lambda_start_time, 2)
        logger.info(f"Step 2 (Snapshot Wait) completed in {step_times['snapshot_wait']}s (total elapsed: {elapsed_total}s)")
        
        # Step 3: Export to S3
        step3_start = time.time()
        logger.info("Step 3 (S3 Export): Copying snapshot to S3...")
        target_snapshot_name, s3_location = export_snapshot_to_s3(snapshot_name)
        step_times['s3_export'] = round(time.time() - step3_start, 2)
        elapsed_total = round(time.time() - lambda_start_time, 2)
        logger.info(f"Step 3 (S3 Export) completed in {step_times['s3_export']}s (total elapsed: {elapsed_total}s)")
        
        # Step 4: Wait for S3 export completion
        step4_start = time.time()
        logger.info("Step 4 (Export Wait): Waiting for S3 export completion...")
        wait_for_export_completion(snapshot_name)
        step_times['export_wait'] = round(time.time() - step4_start, 2)
        elapsed_total = round(time.time() - lambda_start_time, 2)
        logger.info(f"Step 4 (Export Wait) completed in {step_times['export_wait']}s (total elapsed: {elapsed_total}s)")
        
        # Step 5: Cleanup
        step5_start = time.time()
        logger.info("Step 5 (Cleanup): Cleaning up source snapshot...")
        cleanup_snapshot(snapshot_name)
        step_times['cleanup'] = round(time.time() - step5_start, 2)
        elapsed_total = round(time.time() - lambda_start_time, 2)
        logger.info(f"Step 5 (Cleanup) completed in {step_times['cleanup']}s (total elapsed: {elapsed_total}s)")
        
        # Calculate total execution time
        total_execution_time = time.time() - lambda_start_time
        
        success_message = {
            "status": "Success",
            "message": "ElastiCache snapshot backup completed successfully",
            "total_execution_time_seconds": round(total_execution_time, 2),
            "step_timings": step_times,
            "cache_cluster": CACHE_CLUSTER_ID,
            "snapshot_name": snapshot_name,
            "target_snapshot_name": target_snapshot_name,
            "s3_location": s3_location,
            "s3_bucket": S3_BUCKET_NAME
        }
        
        # Detailed timing summary
        logger.info("=== EXECUTION TIMING SUMMARY ===")
        logger.info(f"Snapshot Creation: {step_times['snapshot_creation']}s")
        logger.info(f"Snapshot Wait: {step_times['snapshot_wait']}s")
        logger.info(f"S3 Export: {step_times['s3_export']}s")
        logger.info(f"Export Wait: {step_times['export_wait']}s")
        logger.info(f"Cleanup: {step_times['cleanup']}s")
        logger.info(f"TOTAL EXECUTION TIME: {total_execution_time:.2f}s")
        logger.info("=== Lambda execution completed successfully ===")
        logger.info(f"Final result: {success_message}")
        
        return success_message
        
    except Exception as e:
        total_execution_time = time.time() - lambda_start_time
        
        # Attempt cleanup if snapshot was created
        if snapshot_name:
            cleanup_start = time.time()
            logger.info("Attempting cleanup due to error...")
            try:
                cleanup_snapshot(snapshot_name)
                step_times['cleanup_on_error'] = round(time.time() - cleanup_start, 2)
                logger.info(f"Error cleanup completed in {step_times['cleanup_on_error']}s")
            except Exception as cleanup_error:
                step_times['cleanup_on_error'] = round(time.time() - cleanup_start, 2)
                logger.warning(f"Cleanup during error handling failed after {step_times['cleanup_on_error']}s: {str(cleanup_error)}")
        
        error_message = {
            "status": "Error",
            "message": str(e),
            "total_execution_time_seconds": round(total_execution_time, 2),
            "step_timings": step_times,
            "cache_cluster": CACHE_CLUSTER_ID,
            "snapshot_name": snapshot_name,
            "error_type": type(e).__name__
        }
        
        logger.error("=== EXECUTION TIMING SUMMARY (ERROR) ===")
        for step, duration in step_times.items():
            logger.error(f"{step}: {duration}s")
        logger.error(f"TOTAL EXECUTION TIME: {total_execution_time:.2f}s")
        logger.error(f"=== Lambda execution failed after {total_execution_time:.2f}s ===")
        logger.error(f"Error details: {error_message}")
        
        return error_message

# Local execution
if __name__ == "__main__":
    result = lambda_handler(None, None)
    print(result)
