# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

A monorepo serving as a DevOps toolbox containing Kubernetes utilities, automation scripts, infrastructure code, and engineering documentation.

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

### Container Operations

```bash
make docker-build   # Build Docker image
make docker-push    # Push to ECR (requires AWS credentials)
make deploy         # Deploy to Kubernetes (where available)
```

**Note**: Not all tools have identical Makefile targets. Check project-specific Makefiles for variations:
- `make mod` (cocd) vs `make deps` (idled, jdk-scanner)
- Some projects include `make vet` or `make lint` targets

## High-Level Architecture

### Repository Structure

```
box/
├── kubernetes/             # K8s controllers, policies, helm charts
│   ├── jdk-version-scanner/    # JDK version scanning tool
│   ├── promdrop/          # Prometheus metric filter generator
│   └── policies/          # Kyverno and CEL admission policies
├── tools/                 # CLI utilities (Go-based)
│   ├── cocd/              # GitHub Actions deployment monitor
│   ├── idled/             # AWS idle resource scanner
│   ├── kk/                # Domain connectivity checker
│   └── qg/                # QR code generator
├── containers/            # Custom container images
│   ├── actions-runner/    # GitHub Actions runner
│   ├── filesystem-cleaner/# File system cleanup tool
│   ├── hugo/              # Hugo static site generator
│   └── terraform-console-machine/  # Terraform console container
├── scripts/               # Automation scripts by platform
│   ├── aws/               # AWS resource management
│   ├── github/            # Repository automation
│   └── k8s-registry-io-stat/  # K8s connectivity testing
├── terraform/             # Infrastructure as Code
│   ├── vault/irsa/        # Vault with AWS KMS integration
│   └── terraform-elasticache-*/  # ElastiCache backup Lambda
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

### kk - Domain Connectivity Checker

```bash
# Check domain connectivity
./kk --config configs/domain-example.yaml

# Configuration format (YAML):
domains:
  - www.google.com        # Auto-adds https://
  - reddit.com
  - https://registry.k8s.io/v2/
```

### qg - QR Code Generator

```bash
# Generate QR code from URL
./qg [flags] <url>
./qg --help              # Show usage
```

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

See `box/tools/cocd/docs/performance-optimization-lessons.md` for detailed case study (written in Korean).

## Release Workflow

GitHub Actions automatically builds and releases on tag push:

```bash
# Tool releases (pattern: {tool}/x.y.z)
git tag cocd/1.0.0 && git push --tags
git tag idled/1.0.0 && git push --tags
git tag promdrop/1.0.0 && git push --tags

# Note: Not all tools have automated release workflows
# Check .github/workflows/release-*.yml for available automation

# Available workflows:
# - release-cocd.yml
# - release-idled.yml
# - release-promdrop.yml
# - release-filesystem-cleaner.yml
# - release-actions-runner.yml
# - release-hugo.yml
# - release-backup-utils.yml
```

## Testing Guidelines

**Current State**: Most tools lack test files but Makefiles include test targets.

**When Adding Tests**:
- Place unit tests alongside source files (`*_test.go`)
- Use table-driven tests for multiple scenarios
- Mock AWS API calls using interfaces
- Follow Go's standard testing package conventions
- Test core logic in `internal/` and `pkg/` packages