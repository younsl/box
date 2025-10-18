use crate::config::Config;
use crate::types::{NamespaceStats, Pod, PodList, ScanResult};
use anyhow::{Context, Result};
use futures::stream::{self, StreamExt};
use indicatif::{MultiProgress, ProgressBar, ProgressStyle};
use regex::Regex;
use std::collections::HashMap;
use std::fs::File;
use std::io::Write;
use std::process::Stdio;
use std::sync::{Arc, Mutex};
use tokio::process::Command;
use tracing::{debug, info, warn};

pub struct Scanner {
    config: Config,
}

impl Scanner {
    pub fn new(config: Config) -> Self {
        Self { config }
    }

    pub async fn scan_pods(&self) -> Result<ScanResult> {
        let result = Arc::new(Mutex::new(ScanResult::new()));

        // Create multi-progress for tracking
        let multi_progress = Arc::new(MultiProgress::new());
        let main_pb = multi_progress.add(ProgressBar::new(self.config.namespaces.len() as u64));
        main_pb.set_style(
            ProgressStyle::default_bar()
                .template("{spinner:.green} [{bar:40.cyan/blue}] {pos}/{len} namespaces | {msg}")
                .unwrap()
                .progress_chars("#>-"),
        );
        main_pb.set_message("Scanning...");

        // Process namespaces concurrently
        let tasks: Vec<_> = self
            .config
            .namespaces
            .iter()
            .map(|namespace| {
                let namespace = namespace.clone();
                let result = Arc::clone(&result);
                let config = self.config.clone();
                let main_pb = main_pb.clone();
                let multi_progress = Arc::clone(&multi_progress);

                async move {
                    if config.verbose {
                        debug!("Checking pods in namespace: {}", namespace);
                    }

                    match get_pods(&namespace).await {
                        Ok(pods) => {
                            let pod_count = pods.len();

                            // Create namespace-specific progress bar
                            let ns_pb = multi_progress.add(ProgressBar::new(pod_count as u64));
                            ns_pb.set_style(
                                ProgressStyle::default_bar()
                                    .template("  {spinner:.blue} {prefix:20} [{bar:30.yellow/red}] {pos}/{len} pods")
                                    .unwrap()
                                    .progress_chars("█▓▒░ "),
                            );
                            ns_pb.set_prefix(format!("{}", namespace));

                            if let Err(e) = scan_namespace(&config, &namespace, pods, &result, ns_pb.clone()).await
                            {
                                warn!("Error scanning namespace {}: {}", namespace, e);
                            }

                            ns_pb.finish_and_clear();
                            main_pb.inc(1);
                        }
                        Err(e) => {
                            warn!("Error getting pods in namespace {}: {}", namespace, e);
                            main_pb.inc(1);
                        }
                    }
                }
            })
            .collect();

        // Wait for all namespace scans to complete
        stream::iter(tasks)
            .buffer_unordered(self.config.namespaces.len())
            .collect::<Vec<_>>()
            .await;

        main_pb.finish_and_clear();

        // Give a moment for progress bars to fully clear
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;

        let result = Arc::try_unwrap(result)
            .map_err(|_| anyhow::anyhow!("Failed to unwrap result"))?
            .into_inner()
            .map_err(|_| anyhow::anyhow!("Failed to get inner result"))?;

        Ok(result)
    }

    pub fn print_results(&self, result: &ScanResult, elapsed: std::time::Duration) {
        // Clear any remaining terminal artifacts
        println!();

        // Calculate column widths dynamically
        let mut max_index_width = "INDEX".len();
        let mut max_namespace_width = "NAMESPACE".len();
        let mut max_pod_width = "POD".len();
        let mut max_version_width = "JAVA_VERSION".len();

        let mut total_entries = 0;
        for (namespace, pods) in &result.pod_versions {
            for (pod_name, version) in pods {
                total_entries += 1;
                let index_str = total_entries.to_string();
                max_index_width = max_index_width.max(index_str.len());
                max_namespace_width = max_namespace_width.max(namespace.len());
                max_pod_width = max_pod_width.max(pod_name.len());
                max_version_width = max_version_width.max(version.len());
            }
        }

        // Add padding
        max_index_width += 2;
        max_namespace_width += 2;
        max_pod_width += 2;
        max_version_width += 2;

        // Print header
        println!(
            "{:<width_index$}{:<width_ns$}{:<width_pod$}{:<width_ver$}",
            "INDEX",
            "NAMESPACE",
            "POD",
            "JAVA_VERSION",
            width_index = max_index_width,
            width_ns = max_namespace_width,
            width_pod = max_pod_width,
            width_ver = max_version_width
        );

        // Print data
        let mut index = 1;
        for (namespace, pods) in &result.pod_versions {
            for (pod_name, version) in pods {
                println!(
                    "{:<width_index$}{:<width_ns$}{:<width_pod$}{:<width_ver$}",
                    index,
                    namespace,
                    pod_name,
                    version,
                    width_index = max_index_width,
                    width_ns = max_namespace_width,
                    width_pod = max_pod_width,
                    width_ver = max_version_width
                );
                index += 1;
            }
        }

        // Print namespace-level summary (kubectl style)
        println!();

        // Calculate column widths for namespace summary
        let mut max_ns_width = "NAMESPACE".len();
        let mut max_total_width = "TOTAL PODS".len();
        let mut max_jdk_width = "JDK PODS".len();
        let mut max_ratio_width = "RATIO".len();

        for (ns, stats) in &result.namespace_stats {
            max_ns_width = max_ns_width.max(ns.len());
            max_total_width = max_total_width.max(stats.total_pods.to_string().len());
            max_jdk_width = max_jdk_width.max(stats.jdk_pods.to_string().len());
            // Ratio format: "100.0%" = 6 chars max
            max_ratio_width = max_ratio_width.max(6);
        }

        max_ns_width += 2;
        max_total_width += 2;
        max_jdk_width += 2;
        max_ratio_width += 2;

        // Print namespace summary header
        println!(
            "{:<ns_width$}{:<total_width$}{:<jdk_width$}{:<ratio_width$}{}",
            "NAMESPACE",
            "TOTAL PODS",
            "JDK PODS",
            "RATIO",
            "TIME",
            ns_width = max_ns_width,
            total_width = max_total_width,
            jdk_width = max_jdk_width,
            ratio_width = max_ratio_width
        );

        // Sort namespaces alphabetically
        let mut namespaces: Vec<_> = result.namespace_stats.keys().collect();
        namespaces.sort();

        // Print namespace summary data
        for ns in namespaces {
            if let Some(stats) = result.namespace_stats.get(ns) {
                let ratio = if stats.total_pods > 0 {
                    (stats.jdk_pods as f64 / stats.total_pods as f64) * 100.0
                } else {
                    0.0
                };
                println!(
                    "{:<ns_width$}{:<total_width$}{:<jdk_width$}{:<ratio_width$}{}m {}s",
                    ns,
                    stats.total_pods,
                    stats.jdk_pods,
                    format!("{:.1}%", ratio),
                    elapsed.as_secs() / 60,
                    elapsed.as_secs() % 60,
                    ns_width = max_ns_width,
                    total_width = max_total_width,
                    jdk_width = max_jdk_width,
                    ratio_width = max_ratio_width
                );
            }
        }
    }

}

/// Export scan results to CSV file
pub fn export_to_csv(
    result: &ScanResult,
    output_path: &std::path::Path,
    elapsed: std::time::Duration,
) -> Result<()> {
    let mut file = File::create(output_path)
        .with_context(|| format!("Failed to create CSV file: {}", output_path.display()))?;

    // Write CSV header
    writeln!(file, "INDEX,NAMESPACE,POD,JAVA_VERSION")?;

    // Write data rows
    let mut index = 1;
    for (namespace, pods) in &result.pod_versions {
        for (pod_name, version) in pods {
            writeln!(file, "{},{},{},{}", index, namespace, pod_name, version)?;
            index += 1;
        }
    }

    // Write summary as comments
    writeln!(file)?;
    writeln!(file, "# Scan Summary")?;
    writeln!(file, "# Total pods scanned: {}", result.total_pods)?;
    writeln!(file, "# Pods using JDK: {}", result.jdk_pods)?;
    writeln!(
        file,
        "# Time taken: {}m {}s",
        elapsed.as_secs() / 60,
        elapsed.as_secs() % 60
    )?;

    info!("CSV file exported to: {}", output_path.display());
    println!("CSV file saved to: {}", output_path.display());

    Ok(())
}

async fn get_pods(namespace: &str) -> Result<Vec<Pod>> {
    let output = Command::new("kubectl")
        .args(["get", "pods", "-n", namespace, "-o", "json"])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .await
        .context("Failed to execute kubectl command")?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        anyhow::bail!("kubectl command failed: {}", stderr);
    }

    let pod_list: PodList = serde_json::from_slice(&output.stdout)
        .context("Failed to parse kubectl output")?;

    Ok(pod_list.items)
}

async fn scan_namespace(
    config: &Config,
    namespace: &str,
    pods: Vec<Pod>,
    result: &Arc<Mutex<ScanResult>>,
    progress_bar: ProgressBar,
) -> Result<()> {
    let total_pods = pods.len();

    // Filter pods based on configuration
    let pods_to_scan: Vec<_> = pods
        .into_iter()
        .enumerate()
        .filter(|(i, pod)| {
            if config.verbose {
                debug!(
                    "Scanning pod {} of {}: {}",
                    i + 1,
                    total_pods,
                    pod.metadata.name
                );
            }

            if config.skip_daemonset && pod.is_daemonset() {
                if config.verbose {
                    debug!("Skipping DaemonSet pod: {}", pod.metadata.name);
                }
                progress_bar.inc(1);
                return false;
            }

            true
        })
        .map(|(_, pod)| pod)
        .collect();

    // Scan pods concurrently with semaphore
    let tasks = pods_to_scan.into_iter().map(|pod| {
        let namespace = namespace.to_string();
        let config = config.clone();
        let pb = progress_bar.clone();

        async move {
            let version = get_java_version(&config, &namespace, &pod.metadata.name).await;
            pb.inc(1);
            (pod.metadata.name, version)
        }
    });

    let results: Vec<(String, String)> = stream::iter(tasks)
        .buffer_unordered(config.max_concurrent)
        .collect()
        .await;

    // Update shared result
    let mut result = result.lock().unwrap();
    let scanned_count = results.len();
    result.total_pods += scanned_count;

    let mut jdk_count = 0;
    let namespace_results: HashMap<String, String> = results
        .into_iter()
        .filter(|(_, version)| {
            if version != "Unknown" {
                result.jdk_pods += 1;
                jdk_count += 1;
                true
            } else {
                false
            }
        })
        .collect();

    // Store namespace-specific stats
    result.namespace_stats.insert(
        namespace.to_string(),
        NamespaceStats {
            total_pods: scanned_count,
            jdk_pods: jdk_count,
        },
    );

    if !namespace_results.is_empty() {
        result
            .pod_versions
            .insert(namespace.to_string(), namespace_results);
    }

    Ok(())
}

async fn get_java_version(config: &Config, namespace: &str, pod_name: &str) -> String {
    let timeout = config.timeout_duration();

    let result = tokio::time::timeout(
        timeout,
        Command::new("kubectl")
            .args(["exec", "-n", namespace, pod_name, "--", "java", "-version"])
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output(),
    )
    .await;

    match result {
        Ok(Ok(output)) => {
            if output.status.success() || !output.stderr.is_empty() {
                // java -version outputs to stderr
                let output_str = String::from_utf8_lossy(&output.stderr);
                parse_java_version(&output_str)
            } else {
                if config.verbose {
                    debug!(
                        "Error getting Java version for pod {}: command failed",
                        pod_name
                    );
                }
                "Unknown".to_string()
            }
        }
        Ok(Err(e)) => {
            if config.verbose {
                debug!("Error executing command for pod {}: {}", pod_name, e);
            }
            "Unknown".to_string()
        }
        Err(_) => {
            if config.verbose {
                debug!("Timeout getting Java version for pod {}", pod_name);
            }
            "Unknown".to_string()
        }
    }
}

fn parse_java_version(output: &str) -> String {
    let re = Regex::new(r#"version "([^"]+)""#).unwrap();

    if let Some(captures) = re.captures(output) {
        if let Some(version) = captures.get(1) {
            return version.as_str().to_string();
        }
    }

    "Unknown".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_java_version() {
        let output1 = r#"openjdk version "11.0.16" 2022-07-19"#;
        assert_eq!(parse_java_version(output1), "11.0.16");

        let output2 = r#"java version "1.8.0_292""#;
        assert_eq!(parse_java_version(output2), "1.8.0_292");

        let output3 = r#"openjdk version "17.0.2" 2022-01-18"#;
        assert_eq!(parse_java_version(output3), "17.0.2");

        let output4 = "no version here";
        assert_eq!(parse_java_version(output4), "Unknown");
    }
}
