use anyhow::Result;
use indicatif::{ProgressBar, ProgressStyle};
use log::{debug, info};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tabled::{Table, Tabled};
use tokio::sync::Semaphore;

const MAX_RETRIES: usize = 3;
const RETRY_INTERVAL: Duration = Duration::from_secs(2);
const REQUEST_TIMEOUT: Duration = Duration::from_secs(2);
const MAX_CONCURRENT_REQUESTS: usize = 10;

#[derive(Debug, Clone, Tabled)]
pub struct CheckResult {
    #[tabled(rename = "URL")]
    pub url: String,
    #[tabled(rename = "TIME")]
    pub duration: String,
    #[tabled(rename = "STATUS")]
    pub status: String,
    #[tabled(rename = "CODE")]
    pub status_code: String,
    #[tabled(rename = "ATTEMPTS")]
    pub attempts: String,
}

impl CheckResult {
    fn new(
        url: String,
        duration: Duration,
        status: String,
        status_code: Option<u16>,
        attempts: usize,
    ) -> Self {
        let duration_str = format!("{}ms", duration.as_millis());
        let status_code_str = status_code
            .map(|code| code.to_string())
            .unwrap_or_else(|| "-".to_string());
        let attempts_str = if status == "OK" {
            attempts.to_string()
        } else {
            format!("{} (failed)", attempts)
        };

        Self {
            url,
            duration: duration_str,
            status,
            status_code: status_code_str,
            attempts: attempts_str,
        }
    }
}

/// Normalize URL by adding https:// if no scheme is present
fn normalize_url(domain_or_url: &str) -> String {
    if domain_or_url.starts_with("http://") || domain_or_url.starts_with("https://") {
        domain_or_url.to_string()
    } else {
        format!("https://{}", domain_or_url)
    }
}

/// Check if HTTP status code is successful (2xx)
fn is_successful_status(status_code: u16) -> bool {
    (200..300).contains(&status_code)
}

/// Perform a single HTTP request with timeout
async fn perform_http_request(url: &str) -> Result<(u16, Duration)> {
    let client = reqwest::Client::builder()
        .timeout(REQUEST_TIMEOUT)
        .build()?;

    debug!("Performing HTTP request to {}", url);
    let start_time = Instant::now();

    let response = client.get(url).send().await?;
    let duration = start_time.elapsed();
    let status_code = response.status().as_u16();

    debug!(
        "HTTP request completed for {}: status={}, duration={:?}",
        url, status_code, duration
    );

    Ok((status_code, duration))
}

/// Perform check with retry logic for a single domain/URL
async fn perform_check(domain_or_url: &str) -> CheckResult {
    let checked_url = normalize_url(domain_or_url);
    let mut last_duration = Duration::from_secs(0);
    let mut status = "FAILED".to_string();
    let mut status_code: Option<u16> = None;

    for attempt in 1..=MAX_RETRIES {
        match perform_http_request(&checked_url).await {
            Ok((code, duration)) => {
                last_duration = duration;
                status_code = Some(code);

                if is_successful_status(code) {
                    status = "OK".to_string();
                    return CheckResult::new(
                        checked_url,
                        last_duration,
                        status,
                        status_code,
                        attempt,
                    );
                }

                status = "UNEXPECTED_CODE".to_string();
            }
            Err(e) => {
                debug!("Request failed for {}: {}", checked_url, e);
                // Keep last_duration and status as is
            }
        }

        // Retry if not the last attempt
        if attempt < MAX_RETRIES {
            tokio::time::sleep(RETRY_INTERVAL).await;
        }
    }

    CheckResult::new(checked_url, last_duration, status, status_code, MAX_RETRIES)
}

/// Run checks for all domains/URLs in parallel
pub async fn run_checks(domains: Vec<String>) -> Result<()> {
    info!("Starting domain checks for {} domains", domains.len());
    let total_start_time = Instant::now();
    let total_checks = domains.len();

    // Create progress bar
    let pb = ProgressBar::new(total_checks as u64);
    pb.set_style(
        ProgressStyle::default_bar()
            .template("{spinner:.green} [{bar:40.cyan/blue}] {pos}/{len} {msg}")
            .unwrap()
            .progress_chars("#>-"),
    );
    pb.set_message("Knock knock... Who's there?");

    // Semaphore to limit concurrent requests
    let semaphore = Arc::new(Semaphore::new(MAX_CONCURRENT_REQUESTS));
    let mut tasks = Vec::new();

    for domain in domains {
        let permit = semaphore.clone().acquire_owned().await.unwrap();
        let pb_clone = pb.clone();

        let task = tokio::spawn(async move {
            let result = perform_check(&domain).await;
            pb_clone.inc(1);
            drop(permit); // Release semaphore
            result
        });

        tasks.push(task);
    }

    // Wait for all tasks to complete
    let mut results = Vec::new();
    for task in tasks {
        results.push(task.await?);
    }

    pb.finish_and_clear();

    // Count successful checks
    let success_count = results.iter().filter(|r| r.status == "OK").count();

    // Print results table
    use tabled::settings::Style;
    let table = Table::new(&results).with(Style::sharp()).to_string();
    println!("{}", table);

    let total_duration = total_start_time.elapsed();

    // Print summary
    println!(
        "\nSummary: {}/{} successful checks in {:.1}s.",
        success_count,
        total_checks,
        total_duration.as_secs_f64()
    );

    info!(
        "Checks completed: {}/{} successful in {:.1}s",
        success_count,
        total_checks,
        total_duration.as_secs_f64()
    );

    Ok(())
}
