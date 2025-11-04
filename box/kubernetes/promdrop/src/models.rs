use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Top-level structure of the input JSON from Mimirtool
#[derive(Debug, Deserialize)]
pub struct PrometheusMetricsFile {
    #[serde(rename = "additional_metric_counts")]
    pub additional_metric_counts: Vec<RawMetricCount>,
}

/// Represents a single metric with its job counts
#[derive(Debug, Deserialize)]
pub struct RawMetricCount {
    pub metric: String,
    pub job_counts: Vec<RawJobCount>,
}

/// Represents a job that uses a metric
#[derive(Debug, Deserialize)]
pub struct RawJobCount {
    pub job: String,
}

/// Summary of metrics per job
#[derive(Debug, Clone)]
pub struct JobMetricSummary {
    pub job_name: String,
    pub metric_count: usize,
}

/// Information about a metric group for reporting
#[derive(Debug, Clone)]
pub struct GroupInfo {
    pub prefix: String,
    pub pattern: String,
    pub part: String,
    pub count: usize,
}

/// Prometheus relabel rule configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RelabelRule {
    pub source_labels: Vec<String>,
    pub regex: String,
    pub action: String,
}

/// Job configuration with relabel rules
#[derive(Debug, Serialize, Deserialize)]
pub struct JobConfig {
    pub job_name: String,
    pub metric_relabel_configs: Vec<RelabelRule>,
}

/// Parsed metrics data organized by job
pub struct ParsedMetrics {
    pub job_metrics_map: HashMap<String, Vec<String>>,
    pub summary_data: Vec<JobMetricSummary>,
    pub total_unique_metrics: usize,
}
