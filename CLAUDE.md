# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

A monorepo serving as a DevOps toolbox containing Kubernetes utilities, automation scripts, infrastructure code, and engineering documentation.

**Language Migration Notice**: The repository is migrating CLI tools from Go to Rust for better performance, memory safety, and modern tooling. Completed migrations include `kk`, `qg`, and `jvs`. In-progress migrations include `cocd`, `idled`, and `promdrop`.

## Development Commands

### Go Projects

Standard Makefile patterns across all Go projects:

```bash
# Core build commands
make build          # Build binary
make all            # deps + fmt + vet + test + build
make build-all      # Build for all platforms (linux/darwin, amd64/arm64)

# Development workflow
make run            # Run application
make test           # Run tests
make dev            # Run with debug logging (where available)
make install        # Install to system

# Testing (standard: go test -v ./...)
go test -v ./...                    # Run all tests
go test -v ./pkg/specific/package   # Run tests for specific package
go test -v -run TestFunctionName    # Run specific test function

# Code quality
make fmt            # Format code
make vet            # Static analysis
make lint           # golangci-lint (if installed)
make deps           # go mod tidy + download
make clean          # Remove build artifacts
```

**Note**: Not all tools have identical Makefile targets. Check project-specific Makefiles for variations:
- `make mod` (cocd) vs `make deps` (idled, jvs)
- Some projects include `make vet` or `make lint` targets

### Rust Projects

Standard Makefile patterns for Rust tools (kk, qg, jvs):

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
- **filesystem-cleaner**: Includes `make vet` and `make dev` targets for debug logging
- **actions-runner**, **hugo**: Have specialized build workflows for multi-arch images
- Update ECR_REGISTRY variable in Makefiles before pushing

### Terraform Projects

Standard Terraform workflow for infrastructure modules:

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy
```

**Available Modules**:
- `vault/irsa/` - Vault auto-unseal with AWS KMS integration
- `terraform-elasticache-snapshot-backup-lambda/` - ElastiCache backup automation

## High-Level Architecture

### Repository Structure

```
box/
├── kubernetes/             # K8s controllers, policies, helm charts
│   ├── jvs/               # Java Version Scanner (Rust)
│   ├── promdrop/          # Prometheus metric filter generator
│   └── policies/          # Kyverno and CEL admission policies
├── tools/                 # CLI utilities
│   ├── cocd/              # GitHub Actions deployment monitor (Go)
│   ├── idled/             # AWS idle resource scanner (Go)
│   ├── kk/                # Domain connectivity checker (Rust)
│   └── qg/                # QR code generator (Rust)
├── containers/            # Custom container images
│   ├── actions-runner/    # GitHub Actions runner
│   ├── filesystem-cleaner/# File system cleanup tool (Go)
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
│   ├── vault/irsa/        # Vault auto-unseal with AWS KMS
│   └── terraform-elasticache-snapshot-backup-lambda/  # ElastiCache backup Lambda
├── actions/               # GitHub Actions reusable workflows
└── til/                   # Engineering notes and learnings
```

### Architectural Patterns

**Kubernetes Applications**:
- DaemonSet pattern for node-level operations
- IMDS access via host network when required
- IRSA for AWS API authentication
- Health endpoints on port 8080
- Graceful shutdown handling

**Go Application Structure**:
- `cmd/` - Application entry points
- `pkg/` or `internal/` - Reusable packages
- Version embedding via ldflags (see patterns below)
- Environment-based configuration
- Structured logging (logrus/zap)
- AWS SDK v2 integration

**Version Embedding Patterns**:
```go
// Common ldflags patterns in Makefiles:
// Simple pattern (cocd):
-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

// Package-based pattern (idled):
-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).buildDate=$(BUILD_DATE) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)"
```

**Rust Application Structure**:
- `src/main.rs` - CLI entry point with Clap argument parsing
- `src/lib.rs` - Core library code (if applicable)
- `src/*.rs` - Module files for specific functionality
- `Cargo.toml` - Rust dependencies and metadata
- Tokio async runtime for concurrent operations
- Structured logging with tracing crate

**CI/CD Pipeline**:
- GitHub Actions for releases
- Multi-arch builds (linux/darwin, amd64/arm64)
- Automated binary releases with tags
- Container image push to ECR

## AWS Integration Points

- **ECR**: Container registry for Kubernetes deployments
- **IAM/IRSA**: Service account to IAM role mapping
- **KMS**: Vault auto-unseal encryption
- **EC2**: Instance and resource management (idled scanner)

Configure AWS credentials via environment variables or IAM instance profiles.

## Tool-Specific Notes

### cocd - GitHub Actions Monitor

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

### idled - AWS Idle Resource Scanner

```bash
# Scan idle resources across regions
idled ec2 --regions all          # EC2 instances
idled ebs --regions us-east-1    # EBS volumes
idled s3                          # S3 buckets (global)

# Service-specific idle criteria:
# - EC2: Stopped instances
# - EBS: Unattached volumes
# - S3: No access for 90+ days
# - Lambda: No invocations in 30+ days
# - EIP: Unassociated addresses
```

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

**Note**: kk is written in Rust (previously Go). Uses Tokio for async concurrency and Clap for CLI.

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

**Note**: qg is written in Rust (previously Go). Uses qrcode crate for generation and Clap for CLI.

### jvs - Java Version Scanner (Rust)

```bash
# Scan Java versions in Kubernetes pods
jvs --namespaces production,staging

# Export to CSV
jvs --namespaces production --output results.csv

# Increase concurrency and timeout
jvs -n production -c 50 -t 60

# Include DaemonSet pods and enable verbose logging
jvs --skip-daemonset=false --verbose -n default
```

**Technical Details**:
- Built with Tokio for async/concurrent pod scanning
- Executes `kubectl exec -- java -version` in parallel
- Parses Java version from stderr using regex
- Real-time multi-level progress bars (namespace + pod level)
- Generates kubectl-style tables and per-namespace statistics
- Configurable concurrency, timeouts, and DaemonSet filtering

### promdrop - Prometheus Metric Filter Generator

```bash
# Generate metric drop configs from mimirtool analysis
# First run mimirtool to analyze metrics:
mimirtool analyze prometheus --output=prometheus-metrics.json

# Then generate drop configs:
./promdrop --file prometheus-metrics.json
```

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
# Go/Rust tool releases (pattern: {tool}/x.y.z)
git tag cocd/1.0.0 && git push --tags
git tag idled/1.0.0 && git push --tags
git tag promdrop/1.0.0 && git push --tags

# Container image releases (pattern: {container}/x.y.z)
git tag filesystem-cleaner/1.0.0 && git push --tags
git tag actions-runner/1.0.0 && git push --tags
git tag hugo/1.0.0 && git push --tags

# Note: Not all tools have automated release workflows
# Rust tools (kk, qg, jvs) currently lack automated releases
# Check .github/workflows/release-*.yml for available automation

# Available workflows (Go tools & containers):
# - release-cocd.yml          (Go CLI tool)
# - release-idled.yml         (Go CLI tool)
# - release-promdrop.yml      (Go CLI tool)
# - release-filesystem-cleaner.yml  (Container image)
# - release-actions-runner.yml      (Container image)
# - release-hugo.yml                (Container image)
# - release-backup-utils.yml        (Container image)

# Rust tools without automated releases (manual release required):
# - kk (domain connectivity checker)
# - qg (QR code generator)
# - jvs (Java version scanner)
```

## Testing Guidelines

**Current State**: Most tools lack test files but Makefiles include test targets.

**When Adding Tests**:

Go projects:
- Place unit tests alongside source files (`*_test.go`)
- Use table-driven tests for multiple scenarios
- Mock AWS API calls using interfaces
- Follow Go's standard testing package conventions
- Test core logic in `internal/` and `pkg/` packages

Rust projects:
- Place unit tests in same file using `#[cfg(test)]` module
- Integration tests in `tests/` directory
- Use `cargo test --verbose` for running tests
- Mock external dependencies using traits