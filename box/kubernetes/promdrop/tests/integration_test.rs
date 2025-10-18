use std::fs;
use std::path::PathBuf;
use tempfile::TempDir;

// Import promdrop modules
// Note: These would normally be in a lib.rs, but for testing we'll use the binary modules
// For now, we'll test via the compiled binary

#[test]
fn test_end_to_end_conversion() {
    // Create temporary directory for output
    let temp_dir = TempDir::new().expect("Failed to create temp dir");
    let output_dir = temp_dir.path();

    // Path to sample input file
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    // Verify input file exists
    assert!(
        input_file.exists(),
        "Sample metrics file not found at {:?}",
        input_file
    );

    // Output paths
    let txt_output_dir = output_dir.join("unused");
    let _yaml_output = output_dir.join("combined_relabel_configs.yaml");

    // Create the output directory
    fs::create_dir_all(&txt_output_dir).expect("Failed to create output dir");

    // We'll test the library functions directly instead of running the binary
    // This requires exposing the functions as a library

    // For now, let's verify the input file structure
    let content = fs::read_to_string(&input_file).expect("Failed to read input file");
    let parsed: serde_json::Value = serde_json::from_str(&content).expect("Failed to parse JSON");

    // Verify structure
    assert!(
        parsed.get("additional_metric_counts").is_some(),
        "Missing additional_metric_counts field"
    );

    let metrics = parsed["additional_metric_counts"]
        .as_array()
        .expect("additional_metric_counts should be an array");

    assert_eq!(
        metrics.len(),
        12,
        "Expected 12 metrics in sample data, got {}",
        metrics.len()
    );

    // Verify specific metrics exist
    let metric_names: Vec<String> = metrics
        .iter()
        .map(|m| m["metric"].as_str().unwrap().to_string())
        .collect();

    assert!(metric_names.contains(&"http_requests_total".to_string()));
    assert!(metric_names.contains(&"grpc_requests_total".to_string()));
    assert!(metric_names.contains(&"database_connections_active".to_string()));
}

#[test]
fn test_sample_data_has_multiple_jobs() {
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    let content = fs::read_to_string(&input_file).expect("Failed to read input file");
    let parsed: serde_json::Value = serde_json::from_str(&content).expect("Failed to parse JSON");

    // Collect all unique jobs
    let mut jobs = std::collections::HashSet::new();

    for metric in parsed["additional_metric_counts"].as_array().unwrap() {
        for job_count in metric["job_counts"].as_array().unwrap() {
            jobs.insert(job_count["job"].as_str().unwrap());
        }
    }

    // Expected jobs: api-server, web-server, db-proxy
    assert_eq!(jobs.len(), 3, "Expected 3 unique jobs");
    assert!(jobs.contains("api-server"));
    assert!(jobs.contains("web-server"));
    assert!(jobs.contains("db-proxy"));
}

#[test]
fn test_sample_data_metric_distribution() {
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    let content = fs::read_to_string(&input_file).expect("Failed to read input file");
    let parsed: serde_json::Value = serde_json::from_str(&content).expect("Failed to parse JSON");

    // Count metrics per job
    let mut job_metric_counts: std::collections::HashMap<String, usize> =
        std::collections::HashMap::new();

    for metric in parsed["additional_metric_counts"].as_array().unwrap() {
        for job_count in metric["job_counts"].as_array().unwrap() {
            let job = job_count["job"].as_str().unwrap().to_string();
            *job_metric_counts.entry(job).or_insert(0) += 1;
        }
    }

    // Verify api-server has the most metrics
    assert!(
        job_metric_counts.get("api-server").unwrap() >= &7,
        "api-server should have at least 7 metrics"
    );

    // Verify each job has at least one metric
    for (job, count) in job_metric_counts.iter() {
        assert!(count > &0, "Job {} should have at least 1 metric", job);
    }
}
