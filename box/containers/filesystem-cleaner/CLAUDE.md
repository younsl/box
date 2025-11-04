# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A lightweight Rust-based container designed for automatic filesystem cleanup in Kubernetes environments. Operates as either a sidecar or init container to monitor disk usage and intelligently remove files to prevent storage exhaustion.

**Primary Use Case**: GitHub Actions self-hosted runners where build artifacts and cache data accumulate in the shared workspace volume (`/home/runner/_work`).

**Language Migration**: Ported from Go to Rust for performance, memory safety, and modern tooling benefits.

## Development Commands

### Build Commands

```bash
# Debug build (target/debug/)
make build

# Optimized release build (target/release/)
make release

# Multi-platform build (requires cross-rs)
make build-all

# Quick test run
make run         # Dry-run cleanup of /tmp at 70% threshold

# Debug logging mode
make dev         # Same as run but with --log-level debug
```

### Testing Commands

```bash
# Run all tests
make test
cargo test --verbose

# Run specific test
cargo test test_cleaner_creation --verbose

# Run tests in specific module
cargo test --verbose cleaner::tests
```

### Code Quality

```bash
make fmt         # Format with rustfmt
make check       # Type check without building
make lint        # Run clippy with -D warnings
make deps        # Update dependencies (cargo update)
make clean       # Remove build artifacts
```

### Container Operations

```bash
make docker-build    # Build container image with version info
make docker-push     # Push to ECR (update ECR_REGISTRY first)

# Manual container build with custom version
docker build \
  --build-arg VERSION=0.1.0 \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE="$(date -u '+%Y-%m-%d %H:%M:%S UTC')" \
  -t filesystem-cleaner:0.1.0 .
```

### Installation

```bash
make install     # Install to ~/.cargo/bin/
make uninstall   # Remove from ~/.cargo/bin/
```

## Architecture

### Module Structure

```
src/
├── main.rs      - CLI entry point, logging setup, signal handling
├── config.rs    - Clap argument parsing, CleanupMode enum
└── cleaner.rs   - Core cleanup logic, disk usage monitoring
```

### Key Components

**config.rs**:
- `Args` struct with Clap derive macros for CLI and environment variable parsing
- `CleanupMode` enum: `Once` (init container) vs `Interval` (sidecar)
- Build version info using `option_env!("VERGEN_GIT_SHA")` and `option_env!("VERGEN_BUILD_TIMESTAMP")`

**cleaner.rs**:
- `Cleaner` struct with GlobSet matchers for include/exclude patterns
- `perform_cleanup()`: Checks disk usage threshold, triggers cleanup if exceeded
- `collect_files()`: Recursive directory walk respecting glob patterns
- Atomic shutdown flag for graceful termination
- Structured logging with tracing crate

**main.rs**:
- Tokio async runtime
- Signal handling with `tokio::signal::ctrl_c()`
- Structured logging configuration (compact format, no file/line numbers)

### Disk Usage Detection

Uses `sysinfo::Disks` to find the mount point containing the target path:
1. Enumerates all disk mount points
2. Finds longest matching mount point prefix for target path
3. Calculates percentage: `(total - available) / total * 100`

### Cleanup Logic

**Threshold-based triggering**:
- Only cleans when disk usage > `usage_threshold_percent`
- Logs and skips if below threshold

**File collection**:
- Recursive directory traversal
- Applies exclude patterns to both files and directories
- Applies include patterns only to files
- Default excludes: `.git`, `node_modules`, `*.log`

**Deletion strategy**:
- Deletes files matching include patterns and not matching exclude patterns
- Logs each deletion with size in KB
- Tracks total freed space and reports in MB
- Dry-run mode available for testing

## Kubernetes Deployment Patterns

### Critical Requirements

1. **Volume Sharing**: Both filesystem-cleaner and target container (e.g., actions-runner) must mount the same volume at the same path
2. **Security Context**: Must run as same UID/GID as target container (typically `runAsUser: 1001`, `runAsGroup: 1001`)
3. **Non-root**: Set `runAsNonRoot: true` for security

### Init Container (Once Mode)

Cleans workspace once before main container starts:

```yaml
initContainers:
- name: filesystem-cleaner
  image: ghcr.io/younsl/filesystem-cleaner:0.1.0
  args:
  - "--target-paths=/home/runner/_work"
  - "--usage-threshold-percent=70"
  - "--cleanup-mode=once"
  securityContext:
    runAsUser: 1001
    runAsGroup: 1001
    runAsNonRoot: true
  volumeMounts:
  - name: workspace
    mountPath: /home/runner/_work
```

### Sidecar Container (Interval Mode)

Runs alongside main container for periodic cleanup:

```yaml
containers:
- name: filesystem-cleaner
  image: ghcr.io/younsl/filesystem-cleaner:0.1.0
  args:
  - "--target-paths=/home/runner/_work"
  - "--usage-threshold-percent=80"
  - "--cleanup-mode=interval"
  - "--check-interval-minutes=10"
  securityContext:
    runAsUser: 1001
    runAsGroup: 1001
    runAsNonRoot: true
  volumeMounts:
  - name: workspace
    mountPath: /home/runner/_work
```

## Configuration

All CLI flags can be set via environment variables (uppercase with underscores):

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--target-paths` | `TARGET_PATHS` | `/home/runner/_work` | Comma-separated paths to monitor |
| `--usage-threshold-percent` | `USAGE_THRESHOLD_PERCENT` | `80` | Trigger cleanup at this % (0-100) |
| `--cleanup-mode` | `CLEANUP_MODE` | `interval` | `once` or `interval` |
| `--check-interval-minutes` | `CHECK_INTERVAL_MINUTES` | `10` | Check interval for interval mode |
| `--include-patterns` | `INCLUDE_PATTERNS` | `*` | Comma-separated glob patterns to include |
| `--exclude-patterns` | `EXCLUDE_PATTERNS` | `.git,node_modules,*.log` | Comma-separated glob patterns to exclude |
| `--dry-run` | `DRY_RUN` | `false` | Preview mode without deletion |
| `--log-level` | `LOG_LEVEL` | `info` | `trace`, `debug`, `info`, `warn`, `error` |

## Testing Guidelines

**Current Coverage**:
- Unit tests in `config.rs`: CleanupMode parsing and display
- Unit tests in `cleaner.rs`: Cleaner creation, pattern matching, file collection
- Integration test: `test_collect_files` with tempfile for real filesystem operations

**When Adding Tests**:
- Place unit tests in `#[cfg(test)] mod tests` within same file
- Use `tempfile::TempDir` for filesystem tests
- Use `#[tokio::test]` for async tests
- Test glob pattern matching edge cases
- Mock disk usage checks if possible

## Version Information

Build version info is injected via environment variables during compilation:
- `VERGEN_GIT_SHA`: Git commit hash
- `VERGEN_BUILD_TIMESTAMP`: Build timestamp

Docker builds pass these via `--build-arg`:
```dockerfile
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN VERGEN_GIT_SHA=${GIT_COMMIT} VERGEN_BUILD_TIMESTAMP="${BUILD_DATE}" \
    cargo build --release --locked
```

Accessible via `--version` flag (shows short version) or `--long-version` (shows commit and build date).

## Docker Image Optimization

Multi-stage build with aggressive size reduction:
1. **Builder stage** (rust:1.91-alpine3.22):
   - Dependency caching via dummy main.rs
   - Release build with `strip = true`, `lto = true`, `opt-level = 3`
   - `strip` command to remove debug symbols
   - `upx --best --lzma` for binary compression

2. **Runtime stage** (alpine:3.22):
   - Minimal base (no ca-certificates or tzdata - not needed)
   - Non-root user (UID/GID 1000)
   - Single compressed binary

Result: Very small final image size.
