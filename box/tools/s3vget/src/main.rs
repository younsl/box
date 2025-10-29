use anyhow::{Context, Result};
use aws_sdk_s3::Client;
use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use chrono_tz::Tz;
use clap::Parser;
use dialoguer::Input;
use std::path::PathBuf;
use std::str::FromStr;
use tracing::{debug, error, info, warn};

/// Download all versions of S3 objects with interactive prompts
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// S3 bucket name
    #[arg(short, long)]
    bucket: Option<String>,

    /// S3 object key (path)
    #[arg(short, long)]
    key: Option<String>,

    /// Output directory
    #[arg(short, long, default_value = "versions")]
    output_dir: PathBuf,

    /// Start date (YYYY-MM-DD or 'now')
    #[arg(short, long)]
    start: Option<String>,

    /// End date (YYYY-MM-DD or 'now')
    #[arg(short, long)]
    end: Option<String>,

    /// Skip interactive prompts
    #[arg(long)]
    no_interactive: bool,

    /// Log level (trace, debug, info, warn, error)
    #[arg(long, default_value = "info")]
    log_level: String,

    /// Timezone for date interpretation (e.g., Asia/Seoul, America/New_York, UTC)
    #[arg(short = 'z', long, default_value = "Asia/Seoul")]
    timezone: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Initialize tracing subscriber
    init_tracing(&args.log_level)?;

    info!("Starting s3vget - S3 Object Version Downloader");
    debug!("Parsed arguments: {:?}", args);

    // Parse timezone
    let timezone = Tz::from_str(&args.timezone).with_context(|| {
        format!(
            "Invalid timezone: {}. Use format like 'Asia/Seoul', 'America/New_York', or 'UTC'",
            args.timezone
        )
    })?;
    info!("Using timezone: {}", timezone);

    // Get bucket name (interactive or from args)
    let bucket_name = get_bucket_name(&args)?;
    let object_key = get_object_key(&args)?;

    // Get date range (interactive or from args)
    let (start_date, end_date) = get_date_range(&args, timezone)?;

    // Initialize AWS S3 client
    info!("Initializing AWS S3 client");
    let config = aws_config::load_defaults(aws_config::BehaviorVersion::latest()).await;
    let client = Client::new(&config);

    // Create output directory
    std::fs::create_dir_all(&args.output_dir).context("Failed to create output directory")?;
    info!("Output directory: {}", args.output_dir.display());

    // Download versions
    download_versions(
        &client,
        &bucket_name,
        &object_key,
        &args.output_dir,
        start_date,
        end_date,
        timezone,
    )
    .await?;

    info!("Download complete!");
    Ok(())
}

fn init_tracing(log_level: &str) -> Result<()> {
    use tracing_subscriber::{EnvFilter, fmt};

    let filter = EnvFilter::try_from_default_env()
        .or_else(|_| EnvFilter::try_new(log_level))
        .context("Failed to initialize log filter")?;

    fmt()
        .with_env_filter(filter)
        .with_target(false)
        .with_thread_ids(false)
        .with_file(false)
        .with_line_number(false)
        .init();

    Ok(())
}

fn get_bucket_name(args: &Args) -> Result<String> {
    if let Some(bucket) = &args.bucket {
        debug!("Using bucket from arguments: {}", bucket);
        return Ok(bucket.clone());
    }

    if args.no_interactive {
        anyhow::bail!("Bucket name is required when --no-interactive is set");
    }

    info!("Interactive mode: Enter S3 bucket name");
    let bucket: String = Input::new().with_prompt("S3 bucket name").interact_text()?;

    if bucket.trim().is_empty() {
        anyhow::bail!("Bucket name cannot be empty");
    }

    Ok(bucket)
}

fn get_object_key(args: &Args) -> Result<String> {
    if let Some(key) = &args.key {
        debug!("Using object key from arguments: {}", key);
        return Ok(key.clone());
    }

    if args.no_interactive {
        anyhow::bail!("Object key is required when --no-interactive is set");
    }

    info!("Interactive mode: Enter S3 object key");
    let key: String = Input::new()
        .with_prompt("S3 object key (path)")
        .interact_text()?;

    if key.trim().is_empty() {
        anyhow::bail!("Object key cannot be empty");
    }

    Ok(key)
}

fn get_date_range(
    args: &Args,
    timezone: Tz,
) -> Result<(Option<DateTime<Utc>>, Option<DateTime<Utc>>)> {
    // If dates provided via arguments
    if args.start.is_some() && args.end.is_some() {
        let start = parse_date_input(args.start.as_ref().unwrap(), true, timezone)?;
        let end = parse_date_input(args.end.as_ref().unwrap(), false, timezone)?;

        info!("Using date range from arguments");
        info!("Start: {} ({})", start.with_timezone(&timezone), timezone);
        info!("End: {} ({})", end.with_timezone(&timezone), timezone);

        return Ok((Some(start), Some(end)));
    }

    if args.no_interactive {
        info!("No date filter applied (downloading all versions)");
        return Ok((None, None));
    }

    // Interactive mode
    info!("Interactive mode: Enter date range (or press Enter to skip filtering)");
    println!("\n{}", "=".repeat(60));
    println!("Date Range Selection ({})", timezone);
    println!("{}", "=".repeat(60));
    println!("Enter dates in format: YYYY-MM-DD or 'now'");
    println!("Press Enter to skip date filtering\n");

    let start_input: String = Input::new()
        .with_prompt("Start date")
        .allow_empty(true)
        .interact_text()?;

    if start_input.trim().is_empty() {
        info!("No date filter applied (downloading all versions)");
        return Ok((None, None));
    }

    let end_input: String = Input::new().with_prompt("End date").interact_text()?;

    let start = parse_date_input(&start_input, true, timezone)?;
    let end = parse_date_input(&end_input, false, timezone)?;

    if start > end {
        anyhow::bail!("Start date must be before or equal to end date");
    }

    info!(
        "Date range: {} to {} ({})",
        start.with_timezone(&timezone).format("%Y-%m-%d"),
        end.with_timezone(&timezone).format("%Y-%m-%d"),
        timezone
    );

    Ok((Some(start), Some(end)))
}

fn parse_date_input(input: &str, is_start: bool, timezone: Tz) -> Result<DateTime<Utc>> {
    let input = input.trim();

    // Handle 'now'
    if input.eq_ignore_ascii_case("now") {
        debug!("Parsing 'now' as current time");
        return Ok(Utc::now());
    }

    // Try different date formats
    let formats = ["%Y-%m-%d", "%Y/%m/%d", "%Y.%m.%d", "%Y%m%d"];

    for format in &formats {
        if let Ok(naive_date) = NaiveDate::parse_from_str(input, format) {
            debug!("Parsed date '{}' with format '{}'", input, format);

            // Convert to specified timezone
            let tz_datetime = if is_start {
                // Start of day (00:00:00)
                timezone
                    .from_local_datetime(&naive_date.and_hms_opt(0, 0, 0).unwrap())
                    .single()
                    .context("Failed to convert to specified timezone")?
            } else {
                // End of day (23:59:59)
                timezone
                    .from_local_datetime(&naive_date.and_hms_opt(23, 59, 59).unwrap())
                    .single()
                    .context("Failed to convert to specified timezone")?
            };

            return Ok(tz_datetime.with_timezone(&Utc));
        }
    }

    anyhow::bail!(
        "Unable to parse date: '{}'. Use format: YYYY-MM-DD or 'now'",
        input
    )
}

async fn download_versions(
    client: &Client,
    bucket_name: &str,
    object_key: &str,
    output_dir: &PathBuf,
    start_date: Option<DateTime<Utc>>,
    end_date: Option<DateTime<Utc>>,
    timezone: Tz,
) -> Result<()> {
    info!("Fetching versions for s3://{}/{}", bucket_name, object_key);

    // List all versions
    let mut versions = Vec::new();
    let mut continuation_token: Option<String> = None;

    loop {
        let mut request = client
            .list_object_versions()
            .bucket(bucket_name)
            .prefix(object_key);

        if let Some(token) = continuation_token {
            request = request.key_marker(token);
        }

        let response = request
            .send()
            .await
            .context("Failed to list object versions")?;

        for version in response.versions() {
            if version.key() == Some(object_key) {
                versions.push(version.clone());
            }
        }

        if response.is_truncated() == Some(true) {
            continuation_token = response.next_key_marker().map(|s| s.to_string());
            debug!("Fetching next page of versions");
        } else {
            break;
        }
    }

    if versions.is_empty() {
        warn!("No versions found for {}", object_key);
        return Ok(());
    }

    info!("Found {} total version(s)", versions.len());

    // Sort versions by LastModified (oldest first)
    versions.sort_by(|a, b| {
        a.last_modified()
            .unwrap_or(&aws_sdk_s3::primitives::DateTime::from_secs(0))
            .cmp(
                b.last_modified()
                    .unwrap_or(&aws_sdk_s3::primitives::DateTime::from_secs(0)),
            )
    });

    // Filter by date range if specified
    let filtered_versions = if let (Some(start), Some(end)) = (start_date, end_date) {
        let filtered: Vec<_> = versions
            .into_iter()
            .filter(|v| {
                if let Some(last_modified) = v.last_modified() {
                    let version_time =
                        DateTime::from_timestamp(last_modified.secs(), 0).unwrap_or(Utc::now());
                    version_time >= start && version_time <= end
                } else {
                    false
                }
            })
            .collect();

        info!(
            "Filtered to {} version(s) within date range",
            filtered.len()
        );
        filtered
    } else {
        versions
    };

    if filtered_versions.is_empty() {
        warn!("No versions to download after filtering");
        return Ok(());
    }

    // Extract filename components
    let filename = object_key.split('/').last().unwrap_or(object_key);
    let (name, ext) = filename.rsplit_once('.').unwrap_or((filename, ""));

    println!("\n{}", "=".repeat(60));
    println!("Downloading {} version(s)", filtered_versions.len());
    println!("{}", "=".repeat(60));

    // Download each version
    for (idx, version) in filtered_versions.iter().enumerate() {
        let version_id = version.version_id().unwrap_or("unknown");
        let last_modified = version.last_modified().context("Missing last_modified")?;

        // Convert to specified timezone
        let tz_time = DateTime::from_timestamp(last_modified.secs(), 0)
            .unwrap_or(Utc::now())
            .with_timezone(&timezone);

        let timestamp = tz_time.format("%Y%m%d_%H%M%S");
        let local_filename = if ext.is_empty() {
            format!("{:03}_{}_{}", idx + 1, timestamp, name)
        } else {
            format!("{:03}_{}_{}.{}", idx + 1, timestamp, name, ext)
        };
        let local_path = output_dir.join(&local_filename);

        debug!(
            "Downloading version {}/{}: {}",
            idx + 1,
            filtered_versions.len(),
            version_id
        );

        // Download the version
        match download_version(client, bucket_name, object_key, version_id, &local_path).await {
            Ok(size) => {
                info!(
                    "[{:03}/{:03}] Downloaded: {}",
                    idx + 1,
                    filtered_versions.len(),
                    local_filename
                );
                println!("           Version ID: {}", version_id);
                println!(
                    "           Timestamp: {} {}",
                    tz_time.format("%Y-%m-%d %H:%M:%S"),
                    timezone
                );
                println!("           Size: {:.2} KB\n", size as f64 / 1024.0);
            }
            Err(e) => {
                error!("Failed to download version {}: {:?}", version_id, e);
                println!("[ERROR] Failed to download version {}: {}\n", version_id, e);
            }
        }
    }

    println!("{}", "=".repeat(60));
    println!(
        "✓ Download complete! All versions saved to: {}",
        output_dir.display()
    );
    println!("{}", "=".repeat(60));

    Ok(())
}

async fn download_version(
    client: &Client,
    bucket_name: &str,
    object_key: &str,
    version_id: &str,
    local_path: &PathBuf,
) -> Result<u64> {
    let response = client
        .get_object()
        .bucket(bucket_name)
        .key(object_key)
        .version_id(version_id)
        .send()
        .await
        .context("Failed to get object")?;

    let data = response
        .body
        .collect()
        .await
        .context("Failed to read object body")?;

    let bytes = data.into_bytes();
    let size = bytes.len() as u64;

    std::fs::write(local_path, bytes).context("Failed to write file")?;

    Ok(size)
}
