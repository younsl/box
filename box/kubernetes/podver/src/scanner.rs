use crate::config::Config;
use crate::types::{NamespaceStats, Pod, PodList, PodVersion, ScanResult, Version};
use anyhow::{Context, Result};
use futures::stream::{self, StreamExt};
use indicatif::{MultiProgress, ProgressBar, ProgressStyle};
use regex::Regex;
use std::collections::HashMap;
use std::fs::File;
use std::io::Write;
use std::process::Stdio;
use std::sync::{Arc, Mutex};
use std::time::Instant;
use tokio::process::Command;
use tracing::{debug, info, warn};

pub struct Scanner {
    config: Config,
    start_time: Instant,
}

impl Scanner {
    pub fn new(config: Config, start_time: Instant) -> Self {
        Self { config, start_time }
    }

    pub async fn scan_pods(&self) -> Result<ScanResult> {
        let result = Arc::new(Mutex::new(ScanResult::new()));

        // Log concurrency settings
        let max_concurrent_namespaces = 2.min(self.config.namespaces.len());
        info!(
            "Concurrency settings: {} namespaces at a time, {} pods per namespace",
            max_concurrent_namespaces, self.config.max_concurrent
        );

        // Calculate max namespace name length for alignment
        let max_ns_len = self
            .config
            .namespaces
            .iter()
            .map(|ns| ns.len())
            .max()
            .unwrap_or(15)
            .max(15); // Minimum 15 characters

        // Create multi-progress for tracking
        let multi_progress = Arc::new(MultiProgress::new());
        let main_pb = multi_progress.add(ProgressBar::new(self.config.namespaces.len() as u64));
        main_pb.set_style(
            ProgressStyle::default_bar()
                .template("{spinner:.green} [{bar:40.cyan/blue}] {pos}/{len} namespaces | {msg}")
                .unwrap()
                .progress_chars("#>-"),
        );
        main_pb.set_message("Scanning... 0s");

        // Spawn a task to update elapsed time in progress bar
        let start_time = self.start_time;
        let pb_clone = main_pb.clone();
        let time_update_handle = tokio::spawn(async move {
            loop {
                tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
                let elapsed = start_time.elapsed().as_secs();
                pb_clone.set_message(format!("Scanning... {}s", elapsed));
            }
        });

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
                let max_ns_len = max_ns_len;

                async move {
                    if config.verbose {
                        debug!("Checking pods in namespace: {}", namespace);
                    }

                    match get_pods(&namespace).await {
                        Ok(pods) => {
                            // Create a placeholder progress bar (will be updated in scan_namespace)
                            let ns_pb = multi_progress.add(ProgressBar::new(0));
                            ns_pb.set_style(
                                ProgressStyle::default_bar()
                                    .template(&format!("  {{spinner:.blue}} {{prefix:{}}}<{{bar:30.yellow/red}}> {{pos}}/{{len}} pods", max_ns_len))
                                    .unwrap()
                                    .progress_chars("█▓▒░ "),
                            );
                            ns_pb.set_prefix(format!("{:width$}", namespace, width = max_ns_len));

                            if let Err(e) = scan_namespace(&config, &namespace, pods, &result, ns_pb.clone()).await
                            {
                                warn!("Error scanning namespace {}: {}", namespace, e);
                                // Even on error, ensure namespace stats are recorded with 0 counts
                                let mut result_lock = result.lock().unwrap();
                                result_lock.namespace_stats.entry(namespace.clone()).or_insert(
                                    NamespaceStats {
                                        total_pods: 0,  // We don't know the actual count if scan failed
                                        jdk_pods: 0,
                                        node_pods: 0,
                                    }
                                );
                            }

                            ns_pb.finish_and_clear();
                            main_pb.inc(1);
                        }
                        Err(e) => {
                            warn!("Error getting pods in namespace {}: {}", namespace, e);
                            // Record namespace even if we couldn't get pods
                            let mut result_lock = result.lock().unwrap();
                            result_lock.namespace_stats.entry(namespace.clone()).or_insert(
                                NamespaceStats {
                                    total_pods: 0,
                                    jdk_pods: 0,
                                    node_pods: 0,
                                }
                            );
                            main_pb.inc(1);
                        }
                    }
                }
            })
            .collect();

        // Wait for all namespace scans to complete
        // Limit concurrent namespace scans to 2 to avoid overwhelming the Kubernetes API
        // and prevent data loss due to rate limiting or resource exhaustion
        stream::iter(tasks)
            .buffer_unordered(max_concurrent_namespaces)
            .collect::<Vec<_>>()
            .await;

        // Stop the time update task
        time_update_handle.abort();

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

        // Print filter information if any filters are active
        if self.config.min_java_version.is_some() || self.config.min_node_version.is_some() {
            println!("Applied filters:");
            if let Some(ref java_min) = self.config.min_java_version {
                println!("  - Java version < {}", java_min);
            }
            if let Some(ref node_min) = self.config.min_node_version {
                println!("  - Node.js version < {}", node_min);
            }
            println!();
        }

        // Calculate column widths dynamically
        let mut max_index_width = "INDEX".len();
        let mut max_namespace_width = "NAMESPACE".len();
        let mut max_pod_width = "POD".len();
        let mut max_java_version_width = "JAVA_VERSION".len();
        let mut max_node_version_width = "NODE_VERSION".len();

        let mut total_entries = 0;
        for (namespace, pods) in &result.pod_versions {
            for (pod_name, version) in pods {
                total_entries += 1;
                let index_str = total_entries.to_string();
                max_index_width = max_index_width.max(index_str.len());
                max_namespace_width = max_namespace_width.max(namespace.len());
                max_pod_width = max_pod_width.max(pod_name.len());
                max_java_version_width = max_java_version_width.max(version.java.len());
                max_node_version_width = max_node_version_width.max(version.node.len());
            }
        }

        // Add padding
        max_index_width += 2;
        max_namespace_width += 2;
        max_pod_width += 2;
        max_java_version_width += 2;
        max_node_version_width += 2;

        // Print header
        println!(
            "{:<width_index$}{:<width_ns$}{:<width_pod$}{:<width_java$}{:<width_node$}",
            "INDEX",
            "NAMESPACE",
            "POD",
            "JAVA_VERSION",
            "NODE_VERSION",
            width_index = max_index_width,
            width_ns = max_namespace_width,
            width_pod = max_pod_width,
            width_java = max_java_version_width,
            width_node = max_node_version_width
        );

        // Print data
        let mut index = 1;
        for (namespace, pods) in &result.pod_versions {
            for (pod_name, version) in pods {
                println!(
                    "{:<width_index$}{:<width_ns$}{:<width_pod$}{:<width_java$}{:<width_node$}",
                    index,
                    namespace,
                    pod_name,
                    version.java,
                    version.node,
                    width_index = max_index_width,
                    width_ns = max_namespace_width,
                    width_pod = max_pod_width,
                    width_java = max_java_version_width,
                    width_node = max_node_version_width
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
        let mut max_node_width = "NODE PODS".len();
        let mut max_jdk_ratio_width = "JDK%".len();
        let mut max_node_ratio_width = "NODE%".len();

        for (ns, stats) in &result.namespace_stats {
            max_ns_width = max_ns_width.max(ns.len());
            max_total_width = max_total_width.max(stats.total_pods.to_string().len());
            max_jdk_width = max_jdk_width.max(stats.jdk_pods.to_string().len());
            max_node_width = max_node_width.max(stats.node_pods.to_string().len());
            // Ratio format: "100.0%" = 6 chars max
            max_jdk_ratio_width = max_jdk_ratio_width.max(6);
            max_node_ratio_width = max_node_ratio_width.max(6);
        }

        max_ns_width += 2;
        max_total_width += 2;
        max_jdk_width += 2;
        max_node_width += 2;
        max_jdk_ratio_width += 2;
        max_node_ratio_width += 2;

        // Print namespace summary header
        println!(
            "{:<ns_width$}{:<total_width$}{:<jdk_width$}{:<node_width$}{:<jdk_ratio_width$}{:<node_ratio_width$}{}",
            "NAMESPACE",
            "TOTAL PODS",
            "JDK PODS",
            "NODE PODS",
            "JDK%",
            "NODE%",
            "TIME",
            ns_width = max_ns_width,
            total_width = max_total_width,
            jdk_width = max_jdk_width,
            node_width = max_node_width,
            jdk_ratio_width = max_jdk_ratio_width,
            node_ratio_width = max_node_ratio_width
        );

        // Sort namespaces alphabetically
        let mut namespaces: Vec<_> = result.namespace_stats.keys().collect();
        namespaces.sort();

        // Print namespace summary data
        for ns in namespaces {
            if let Some(stats) = result.namespace_stats.get(ns) {
                let jdk_ratio = if stats.total_pods > 0 {
                    (stats.jdk_pods as f64 / stats.total_pods as f64) * 100.0
                } else {
                    0.0
                };
                let node_ratio = if stats.total_pods > 0 {
                    (stats.node_pods as f64 / stats.total_pods as f64) * 100.0
                } else {
                    0.0
                };
                println!(
                    "{:<ns_width$}{:<total_width$}{:<jdk_width$}{:<node_width$}{:<jdk_ratio_width$}{:<node_ratio_width$}{}m {}s",
                    ns,
                    stats.total_pods,
                    stats.jdk_pods,
                    stats.node_pods,
                    format!("{:.1}%", jdk_ratio),
                    format!("{:.1}%", node_ratio),
                    elapsed.as_secs() / 60,
                    elapsed.as_secs() % 60,
                    ns_width = max_ns_width,
                    total_width = max_total_width,
                    jdk_width = max_jdk_width,
                    node_width = max_node_width,
                    jdk_ratio_width = max_jdk_ratio_width,
                    node_ratio_width = max_node_ratio_width
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
    writeln!(file, "INDEX,NAMESPACE,POD,JAVA_VERSION,NODE_VERSION")?;

    // Write data rows
    let mut index = 1;
    for (namespace, pods) in &result.pod_versions {
        for (pod_name, version) in pods {
            writeln!(
                file,
                r#"{},{},"{}","{}","{}""#,
                index, namespace, pod_name, version.java, version.node
            )?;
            index += 1;
        }
    }

    // Write namespace summary table
    writeln!(file)?;
    writeln!(file, "# Namespace Summary")?;
    writeln!(file, "NAMESPACE,TOTAL PODS,JDK PODS,NODE PODS,JDK%,NODE%,TIME")?;

    // Sort namespaces alphabetically
    let mut namespaces: Vec<_> = result.namespace_stats.keys().collect();
    namespaces.sort();

    // Write namespace summary data
    for ns in namespaces {
        if let Some(stats) = result.namespace_stats.get(ns) {
            let jdk_ratio = if stats.total_pods > 0 {
                (stats.jdk_pods as f64 / stats.total_pods as f64) * 100.0
            } else {
                0.0
            };
            let node_ratio = if stats.total_pods > 0 {
                (stats.node_pods as f64 / stats.total_pods as f64) * 100.0
            } else {
                0.0
            };
            writeln!(
                file,
                r#"{},"{}","{}","{}","{:.1}%","{:.1}%","{}m {}s""#,
                ns,
                stats.total_pods,
                stats.jdk_pods,
                stats.node_pods,
                jdk_ratio,
                node_ratio,
                elapsed.as_secs() / 60,
                elapsed.as_secs() % 60
            )?;
        }
    }

    // Write overall summary
    writeln!(file)?;
    writeln!(file, "# Overall Summary")?;
    writeln!(file, "# Total pods scanned: {}", result.total_pods)?;
    writeln!(file, "# Pods using JDK: {}", result.jdk_pods)?;
    writeln!(file, "# Pods using Node.js: {}", result.node_pods)?;
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

    // Filter pods based on configuration FIRST (before any async work)
    let pods_to_scan: Vec<_> = pods
        .into_iter()
        .filter(|pod| {
            if config.skip_daemonset && pod.is_daemonset() {
                if config.verbose {
                    debug!("Skipping DaemonSet pod: {}", pod.metadata.name);
                }
                return false;
            }
            true
        })
        .collect();

    // CRITICAL: Set progress bar length BEFORE starting any scans
    // This prevents race condition where inc(1) is called before set_length()
    let pods_to_scan_count = pods_to_scan.len();
    progress_bar.set_length(pods_to_scan_count as u64);

    // Ensure progress bar is updated before continuing
    progress_bar.tick();

    if config.verbose {
        debug!(
            "Scanning {} pods in namespace {} (skipped {} DaemonSet pods)",
            pods_to_scan_count,
            namespace,
            total_pods - pods_to_scan_count
        );
    }

    // Scan pods concurrently with semaphore
    let tasks = pods_to_scan.into_iter().map(|pod| {
        let namespace = namespace.to_string();
        let config = config.clone();
        let pb = progress_bar.clone();

        async move {
            let java_version = get_java_version(&config, &namespace, &pod.metadata.name).await;
            let node_version = get_node_version(&config, &namespace, &pod.metadata.name).await;
            pb.inc(1);
            (pod.metadata.name, PodVersion {
                java: java_version,
                node: node_version,
            })
        }
    });

    let results: Vec<(String, PodVersion)> = stream::iter(tasks)
        .buffer_unordered(config.max_concurrent)
        .collect()
        .await;

    // Parse minimum version filters if provided
    let min_java_version = config
        .min_java_version
        .as_ref()
        .and_then(|v| Version::parse(v));
    let min_node_version = config
        .min_node_version
        .as_ref()
        .and_then(|v| Version::parse(v));

    // Store total scanned count BEFORE filtering
    let total_scanned = results.len();

    // First, filter results locally WITHOUT holding the mutex
    let mut jdk_count = 0;
    let mut node_count = 0;
    let namespace_results: HashMap<String, PodVersion> = results
        .into_iter()
        .filter_map(|(pod_name, version)| {
            let has_java = version.has_java();
            let has_node = version.has_node();

            // Apply version filters
            let mut should_include = false;

            if has_java {
                let include_java = if let Some(ref min_ver) = min_java_version {
                    // Only include if version is below minimum
                    version.java_below_min(min_ver)
                } else {
                    // No filter, include all
                    true
                };

                if include_java {
                    jdk_count += 1;
                    should_include = true;
                }
            }

            if has_node {
                let include_node = if let Some(ref min_ver) = min_node_version {
                    // Only include if version is below minimum
                    version.node_below_min(min_ver)
                } else {
                    // No filter, include all
                    true
                };

                if include_node {
                    node_count += 1;
                    should_include = true;
                }
            }

            if should_include {
                Some((pod_name, version))
            } else {
                None
            }
        })
        .collect();

    // Now update shared result with the mutex locked only once
    let mut result = result.lock().unwrap();
    result.total_pods += total_scanned;
    result.jdk_pods += jdk_count;
    result.node_pods += node_count;

    // Store namespace-specific stats (always store)
    result.namespace_stats.insert(
        namespace.to_string(),
        NamespaceStats {
            total_pods: total_scanned,
            jdk_pods: jdk_count,
            node_pods: node_count,
        },
    );

    // Only store pod versions if there are results to show
    // This prevents empty namespaces from appearing in the pod list table
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

async fn get_node_version(config: &Config, namespace: &str, pod_name: &str) -> String {
    let timeout = config.timeout_duration();

    let result = tokio::time::timeout(
        timeout,
        Command::new("kubectl")
            .args(["exec", "-n", namespace, pod_name, "--", "node", "--version"])
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output(),
    )
    .await;

    match result {
        Ok(Ok(output)) => {
            if output.status.success() {
                let output_str = String::from_utf8_lossy(&output.stdout);
                parse_node_version(&output_str)
            } else {
                if config.verbose {
                    debug!(
                        "Error getting Node.js version for pod {}: command failed",
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
                debug!("Timeout getting Node.js version for pod {}", pod_name);
            }
            "Unknown".to_string()
        }
    }
}

fn parse_node_version(output: &str) -> String {
    // Node.js version format: v18.17.0, v20.5.1, etc.
    let re = Regex::new(r"v(\d+\.\d+\.\d+)").unwrap();

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

    #[test]
    fn test_parse_node_version() {
        let output1 = "v18.17.0\n";
        assert_eq!(parse_node_version(output1), "18.17.0");

        let output2 = "v20.5.1";
        assert_eq!(parse_node_version(output2), "20.5.1");

        let output3 = "v16.20.2\n";
        assert_eq!(parse_node_version(output3), "16.20.2");

        let output4 = "no version here";
        assert_eq!(parse_node_version(output4), "Unknown");
    }
}
