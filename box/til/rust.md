# Rust Development Best Practices

Production-tested patterns extracted from Kubernetes operators and container tools in this repository. Focuses on async runtime (Tokio), structured logging (tracing), CLI design (Clap), and containerization best practices. All examples are from real tools: ec2-statuscheck-rebooter, elasticache-backup, podver, promdrop, and filesystem-cleaner.

## Project Structure

```
project/
├── Cargo.toml
├── src/
│   ├── main.rs       # Entry point, logging, signal handling
│   ├── config.rs     # CLI args with Clap
│   ├── error.rs      # Custom errors with thiserror
│   ├── scanner.rs    # Business logic
│   └── backup.rs     # More modules (flat structure)
├── tests/            # Integration tests
└── Makefile
```

**Note**: Avoid `mod.rs` pattern. Use flat file-based modules (`scanner.rs`, `backup.rs`) instead of nested directory modules (`scanner/mod.rs`).

## Essential Crates

```toml
# Async runtime
tokio = { version = "1.40", features = ["full"] }

# CLI parsing
clap = { version = "4.5", features = ["derive", "env"] }

# Error handling
anyhow = "1.0"           # Applications
thiserror = "2.0"        # Libraries

# Logging
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter", "json"] }

# Serialization
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# AWS
aws-config = { version = "1.5", features = ["behavior-version-latest"] }
aws-sdk-ec2 = "1.78"
aws-sdk-s3 = "1.56"

# Utilities
regex = "1.10"
futures = "0.3"
indicatif = "0.17"      # Progress bars
tempfile = "3.10"       # Testing
```

## CLI Pattern

```rust
use clap::Parser;

#[derive(Parser)]
#[command(name = "tool", version, about = "Description")]
pub struct Config {
    #[arg(long, env = "CHECK_INTERVAL", default_value = "300")]
    pub check_interval: u64,

    #[arg(long, env = "DRY_RUN", default_value = "false")]
    pub dry_run: bool,

    #[arg(long, env = "TAGS", value_delimiter = ',')]
    pub tags: Vec<String>,
}
```

## Logging

```rust
// Initialize (JSON for prod, pretty for dev)
let format = std::env::var("LOG_FORMAT").unwrap_or("pretty".into());
match format.as_str() {
    "json" => tracing_subscriber::fmt().json().init(),
    _ => tracing_subscriber::fmt().compact().init(),
}

// Structured logging
info!(cluster_id = %id, bucket = %name, "Backup started");
error!(error = %e, "Operation failed");
```

## Async Patterns

```rust
#[tokio::main]
async fn main() -> Result<()> {
    let config = Config::parse();

    // Graceful shutdown
    tokio::select! {
        result = worker.run() => result?,
        _ = tokio::signal::ctrl_c() => {
            info!("Shutting down");
        }
    }
    Ok(())
}

// Concurrent processing
use futures::stream::{self, StreamExt};
let results = stream::iter(items)
    .map(|item| async move { process(item).await })
    .buffer_unordered(50)
    .collect::<Vec<_>>()
    .await;
```

## Error Handling

```rust
// Custom errors with thiserror
use thiserror::Error;

#[derive(Error, Debug)]
pub enum BackupError {
    #[error("Snapshot not found: {0}")]
    NotFound(String),
    #[error("Operation timed out: {0}")]
    Timeout(String),
}

// Context with anyhow
use anyhow::{Context, Result};
fs::create_dir_all(&dir)
    .with_context(|| format!("Failed to create: {}", dir.display()))?;
```

## Release Optimization

```toml
[package]
edition = "2024"
rust-version = "1.91"

[profile.release]
opt-level = 3           # Performance: 3, Size: "z"
lto = "thin"
codegen-units = 1
strip = true
panic = "abort"
```

## Testing

```rust
// Unit tests
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse() {
        let v = Version::parse("1.8.0").unwrap();
        assert_eq!(v.major, 1);
    }

    #[tokio::test]
    async fn test_async() {
        let result = fetch().await.unwrap();
        assert_eq!(result.len(), 10);
    }
}
```

## Kubernetes Pattern

```rust
// Health endpoints (/healthz, /readyz)
use hyper::{Body, Request, Response, Server};

async fn handler(req: Request<Body>) -> Result<Response<Body>> {
    match req.uri().path() {
        "/healthz" | "/readyz" => Ok(Response::new(Body::from("OK"))),
        _ => Ok(Response::builder().status(404).body(Body::empty())?),
    }
}

// Graceful shutdown
tokio::select! {
    _ = tokio::signal::ctrl_c() => info!("Shutting down"),
    result = worker.run() => result?,
}
```

## Container Build

```dockerfile
FROM rust:1.91-alpine AS builder
WORKDIR /app
COPY . .
RUN cargo build --release --locked

FROM alpine:3.22
RUN adduser -D -u 1000 appuser
COPY --from=builder /app/target/release/app /usr/local/bin/
USER appuser
ENTRYPOINT ["/usr/local/bin/app"]
```

## Common Crates

| Category | Crate | Purpose |
|----------|-------|---------|
| Async | `tokio`, `futures` | Runtime, streams |
| CLI | `clap`, `dialoguer` | Args, prompts |
| Error | `anyhow`, `thiserror` | Handling, custom types |
| Logging | `tracing`, `tracing-subscriber` | Structured logs |
| Serialization | `serde`, `serde_json`, `serde_yaml` | Data formats |
| AWS | `aws-config`, `aws-sdk-*` | Cloud services |
| Text | `regex`, `prettytable-rs` | Processing, tables |
| System | `globset`, `sysinfo`, `tempfile` | Files, info, tests |
| HTTP | `hyper` | Health checks |
| Time | `chrono`, `chrono-tz` | Dates, timezones |

## Makefile

```makefile
build:
	cargo build

release:
	cargo build --release

test:
	cargo test --verbose

fmt:
	cargo fmt

lint:
	cargo clippy -- -D warnings

clean:
	cargo clean

install:
	cargo install --path .
```

## Best Practices

```rust
// Simple fixed delays (avoid complex adaptive backoff)
tokio::time::sleep(Duration::from_secs(30)).await;

// Concurrency limits
stream::iter(items)
    .map(|item| process(item))
    .buffer_unordered(50)
    .collect::<Vec<_>>()
    .await;

// Timeouts for external operations
tokio::time::timeout(Duration::from_secs(30), operation()).await??;
```

## Common Pitfalls

```rust
// Progress bars: Set length before async work
let items = prepare_items();
progress_bar.set_length(items.len());
process_items(items).await;  // Not before this

// Shared state across tasks
let results = Arc::new(Mutex::new(Vec::new()));
stream::iter(items).for_each(|item| {
    let results = Arc::clone(&results);
    async move {
        results.lock().unwrap().push(process(item).await);
    }
}).await;
```

## Production Examples

- **ec2-statuscheck-rebooter** - AWS automation, health endpoints
- **elasticache-backup** - Workflow orchestration, structured logs
- **podver** - Concurrent scanning, progress bars
- **promdrop** - JSON/YAML processing
- **filesystem-cleaner** - Sidecar pattern, glob matching
