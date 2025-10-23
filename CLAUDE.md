# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

A monorepo serving as a DevOps toolbox containing Kubernetes utilities, automation scripts, infrastructure code, and engineering documentation.

All applications in `kubernetes/` and `tools/` are built with **[Rust](https://github.com/rust-lang/rust) 1.90+** (except `cocd` which uses Go). Rust provides key operational benefits: minimal container sizes, low memory footprint, single static binaries with no runtime dependencies, memory safety preventing null pointer and buffer overflow crashes, and compile-time guarantees ensuring system stability in production.

## Development Commands

### Go Projects

Standard Makefile patterns for Go tools (cocd):

```bash
# Core build commands
make build          # Build binary
make build-all      # Build for all platforms (linux/darwin, amd64/arm64, windows)
make run            # Run application
make install        # Install to system

# Testing
make test           # Run tests (go test -v ./...)
go test -v ./...                    # Run all tests
go test -v ./pkg/specific/package   # Run tests for specific package
go test -v -run TestFunctionName    # Run specific test function

# Code quality
make fmt            # Format code
make lint           # golangci-lint (if installed)
make mod            # go mod tidy + vendor
make clean          # Remove build artifacts
```

**Note**: cocd uses `make mod` instead of `make deps` for module management.

### Rust Projects

Standard Makefile patterns for Rust tools (kk, qg, podver, promdrop, filesystem-cleaner, elasticache-backup):

```bash
# Core build commands
make build          # Build debug binary (target/debug/)
make release        # Build optimized release binary (target/release/)
make build-all      # Build for all platforms (requires cross)

# Development workflow
make run            # Build and run with example
make dev            # Run with verbose/debug logging
make install        # Install to ~/.cargo/bin/
make test           # Run tests (cargo test --verbose)

# Code quality
make fmt            # Format code (cargo fmt)
make lint           # Run clippy (cargo clippy -- -D warnings)
make check          # Check code without building
make deps           # Update dependencies (cargo update)
make clean          # Remove build artifacts

# Direct cargo commands for specific tests
cargo test --verbose                    # Run all tests
cargo test --verbose test_name          # Run specific test
cargo test --package package_name       # Run tests for specific package
```

### Container Operations

```bash
make docker-build   # Build Docker image
make docker-push    # Push to ECR (requires AWS credentials)
make deploy         # Deploy to Kubernetes (where available)
```

**Container-Specific Notes**:
- **filesystem-cleaner**: Includes `make all` target that runs fmt + lint + test + build
- **elasticache-backup**: Supports multi-arch builds and has `run-json` target for JSON log testing
- **actions-runner**, **hugo**: Have specialized build workflows for multi-arch images
- Update ECR_REGISTRY variable in Makefiles before pushing

### Terraform Projects

Standard Terraform workflow for infrastructure modules:

```bash
terraform init      # Initialize Terraform
terraform validate  # Validate configuration
terraform plan      # Plan changes
terraform apply     # Apply changes
terraform destroy   # Destroy resources
```

**Available Modules**:
- `vault/irsa/` - Vault auto-unseal with AWS KMS integration

## High-Level Architecture

### Repository Structure

```
box/
├── kubernetes/             # K8s controllers, policies, helm charts
│   ├── podver/            # Pod Version Scanner (Rust, container)
│   ├── promdrop/          # Prometheus metric filter generator (Rust, CLI + container)
│   ├── elasticache-backup/# ElastiCache S3 backup automation (Rust, container)
│   └── policies/          # Kyverno and CEL admission policies
├── tools/                 # CLI utilities
│   ├── cocd/              # GitHub Actions deployment monitor (Go, TUI)
│   ├── kk/                # Domain connectivity checker (Rust)
│   └── qg/                # QR code generator (Rust)
├── containers/            # Custom container images
│   ├── actions-runner/    # GitHub Actions runner
│   ├── filesystem-cleaner/# File system cleanup tool (Rust)
│   ├── hugo/              # Hugo static site generator image
│   ├── ab/                # Apache Bench container
│   ├── mageai/            # Mage AI custom image
│   ├── yarn/              # Yarn package manager container
│   └── terraform-console-machine/  # Terraform console container
├── scripts/               # Automation scripts by platform
│   ├── aws/               # AWS resource management
│   ├── github/            # Repository automation
│   └── k8s-registry-io-stat/  # K8s connectivity testing
├── terraform/             # Infrastructure as Code
│   └── vault/irsa/        # Vault auto-unseal with AWS KMS
├── actions/               # GitHub Actions reusable workflows
└── til/                   # Engineering notes and learnings
```

### Architectural Patterns

**Kubernetes Applications**:
- DaemonSet pattern for node-level operations
- IMDS access via host network when required
- IRSA for AWS API authentication
- Health endpoints on port 8080
- Graceful shutdown handling with signal handling

**Go Application Structure** (cocd):
- `cmd/` - Application entry points
- `pkg/` or `internal/` - Reusable packages
- Version embedding via ldflags: `-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"`
- Environment-based configuration
- Structured logging (logrus/zap)

**Rust Application Structure**:
- `src/main.rs` - CLI entry point with Clap argument parsing
- `src/lib.rs` - Core library code (if applicable)
- `src/*.rs` - Module files for specific functionality
- `Cargo.toml` - Rust dependencies and metadata
- Tokio async runtime for concurrent operations
- Structured logging with tracing crate
- Environment variable support via Clap

**CI/CD Pipeline**:
- GitHub Actions for automated releases
- Multi-arch builds (linux/darwin, amd64/arm64)
- Automated binary releases with tags
- Container image push to GHCR (GitHub Container Registry) and ECR

**Rust Cross-Compilation Requirements**:

Rust cross-compilation is more complex than Go and requires additional setup:

```yaml
# For ARM64 cross-compilation on x86_64 (GitHub Actions example)
- name: Install cross-compilation tools (Linux ARM64)
  if: matrix.target == 'aarch64-unknown-linux-gnu'
  run: |
    sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu

- name: Configure cross-compilation (Linux ARM64)
  if: matrix.target == 'aarch64-unknown-linux-gnu'
  run: |
    echo "CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc" >> $GITHUB_ENV
    echo "CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc" >> $GITHUB_ENV
    echo "CXX_aarch64_unknown_linux_gnu=aarch64-linux-gnu-g++" >> $GITHUB_ENV
```

**Why Rust cross-compilation is more complex than Go**:
- Go: Self-contained compiler with built-in cross-compilation (`GOOS=linux GOARCH=arm64 go build`)
- Rust: Requires system linker and C toolchain for target architecture
- Rust crates often depend on C libraries (openssl, sqlite, etc.)
- Must install target-specific gcc/g++ and configure linker paths

See `.github/workflows/release-promdrop.yml` for complete ARM64 cross-compilation example.

## AWS Integration Points

- **ECR**: Container registry for Kubernetes deployments
- **GHCR**: GitHub Container Registry for public container images
- **IAM/IRSA**: Service account to IAM role mapping for Kubernetes pods
- **KMS**: Vault auto-unseal encryption
- **ElastiCache**: Snapshot and backup management
- **S3**: Backup storage and lifecycle management

Configure AWS credentials via environment variables or IAM instance profiles.

## Tool-Specific Notes

### cocd - GitHub Actions Monitor

A TUI (Terminal User Interface) application for monitoring GitHub Actions jobs waiting for approval, inspired by k9s.

```bash
# Environment configuration
export COCD_GITHUB_TOKEN="ghp_..."
export COCD_GITHUB_ORG="your-org"
export COCD_CONFIG_PATH="./config.yaml"

# Authentication hierarchy (first available wins):
# 1. Config file: github.token field
# 2. Environment: GITHUB_TOKEN or COCD_GITHUB_TOKEN
# 3. GitHub CLI: gh auth token

# Repository scanning limitation
# ⚠️ No org-level workflow API exists
# Must iterate repositories individually
```

**Features**:
- Monitor jobs waiting for approval
- View recent workflow runs
- Approve/cancel jobs from TUI
- Real-time updates with configurable refresh

### kk - Domain Connectivity Checker (Rust)

```bash
# Check domain connectivity
./target/release/kk --config configs/domain-example.yaml

# Or use Makefile
make run        # Build and run with example config
make dev        # Run with verbose logging

# Build commands
make build      # Debug build
make release    # Optimized release build
make install    # Install to ~/.cargo/bin/

# Configuration format (YAML):
domains:
  - www.google.com        # Auto-adds https://
  - reddit.com
  - https://registry.k8s.io/v2/
```

**Note**: Uses Tokio for async concurrency and Clap for CLI.

### qg - QR Code Generator (Rust)

```bash
# Generate QR code from URL
./target/release/qg https://github.com/

# Or use Makefile
make run        # Build and run with example URL

# Build commands
make build      # Debug build
make release    # Optimized release build
make install    # Install to ~/.cargo/bin/

# Custom options
qg --width 200 --height 200 --filename custom.png https://example.com
qg --quiet https://example.com  # Suppress output
```

**Note**: Uses qrcode crate for generation and Clap for CLI.

### podver - Pod Version Scanner (Rust)

Scans Java and Node.js versions in Kubernetes pods.

```bash
# Scan Java and Node.js versions in Kubernetes pods
podver --namespaces production,staging

# Export to CSV
podver --namespaces production --output results.csv

# Increase concurrency and timeout
podver -n production -c 50 -t 60

# Include DaemonSet pods and enable verbose logging
podver --skip-daemonset=false --verbose -n default
```

**Technical Details**:
- Built with Tokio for async/concurrent pod scanning
- Executes `kubectl exec -- java -version` and `kubectl exec -- node --version` in parallel
- Parses Java version from stderr and Node.js version from stdout using regex
- Real-time multi-level progress bars (namespace + pod level)
- Generates kubectl-style tables and per-namespace statistics
- Configurable concurrency, timeouts, and DaemonSet filtering

### promdrop - Prometheus Metric Filter Generator (Rust)

Generates Prometheus metric drop configurations from mimirtool analysis.

```bash
# Generate metric drop configs from mimirtool analysis
# First run mimirtool to analyze metrics:
mimirtool analyze prometheus --output=prometheus-metrics.json

# Then generate drop configs (Rust version):
./target/release/promdrop --file prometheus-metrics.json

# Or use Makefile
make run        # Build and run with example
make release    # Optimized release build

# Custom output locations
promdrop --file prometheus-metrics.json \
  --txt-output-dir ./unused \
  --output combined_relabel_configs.yaml

# Container usage
docker run --rm -v $(pwd):/data \
  ghcr.io/younsl/promdrop:latest \
  --file /data/prometheus-metrics.json
```

**Technical Details**:
- Built with Rust using serde for JSON/YAML parsing
- CLI built with Clap for argument parsing
- Available as both CLI binary and container image
- Multi-arch Docker images (linux/amd64, linux/arm64)
- Automated releases via GitHub Actions (tag pattern: `promdrop/x.y.z`)

### filesystem-cleaner - Kubernetes Filesystem Cleanup Tool (Rust)

Lightweight container for automatic filesystem cleanup in Kubernetes environments.

```bash
# Quick test run
make run         # Dry-run cleanup of /tmp at 70% threshold

# Debug logging mode
make dev         # Same as run but with --log-level debug

# Container usage (as sidecar or init container)
docker run --rm -v /path:/path \
  ghcr.io/younsl/filesystem-cleaner:latest \
  --target-paths=/path --usage-threshold-percent=80
```

**Technical Details**:
- Operates as sidecar or init container
- Monitors disk usage and removes files when threshold exceeded
- Glob pattern support for include/exclude
- Graceful shutdown handling
- Structured logging with tracing crate

### elasticache-backup - ElastiCache S3 Backup Automation (Rust)

Automates ElastiCache snapshot creation and S3 export for Kubernetes CronJobs.

```bash
# Run locally with pretty logs
LOG_FORMAT=pretty ./target/debug/elasticache-backup \
  --cache-cluster-id "your-redis-cluster-001" \
  --s3-bucket-name "your-elasticache-backups"

# Run with debug logging
make dev

# Deploy as Kubernetes CronJob via Helm
helm install elasticache-backup ./box/kubernetes/elasticache-backup/charts/elasticache-backup \
  --set image.tag=0.1.0 \
  --set elasticache.cacheClusterId=your-cluster-id \
  --set s3.bucketName=your-bucket-name \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::ACCOUNT:role/ROLE_NAME
```

**Technical Details**:
- Creates ElastiCache snapshots from read replica nodes
- Exports snapshots to S3 buckets
- Automatic cleanup of source snapshots after export
- Configurable timeouts and retry intervals
- Structured JSON logging for CloudWatch/Loki integration
- IRSA support for AWS authentication
- Multi-architecture container images
- Helm chart for easy deployment

**Workflow**: Snapshot Creation → Wait → S3 Export → Wait → Cleanup

## Performance & API Guidelines

### GitHub API Constraints

**Critical**: `/orgs/{org}/actions/runs` does NOT exist. Must use:
1. List repos: `/orgs/{org}/repos`
2. Per-repo runs: `/repos/{owner}/{repo}/actions/runs`
3. Aggregate results manually

### Performance Anti-Patterns

Avoid:
- Complex adaptive delays without measurement
- Backpressure multipliers >1.5x
- Response time thresholds <2s for "slow"
- Dynamic behavior that confuses users

Prefer:
- Fixed, predictable delays
- Simple rate limiting
- Measurement before optimization
- User experience over theoretical efficiency

See `box/tools/cocd/docs/performance-optimization-lessons.md` for detailed case study (Korean).

## Release Workflow

GitHub Actions automatically builds and releases on tag push:

```bash
# CLI tool releases (pattern: {tool}/x.y.z)
git tag cocd/1.0.0 && git push --tags      # Go
git tag promdrop/1.0.0 && git push --tags  # Rust

# Container image releases (pattern: {container}/x.y.z)
git tag filesystem-cleaner/1.0.0 && git push --tags
git tag elasticache-backup/1.0.0 && git push --tags
git tag actions-runner/1.0.0 && git push --tags
git tag hugo/1.0.0 && git push --tags

# Available workflows:
# - release-cocd.yml                (Go CLI tool)
# - release-promdrop.yml            (Rust CLI + container)
# - release-filesystem-cleaner.yml  (Rust container)
# - release-elasticache-backup.yml  (Rust container)
# - release-actions-runner.yml      (Container image)
# - release-hugo.yml                (Container image)
# - release-backup-utils.yml        (Container image)

# Rust tools without automated releases (manual release required):
# - kk (domain connectivity checker)
# - qg (QR code generator)
# - podver (Pod version scanner - has Makefile docker-build/push targets)
```

## Testing Guidelines

**Current State**: Most tools lack comprehensive test coverage but Makefiles include test targets.

**When Adding Tests**:

Go projects (cocd):
- Place unit tests alongside source files (`*_test.go`)
- Use table-driven tests for multiple scenarios
- Mock AWS API calls using interfaces
- Follow Go's standard testing package conventions
- Test core logic in `internal/` and `pkg/` packages

Rust projects:
- Place unit tests in same file using `#[cfg(test)]` module
- Integration tests in `tests/` directory
- Use `cargo test --verbose` for running tests
- Use `#[tokio::test]` for async tests
- Mock external dependencies using traits
- Use `tempfile::TempDir` for filesystem tests

## Common Patterns

**Version Embedding**:
- Go: ldflags injection in Makefile
- Rust: Build-time environment variables or `build.rs` scripts

**Logging**:
- Go: logrus or zap
- Rust: tracing crate with configurable format (JSON/pretty)

**Configuration**:
- Both: Environment variables with CLI flag overrides via Clap (Rust) or standard libraries (Go)

**Async/Concurrency**:
- Go: Goroutines and channels
- Rust: Tokio async runtime with `async/await`

**Graceful Shutdown**:
- Both: Signal handling (SIGTERM, SIGINT) with cleanup logic
