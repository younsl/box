use regex::Regex;
use std::collections::HashMap;

use crate::models::{GroupInfo, RelabelRule};

const MAX_REGEX_LENGTH: usize = 1000;

/// Group metrics by their prefix (text before first underscore)
pub fn group_metrics_by_prefix(metrics: &[String]) -> HashMap<String, Vec<String>> {
    let prefix_regex = Regex::new(r"^([^_]+)_").unwrap();
    let mut groups: HashMap<String, Vec<String>> = HashMap::new();

    for metric in metrics {
        let prefix = if let Some(captures) = prefix_regex.captures(metric) {
            captures.get(1).unwrap().as_str().to_string()
        } else {
            "other".to_string()
        };

        groups.entry(prefix).or_default().push(metric.clone());
    }

    // Sort metrics within each group for stable output
    for metrics_list in groups.values_mut() {
        metrics_list.sort();
    }

    groups
}

/// Generate Prometheus relabel rules from metrics
/// Returns (rules, group_info) where group_info is for reporting
pub fn generate_relabel_rules(metrics: &[String]) -> (Vec<RelabelRule>, Vec<GroupInfo>) {
    let groups = group_metrics_by_prefix(metrics);
    let mut relabel_rules = Vec::new();
    let mut group_info_list = Vec::new();

    // Sort prefixes for consistent output
    let mut prefixes: Vec<_> = groups.keys().cloned().collect();
    prefixes.sort();

    for prefix in prefixes {
        let metrics_group = groups.get(&prefix).unwrap();
        let chunks = chunk_metrics(metrics_group, MAX_REGEX_LENGTH);

        let pattern_info = if prefix == "other" {
            "(no prefix)".to_string()
        } else {
            format!("{}_*", prefix)
        };

        for (i, chunk) in chunks.iter().enumerate() {
            let part_str = format!("{}/{}", i + 1, chunks.len());
            let regex_content = chunk.join("|");

            let rule = RelabelRule {
                source_labels: vec!["__name__".to_string()],
                regex: regex_content,
                action: "drop".to_string(),
            };

            relabel_rules.push(rule);

            group_info_list.push(GroupInfo {
                prefix: prefix.clone(),
                pattern: pattern_info.clone(),
                part: part_str,
                count: chunk.len(),
            });
        }
    }

    (relabel_rules, group_info_list)
}

/// Split metrics into chunks that fit within max_regex_length
fn chunk_metrics(metrics: &[String], max_length: usize) -> Vec<Vec<String>> {
    let mut chunks = Vec::new();
    let mut current_chunk = Vec::new();
    let mut current_length = 0;

    for metric in metrics {
        // Account for the pipe separator: +1
        let metric_length = metric.len() + 1;

        if current_length + metric_length > max_length && !current_chunk.is_empty() {
            chunks.push(current_chunk);
            current_chunk = Vec::new();
            current_length = 0;
        }

        current_chunk.push(metric.clone());
        current_length += metric_length;
    }

    if !current_chunk.is_empty() {
        chunks.push(current_chunk);
    }

    chunks
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_group_metrics_by_prefix() {
        let metrics = vec![
            "http_requests_total".to_string(),
            "http_errors_total".to_string(),
            "grpc_requests_total".to_string(),
            "noprefix".to_string(), // No underscore, will be "other"
        ];

        let groups = group_metrics_by_prefix(&metrics);

        assert_eq!(groups.len(), 3);
        assert_eq!(groups.get("http").unwrap().len(), 2);
        assert_eq!(groups.get("grpc").unwrap().len(), 1);
        assert_eq!(groups.get("other").unwrap().len(), 1);
    }

    #[test]
    fn test_chunk_metrics() {
        let metrics: Vec<String> = (0..10).map(|i| format!("metric_{}", i)).collect();
        let chunks = chunk_metrics(&metrics, 30);

        // Each metric is about 8-9 chars + 1 for separator
        // So we expect multiple chunks
        assert!(chunks.len() > 1);
    }

    #[test]
    fn test_generate_relabel_rules() {
        let metrics = vec![
            "http_requests_total".to_string(),
            "http_errors_total".to_string(),
        ];

        let (rules, info) = generate_relabel_rules(&metrics);

        assert_eq!(rules.len(), 1);
        assert_eq!(info.len(), 1);
        assert_eq!(rules[0].action, "drop");
        assert_eq!(rules[0].source_labels, vec!["__name__"]);
        assert!(rules[0].regex.contains("http_requests_total"));
    }
}
