use anyhow::{Context, Result};
use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::Path;

use crate::models::{JobMetricSummary, ParsedMetrics, PrometheusMetricsFile};

/// Parse the Mimirtool JSON file and extract metrics per job
pub fn parse_metrics_file<P: AsRef<Path>>(path: P) -> Result<ParsedMetrics> {
    let path = path.as_ref();

    // Read the JSON file
    let content = fs::read_to_string(path)
        .with_context(|| format!("Failed to read JSON file: {}", path.display()))?;

    // Parse JSON
    let data: PrometheusMetricsFile =
        serde_json::from_str(&content).context("Failed to parse JSON file")?;

    // Build job -> metrics mapping
    let mut job_metrics_map: HashMap<String, Vec<String>> = HashMap::new();
    let mut job_metric_sets: HashMap<String, HashSet<String>> = HashMap::new();

    for metric_count in &data.additional_metric_counts {
        let metric_name = &metric_count.metric;

        if metric_name.is_empty() {
            continue;
        }

        for job_count in &metric_count.job_counts {
            let job_name = &job_count.job;

            if job_name.is_empty() {
                continue;
            }

            // Use HashSet to avoid duplicates
            job_metric_sets
                .entry(job_name.clone())
                .or_default()
                .insert(metric_name.clone());
        }
    }

    // Convert HashSet to sorted Vec for each job
    for (job_name, metric_set) in job_metric_sets {
        let mut metrics: Vec<String> = metric_set.into_iter().collect();
        metrics.sort();
        job_metrics_map.insert(job_name, metrics);
    }

    // Build summary data
    let mut summary_data: Vec<JobMetricSummary> = job_metrics_map
        .iter()
        .map(|(job_name, metrics)| JobMetricSummary {
            job_name: job_name.clone(),
            metric_count: metrics.len(),
        })
        .collect();

    // Sort summary by job name
    summary_data.sort_by(|a, b| a.job_name.cmp(&b.job_name));

    // Calculate total unique metrics
    let mut unique_metrics = HashSet::new();
    for metrics in job_metrics_map.values() {
        for metric in metrics {
            unique_metrics.insert(metric.clone());
        }
    }
    let total_unique_metrics = unique_metrics.len();

    Ok(ParsedMetrics {
        job_metrics_map,
        summary_data,
        total_unique_metrics,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[test]
    fn test_parse_metrics_file() {
        // Create a temporary JSON file
        let mut temp_file = NamedTempFile::new().unwrap();
        let json_content = r#"{
            "additional_metric_counts": [
                {
                    "metric": "http_requests_total",
                    "job_counts": [
                        {"job": "api-server"},
                        {"job": "web-server"}
                    ]
                },
                {
                    "metric": "http_errors_total",
                    "job_counts": [
                        {"job": "api-server"}
                    ]
                }
            ]
        }"#;
        temp_file.write_all(json_content.as_bytes()).unwrap();

        let result = parse_metrics_file(temp_file.path()).unwrap();

        assert_eq!(result.total_unique_metrics, 2);
        assert_eq!(result.summary_data.len(), 2);
        assert_eq!(result.job_metrics_map.get("api-server").unwrap().len(), 2);
        assert_eq!(result.job_metrics_map.get("web-server").unwrap().len(), 1);
    }
}
