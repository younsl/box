use anyhow::{Context, Result};
use prettytable::{Table, row};
use std::fs::{self, File};
use std::io::Write;
use std::path::Path;

use crate::grouper::generate_relabel_rules;
use crate::models::{GroupInfo, JobConfig, JobMetricSummary};

/// Print summary table of metrics per job
pub fn print_summary_table(summary_data: &[JobMetricSummary]) {
    let mut table = Table::new();
    table.add_row(row!["JOB NAME", "METRIC COUNT"]);

    let mut total_metrics = 0;

    for item in summary_data {
        table.add_row(row![item.job_name, item.metric_count]);
        total_metrics += item.metric_count;
    }

    table.add_row(row![
        format!("TOTAL ({} jobs)", summary_data.len()),
        total_metrics
    ]);

    table.printstd();
}

/// Print group summary for a single job
pub fn print_group_summary(group_info: &[GroupInfo]) {
    if group_info.is_empty() {
        return;
    }

    println!("  [Metric Group Summary]");
    let mut table = Table::new();
    table.add_row(row!["PREFIX", "PATTERN", "PART", "METRIC COUNT"]);

    let mut total_processed = 0;

    for info in group_info {
        table.add_row(row![info.prefix, info.pattern, info.part, info.count]);
        total_processed += info.count;
    }

    table.printstd();
    println!(
        "  Total metrics included in YAML rules: {}",
        total_processed
    );
}

/// Generate summary.txt file
pub fn generate_summary_file<P: AsRef<Path>>(
    output_dir: P,
    summary_data: &[JobMetricSummary],
    json_path: &str,
) -> Result<()> {
    let output_dir = output_dir.as_ref();
    let summary_path = output_dir.join("summary.txt");

    let mut file = File::create(&summary_path)
        .with_context(|| format!("Failed to create summary file: {}", summary_path.display()))?;

    writeln!(file, "Unused Metric Summary")?;
    writeln!(file, "Source: {}", json_path)?;
    writeln!(file)?;
    writeln!(file, "{:<30} {:<15}", "JOB NAME", "METRIC COUNT")?;
    writeln!(file, "{}", "-".repeat(50))?;

    let mut total_metrics = 0;

    for item in summary_data {
        writeln!(file, "{:<30} {:<15}", item.job_name, item.metric_count)?;
        total_metrics += item.metric_count;
    }

    writeln!(file, "{}", "-".repeat(50))?;
    writeln!(
        file,
        "{:<30} {:<15}",
        format!("TOTAL ({} jobs)", summary_data.len()),
        total_metrics
    )?;

    println!("[Info] Summary file generated: {}", summary_path.display());

    Ok(())
}

/// Generate .txt file listing metrics for a job
pub fn generate_txt_file<P: AsRef<Path>>(
    output_dir: P,
    job_name: &str,
    metrics: &[String],
) -> Result<()> {
    let output_dir = output_dir.as_ref();
    let filename = format!("{}_unused_metrics.txt", sanitize_filename(job_name));
    let file_path = output_dir.join(&filename);

    let mut file = File::create(&file_path)
        .with_context(|| format!("Failed to create file: {}", file_path.display()))?;

    for metric in metrics {
        writeln!(file, "{}", metric)?;
    }

    println!("[Info] Generated .txt file: {}", file_path.display());

    Ok(())
}

/// Generate combined YAML config file
pub fn generate_yaml_file<P: AsRef<Path>>(
    output_path: P,
    job_metrics_map: &std::collections::HashMap<String, Vec<String>>,
    summary_data: &[JobMetricSummary],
    total_unique_metrics: usize,
) -> Result<usize> {
    let output_path = output_path.as_ref();
    let mut job_configs = Vec::new();
    let mut processed_count = 0;

    for item in summary_data {
        let job_name = &item.job_name;
        let metrics = match job_metrics_map.get(job_name) {
            Some(m) => m,
            None => continue,
        };

        if metrics.is_empty() {
            println!(
                "[Warning] Metric list for job '{}' is empty. Skipping YAML rule generation.",
                job_name
            );
            continue;
        }

        println!("--- Processing (YAML): Job: {} ---", job_name);
        println!("[Info] Found {} metrics", metrics.len());

        let (relabel_rules, group_info) = generate_relabel_rules(metrics);
        println!("[Info] Grouped into {} rule entries", relabel_rules.len());

        print_group_summary(&group_info);

        let total_metrics_in_job: usize = group_info.iter().map(|g| g.count).sum();
        let num_prefix_groups = group_info
            .iter()
            .map(|g| g.prefix.clone())
            .collect::<std::collections::HashSet<_>>()
            .len();

        let num_rules = relabel_rules.len();

        let job_config = JobConfig {
            job_name: job_name.clone(),
            metric_relabel_configs: relabel_rules,
        };

        // Store with comment info
        job_configs.push((
            job_config,
            total_metrics_in_job,
            num_prefix_groups,
            num_rules,
        ));

        processed_count += 1;
    }

    if job_configs.is_empty() {
        println!("\n[Info] No jobs with metrics processed, YAML file not generated.");
        return Ok(0);
    }

    // Write YAML file with comments
    let yaml_content = build_yaml_with_comments(&job_configs, total_unique_metrics)?;

    fs::write(output_path, yaml_content)
        .with_context(|| format!("Failed to write YAML file: {}", output_path.display()))?;

    println!(
        "\n[Success] Combined result saved to '{}'",
        output_path.display()
    );

    Ok(processed_count)
}

/// Build YAML content with summary comments
fn build_yaml_with_comments(
    job_configs: &[(JobConfig, usize, usize, usize)],
    total_unique_metrics: usize,
) -> Result<String> {
    let mut output = String::new();

    for (job_config, total_metrics, num_prefix_groups, num_rules) in job_configs {
        // Add summary comment
        output.push_str(&format!(
            "# Summary: {} of {} unused metrics / {} prefix groups / {} rules generated\n",
            total_metrics, total_unique_metrics, num_prefix_groups, num_rules
        ));

        // Serialize job config to YAML
        let yaml = serde_yaml::to_string(job_config)?;
        output.push_str(&yaml);
        output.push('\n');
    }

    Ok(output)
}

/// Sanitize filename by replacing invalid characters
fn sanitize_filename(name: &str) -> String {
    name.replace(['/', '\\', ':', '*', '?', '"', '<', '>', '|'], "_")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sanitize_filename() {
        assert_eq!(sanitize_filename("normal-name"), "normal-name");
        assert_eq!(
            sanitize_filename("name/with:special*chars"),
            "name_with_special_chars"
        );
    }
}
