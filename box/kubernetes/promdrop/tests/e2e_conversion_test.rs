use promdrop::{grouper, models, output, parser};
use std::fs;
use std::path::PathBuf;
use tempfile::TempDir;

#[test]
fn test_full_conversion_pipeline() {
    // Setup: paths and temp directory
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    let temp_dir = TempDir::new().expect("Failed to create temp dir");
    let txt_output_dir = temp_dir.path().join("unused");
    let yaml_output = temp_dir.path().join("combined_relabel_configs.yaml");

    fs::create_dir_all(&txt_output_dir).expect("Failed to create output dir");

    // Step 1: Parse the input file
    let parsed = parser::parse_metrics_file(&input_file).expect("Failed to parse metrics");

    // Verify parsing results
    assert!(
        parsed.summary_data.len() >= 3,
        "Should have at least 3 jobs"
    );
    assert!(
        parsed.total_unique_metrics >= 12,
        "Should have at least 12 unique metrics"
    );

    println!("✓ Parsed {} unique metrics", parsed.total_unique_metrics);
    println!("✓ Found {} jobs", parsed.summary_data.len());

    // Step 2: Verify job metrics map
    for (job_name, metrics) in &parsed.job_metrics_map {
        println!("  - Job '{}' has {} metrics", job_name, metrics.len());
        assert!(
            !metrics.is_empty(),
            "Job {} should have at least one metric",
            job_name
        );
    }

    // Step 3: Test metric grouping for api-server
    if let Some(api_metrics) = parsed.job_metrics_map.get("api-server") {
        let (rules, group_info) = grouper::generate_relabel_rules(api_metrics);

        println!("\n✓ Generated {} relabel rules for api-server", rules.len());
        assert!(!rules.is_empty(), "Should generate at least one rule");

        // Verify rule structure
        for rule in &rules {
            assert_eq!(rule.source_labels, vec!["__name__"]);
            assert_eq!(rule.action, "drop");
            assert!(!rule.regex.is_empty(), "Regex should not be empty");
        }

        // Verify group info
        assert!(
            !group_info.is_empty(),
            "Should have group information for reporting"
        );

        for info in &group_info {
            println!(
                "  - Prefix: {}, Pattern: {}, Part: {}, Count: {}",
                info.prefix, info.pattern, info.part, info.count
            );
        }
    }

    // Step 4: Generate summary file
    output::generate_summary_file(
        &txt_output_dir,
        &parsed.summary_data,
        input_file.to_str().unwrap(),
    )
    .expect("Failed to generate summary file");

    let summary_file = txt_output_dir.join("summary.txt");
    assert!(summary_file.exists(), "Summary file should be created");

    let summary_content = fs::read_to_string(&summary_file).expect("Failed to read summary");
    assert!(
        summary_content.contains("Unused Metric Summary"),
        "Summary should have header"
    );
    assert!(
        summary_content.contains("api-server"),
        "Summary should mention api-server"
    );

    println!("✓ Generated summary file");

    // Step 5: Generate per-job txt files
    for (job_name, metrics) in &parsed.job_metrics_map {
        output::generate_txt_file(&txt_output_dir, job_name, metrics)
            .expect("Failed to generate txt file");

        let job_file = txt_output_dir.join(format!("{}_unused_metrics.txt", job_name));
        assert!(job_file.exists(), "Job file for {} should exist", job_name);

        let job_content = fs::read_to_string(&job_file).expect("Failed to read job file");
        let lines: Vec<&str> = job_content.lines().collect();

        assert_eq!(
            lines.len(),
            metrics.len(),
            "Job file should have one line per metric"
        );

        println!("✓ Generated txt file for job '{}'", job_name);
    }

    // Step 6: Generate YAML configuration
    let processed_count = output::generate_yaml_file(
        &yaml_output,
        &parsed.job_metrics_map,
        &parsed.summary_data,
        parsed.total_unique_metrics,
    )
    .expect("Failed to generate YAML");

    assert!(processed_count > 0, "Should process at least one job");
    assert!(yaml_output.exists(), "YAML output file should exist");

    let yaml_content = fs::read_to_string(&yaml_output).expect("Failed to read YAML");

    // Verify YAML structure
    assert!(
        yaml_content.contains("job_name:"),
        "YAML should have job_name field"
    );
    assert!(
        yaml_content.contains("metric_relabel_configs:"),
        "YAML should have metric_relabel_configs"
    );
    assert!(
        yaml_content.contains("source_labels:"),
        "YAML should have source_labels"
    );
    assert!(
        yaml_content.contains("regex:"),
        "YAML should have regex field"
    );
    assert!(
        yaml_content.contains("action: drop"),
        "YAML should have drop action"
    );

    // Verify comments are included
    assert!(
        yaml_content.contains("# Summary:"),
        "YAML should have summary comments"
    );

    println!("✓ Generated YAML configuration file");
    println!("\n✅ Full conversion pipeline test passed!");
}

#[test]
fn test_metric_prefix_grouping() {
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    let parsed = parser::parse_metrics_file(&input_file).expect("Failed to parse");

    // Check that metrics are properly grouped by prefix
    if let Some(api_metrics) = parsed.job_metrics_map.get("api-server") {
        let groups = grouper::group_metrics_by_prefix(api_metrics);

        // Expected prefixes: http, grpc, database, memory, cpu, other
        println!("Found {} prefix groups", groups.len());

        for (prefix, metrics) in &groups {
            println!("  Prefix '{}': {} metrics", prefix, metrics.len());

            if prefix != "other" {
                // Verify all metrics in group start with prefix_
                for metric in metrics {
                    assert!(
                        metric.starts_with(&format!("{}_", prefix))
                            || metric == "orphan_metric_without_prefix",
                        "Metric '{}' should start with '{}_'",
                        metric,
                        prefix
                    );
                }
            }
        }

        // Verify specific groups exist
        assert!(groups.contains_key("http"), "Should have http prefix group");
        assert!(groups.contains_key("grpc"), "Should have grpc prefix group");
        assert!(
            groups.contains_key("database"),
            "Should have database prefix group"
        );
    }
}

#[test]
fn test_yaml_parsing_validity() {
    let input_file = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests")
        .join("fixtures")
        .join("sample-metrics.json");

    let temp_dir = TempDir::new().expect("Failed to create temp dir");
    let yaml_output = temp_dir.path().join("test_output.yaml");

    let parsed = parser::parse_metrics_file(&input_file).expect("Failed to parse");

    output::generate_yaml_file(
        &yaml_output,
        &parsed.job_metrics_map,
        &parsed.summary_data,
        parsed.total_unique_metrics,
    )
    .expect("Failed to generate YAML");

    // Try to parse the generated YAML to ensure it's valid
    let yaml_content = fs::read_to_string(&yaml_output).expect("Failed to read YAML");

    // The YAML file contains multiple documents separated by comments and newlines
    // Let's verify the basic structure instead of parsing the whole thing
    assert!(
        yaml_content.contains("job_name:"),
        "YAML should contain job_name field"
    );
    assert!(
        yaml_content.contains("metric_relabel_configs:"),
        "YAML should contain metric_relabel_configs"
    );
    assert!(
        yaml_content.contains("source_labels:"),
        "YAML should contain source_labels"
    );

    // Try to parse each job config separately by splitting on "# Summary:"
    let job_sections: Vec<&str> = yaml_content
        .split("# Summary:")
        .filter(|s| !s.trim().is_empty())
        .collect();

    assert!(
        !job_sections.is_empty(),
        "Should have at least one job section"
    );

    for section in job_sections {
        // Skip the first line (the summary comment) and try to parse the YAML
        let yaml_part = section.lines().skip(1).collect::<Vec<_>>().join("\n");

        if !yaml_part.trim().is_empty() {
            let parse_result: Result<models::JobConfig, _> = serde_yaml::from_str(&yaml_part);
            assert!(
                parse_result.is_ok(),
                "Job config YAML should be valid: {:?}\n\nYAML content:\n{}",
                parse_result.err(),
                yaml_part
            );
        }
    }

    println!("✓ Generated YAML is valid and parseable");
}

#[test]
fn test_empty_metrics_handling() {
    // Create a minimal JSON with no metrics
    let temp_dir = TempDir::new().expect("Failed to create temp dir");
    let empty_json = temp_dir.path().join("empty.json");

    let empty_content = r#"{
        "additional_metric_counts": []
    }"#;

    fs::write(&empty_json, empty_content).expect("Failed to write test file");

    let parsed = parser::parse_metrics_file(&empty_json).expect("Should parse empty file");

    assert_eq!(parsed.summary_data.len(), 0, "Should have no jobs");
    assert_eq!(
        parsed.total_unique_metrics, 0,
        "Should have no unique metrics"
    );

    println!("✓ Correctly handles empty metrics file");
}
