use anyhow::{Context, Result};
use aws_config::BehaviorVersion;
use aws_sdk_ec2::Client;
use aws_sdk_ec2::types::Filter;
use tracing::{debug, info, warn};

#[derive(Debug, Clone)]
pub struct InstanceStatus {
    pub instance_id: String,
    pub instance_name: Option<String>,
    pub instance_type: String,
    pub availability_zone: String,
    pub system_status: String,
    pub instance_status: String,
    pub failure_count: u32,
}

pub struct Ec2Client {
    client: Client,
    region: String,
}

impl Ec2Client {
    pub async fn new(region: Option<String>) -> Result<Self> {
        debug!("Initializing AWS SDK configuration");

        let config = if let Some(region) = &region {
            info!(region = %region, "Using explicit AWS region");
            aws_config::defaults(BehaviorVersion::latest())
                .region(aws_config::Region::new(region.clone()))
                .load()
                .await
        } else {
            debug!("Using default AWS region from environment/IMDS");
            aws_config::load_defaults(BehaviorVersion::latest()).await
        };

        let region_name = config.region().map(|r| r.as_ref()).unwrap_or("unknown");
        let client = Client::new(&config);

        info!(
            region = %region_name,
            "AWS EC2 client initialized successfully"
        );

        Ok(Self {
            client,
            region: region_name.to_string(),
        })
    }

    pub fn region(&self) -> &str {
        &self.region
    }

    /// Test EC2 API connectivity by making a simple DescribeRegions call
    pub async fn test_connectivity(&self) -> Result<()> {
        let endpoint = format!("https://ec2.{}.amazonaws.com", self.region);
        debug!(
            endpoint = %endpoint,
            "Testing EC2 API connectivity"
        );

        let start_time = std::time::Instant::now();

        self.client
            .describe_regions()
            .send()
            .await
            .context("Failed to connect to EC2 API endpoint")?;

        let response_time_ms = start_time.elapsed().as_millis();

        info!(
            region = %self.region,
            endpoint = %endpoint,
            response_time_ms = response_time_ms,
            "EC2 API connectivity test successful"
        );

        Ok(())
    }

    pub async fn get_instance_statuses(
        &self,
        tag_filters: &[String],
        failure_tracker: &std::collections::HashMap<String, u32>,
    ) -> Result<(usize, Vec<InstanceStatus>)> {
        let mut request = self
            .client
            .describe_instance_status()
            .include_all_instances(false); // Only running instances

        // Add tag filters if specified
        let mut filter_count = 0;
        for tag_filter in tag_filters {
            if let Some((key, value)) = tag_filter.split_once('=') {
                debug!(
                    tag_key = %key,
                    tag_value = %value,
                    "Adding tag filter to EC2 API request"
                );
                let filter = Filter::builder()
                    .name(format!("tag:{}", key))
                    .values(value)
                    .build();
                request = request.filters(filter);
                filter_count += 1;
            } else {
                warn!(
                    invalid_filter = %tag_filter,
                    "Skipping invalid tag filter (expected format: Key=Value)"
                );
            }
        }

        debug!(
            applied_filters = filter_count,
            "Sending DescribeInstanceStatus API request"
        );

        let response = request
            .send()
            .await
            .context("Failed to describe instance status")?;

        let total_instances = response.instance_statuses().len();
        debug!(
            total_instances = total_instances,
            "Received response from DescribeInstanceStatus API"
        );

        // Collect instance IDs for tag lookup
        let instance_ids: Vec<String> = response
            .instance_statuses()
            .iter()
            .filter_map(|s| s.instance_id().map(|id| id.to_string()))
            .collect();

        // Fetch tags for all instances
        let tags_map = self.get_instance_tags(&instance_ids).await?;

        let mut statuses = Vec::new();

        for status in response.instance_statuses() {
            let instance_id = status.instance_id().unwrap_or("unknown").to_string();
            let instance_name = tags_map.get(&instance_id).cloned();
            let availability_zone = status.availability_zone().unwrap_or("unknown").to_string();

            let system_status = status
                .system_status()
                .and_then(|s| s.status())
                .map(|s| s.as_str())
                .unwrap_or("unknown")
                .to_string();

            let instance_status = status
                .instance_status()
                .and_then(|s| s.status())
                .map(|s| s.as_str())
                .unwrap_or("unknown")
                .to_string();

            debug!(
                instance_id = %instance_id,
                instance_name = ?instance_name,
                availability_zone = %availability_zone,
                system_status = %system_status,
                instance_status = %instance_status,
                "Checking instance status"
            );

            // Check if either system or instance status has failed
            let has_failure = system_status == "impaired" || instance_status == "impaired";

            if has_failure {
                let failure_count = failure_tracker.get(&instance_id).unwrap_or(&0) + 1;

                debug!(
                    instance_id = %instance_id,
                    instance_name = ?instance_name,
                    failure_count = failure_count,
                    "Instance has impaired status, incrementing failure count"
                );

                statuses.push(InstanceStatus {
                    instance_id,
                    instance_name,
                    instance_type: String::new(),
                    availability_zone,
                    system_status,
                    instance_status,
                    failure_count,
                });
            }
        }

        if !statuses.is_empty() {
            info!(
                total_checked = total_instances,
                impaired_count = statuses.len(),
                "Status check summary"
            );
        }

        Ok((total_instances, statuses))
    }

    async fn get_instance_tags(
        &self,
        instance_ids: &[String],
    ) -> Result<std::collections::HashMap<String, String>> {
        if instance_ids.is_empty() {
            return Ok(std::collections::HashMap::new());
        }

        debug!(
            instance_count = instance_ids.len(),
            "Fetching tags for instances"
        );

        let response = self
            .client
            .describe_instances()
            .set_instance_ids(Some(instance_ids.to_vec()))
            .send()
            .await
            .context("Failed to describe instances for tags")?;

        let mut tags_map = std::collections::HashMap::new();

        for reservation in response.reservations() {
            for instance in reservation.instances() {
                if let Some(instance_id) = instance.instance_id() {
                    let tags = instance.tags();
                    for tag in tags {
                        if tag.key() == Some("Name") {
                            if let Some(value) = tag.value() {
                                tags_map.insert(instance_id.to_string(), value.to_string());
                            }
                        }
                    }
                }
            }
        }

        debug!(
            tagged_instances = tags_map.len(),
            "Fetched instance name tags"
        );

        Ok(tags_map)
    }

    pub async fn reboot_instance(&self, instance_id: &str) -> Result<()> {
        debug!(
            instance_id = %instance_id,
            "Calling EC2 RebootInstances API"
        );

        self.client
            .reboot_instances()
            .instance_ids(instance_id)
            .send()
            .await
            .context(format!("Failed to reboot instance {}", instance_id))?;

        Ok(())
    }
}
