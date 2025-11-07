use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct ExecutionSummary {
    pub status: String,
    pub message: String,
    pub total_execution_time_seconds: f64,
    pub step_timings: StepTimings,
    pub cache_cluster: String,
    pub snapshot_name: Option<String>,
    pub target_snapshot_name: Option<String>,
    pub s3_location: Option<String>,
    pub s3_bucket: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub retention_info: Option<RetentionInfo>,
}

#[derive(Debug, Serialize, Default)]
pub struct StepTimings {
    pub snapshot_creation: f64,
    pub snapshot_wait: f64,
    pub s3_export: f64,
    pub export_wait: f64,
    pub cleanup: f64,
    pub retention: f64,
}

#[derive(Debug, Serialize)]
pub struct RetentionInfo {
    pub enabled: bool,
    pub retention_count: u32,
    pub deleted_count: usize,
}
