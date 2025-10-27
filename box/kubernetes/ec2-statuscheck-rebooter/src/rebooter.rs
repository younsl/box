use anyhow::Result;
use std::collections::HashMap;
use std::time::Duration;
use tokio::time::sleep;
use tracing::{debug, error, info, warn};

use crate::config::Config;
use crate::ec2::{Ec2Client, InstanceStatus};
use crate::health::HealthServer;

pub struct StatusCheckRebooter {
    ec2_client: Ec2Client,
    config: Config,
    failure_tracker: HashMap<String, u32>,
    health_server: HealthServer,
}

impl StatusCheckRebooter {
    pub async fn new(config: Config, health_server: HealthServer) -> Result<Self> {
        let ec2_client = Ec2Client::new(config.region.clone()).await?;

        Ok(Self {
            ec2_client,
            config,
            failure_tracker: HashMap::new(),
            health_server,
        })
    }

    pub async fn run(&mut self) -> Result<()> {
        // Test EC2 API connectivity before starting monitoring
        self.ec2_client.test_connectivity().await?;

        let actual_region = self.ec2_client.region();
        self.config.display(actual_region);

        // Mark application as ready after successful initialization
        self.health_server.set_ready(true);

        info!("Starting monitoring loop");

        let mut check_count = 0u64;
        loop {
            check_count += 1;
            debug!(
                check_number = check_count,
                tracked_failures = self.failure_tracker.len(),
                "Starting status check cycle"
            );

            if let Err(e) = self.check_and_reboot().await {
                error!(
                    error = %e,
                    check_number = check_count,
                    "Failed to check and reboot instances"
                );
            }

            debug!(
                check_number = check_count,
                next_check_in_seconds = self.config.check_interval_seconds,
                "Completed status check cycle, sleeping"
            );

            sleep(Duration::from_secs(self.config.check_interval_seconds)).await;
        }
    }

    async fn check_and_reboot(&mut self) -> Result<()> {
        debug!("Querying AWS EC2 DescribeInstanceStatus API");
        let start_time = std::time::Instant::now();

        let (total_scanned, instances) = self
            .ec2_client
            .get_instance_statuses(&self.config.tag_filters, &self.failure_tracker)
            .await?;

        let scan_duration_seconds = start_time.elapsed().as_secs_f64();

        info!(
            region = self.ec2_client.region(),
            total_scanned = total_scanned,
            impaired_count = instances.len(),
            healthy_count = total_scanned - instances.len(),
            tracked_failures = self.failure_tracker.len(),
            scan_duration_seconds = format!("{:.2}", scan_duration_seconds),
            "Completed EC2 instance status scan"
        );

        if instances.is_empty() {
            info!(
                region = self.ec2_client.region(),
                total_scanned = total_scanned,
                healthy_count = total_scanned,
                tracked_failures = self.failure_tracker.len(),
                scan_duration_seconds = format!("{:.2}", scan_duration_seconds),
                "All scanned instances are healthy with no status check failures, no reboot action required"
            );
            return Ok(());
        }

        warn!(
            impaired_instance_count = instances.len(),
            tracked_failures = self.failure_tracker.len(),
            "Found instances with status check failures"
        );

        for instance in instances {
            self.process_instance_status(instance).await?;
        }

        Ok(())
    }

    async fn process_instance_status(&mut self, status: InstanceStatus) -> Result<()> {
        let checks_until_reboot = self
            .config
            .failure_threshold
            .saturating_sub(status.failure_count);

        warn!(
            instance_id = %status.instance_id,
            instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
            instance_type = %status.instance_type,
            availability_zone = %status.availability_zone,
            system_status = %status.system_status,
            instance_status = %status.instance_status,
            failure_count = status.failure_count,
            failure_threshold = self.config.failure_threshold,
            checks_until_reboot = checks_until_reboot,
            check_interval_seconds = self.config.check_interval_seconds,
            "Instance status check failure detected"
        );

        // Update failure tracker
        self.failure_tracker
            .insert(status.instance_id.clone(), status.failure_count);

        // Check if failure count exceeds threshold
        if status.failure_count >= self.config.failure_threshold {
            info!(
                instance_id = %status.instance_id,
                instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                instance_type = %status.instance_type,
                availability_zone = %status.availability_zone,
                system_status = %status.system_status,
                instance_status = %status.instance_status,
                failure_count = status.failure_count,
                failure_threshold = self.config.failure_threshold,
                checks_until_reboot = checks_until_reboot,
                check_interval_seconds = self.config.check_interval_seconds,
                "Failure threshold reached, initiating reboot"
            );

            self.reboot_instance(&status).await?;

            // Reset failure count after reboot
            self.failure_tracker.remove(&status.instance_id);

            info!(
                instance_id = %status.instance_id,
                instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                "Removed instance from failure tracker after reboot"
            );
        } else {
            info!(
                instance_id = %status.instance_id,
                instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                instance_type = %status.instance_type,
                availability_zone = %status.availability_zone,
                system_status = %status.system_status,
                instance_status = %status.instance_status,
                failure_count = status.failure_count,
                failure_threshold = self.config.failure_threshold,
                checks_until_reboot = checks_until_reboot,
                check_interval_seconds = self.config.check_interval_seconds,
                "Instance below failure threshold, continuing to monitor"
            );
        }

        Ok(())
    }

    async fn reboot_instance(&self, status: &InstanceStatus) -> Result<()> {
        if self.config.dry_run {
            warn!(
                instance_id = %status.instance_id,
                instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                instance_type = %status.instance_type,
                availability_zone = %status.availability_zone,
                system_status = %status.system_status,
                instance_status = %status.instance_status,
                failure_count = status.failure_count,
                action = "reboot",
                "DRY RUN: Would reboot instance (no action taken)"
            );
            return Ok(());
        }

        warn!(
            instance_id = %status.instance_id,
            instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
            instance_type = %status.instance_type,
            availability_zone = %status.availability_zone,
            system_status = %status.system_status,
            instance_status = %status.instance_status,
            failure_count = status.failure_count,
            action = "reboot",
            "Initiating instance reboot due to persistent status check failures"
        );

        match self.ec2_client.reboot_instance(&status.instance_id).await {
            Ok(_) => {
                info!(
                    instance_id = %status.instance_id,
                    instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                    instance_type = %status.instance_type,
                    availability_zone = %status.availability_zone,
                    action = "reboot",
                    result = "success",
                    "Successfully initiated instance reboot"
                );
                Ok(())
            }
            Err(e) => {
                error!(
                    instance_id = %status.instance_id,
                    instance_name = %status.instance_name.as_deref().unwrap_or("N/A"),
                    instance_type = %status.instance_type,
                    availability_zone = %status.availability_zone,
                    error = %e,
                    action = "reboot",
                    result = "failed",
                    "Failed to reboot instance"
                );
                Err(anyhow::anyhow!(
                    "Failed to reboot instance {}: {}",
                    status.instance_id,
                    e
                ))
            }
        }
    }
}
