# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

idled (idle detector) is a CLI tool that scans AWS resources across regions to identify idle/underutilized resources and calculate potential cost savings. It supports EC2, EBS, S3, Lambda, EIP, IAM, Config, ELB, CloudWatch Logs, ECR, MSK, and Secrets Manager.

## Development Commands

### Build & Development

```bash
make build          # Build binary for current OS/arch
make all            # Clean + format + test + build
make run            # Build and run the application
make install        # Install to GOPATH/bin
make clean          # Remove build artifacts
make deps           # Update dependencies (go mod tidy)
```

### Code Quality

```bash
make fmt            # Format code with gofmt
make test           # Run all tests
```

### Cross-Compilation

```bash
# Build for specific OS/arch
GOOS=linux GOARCH=amd64 make build
GOOS=darwin GOARCH=arm64 make build
```

## Architecture

### Package Structure

```
cmd/idled/          # CLI entry point with cobra commands
pkg/
├── aws/            # AWS SDK v2 service clients and resource scanning
├── models/         # Data structures for each AWS service
├── formatter/      # Table formatters for output display
├── pricing/        # AWS Pricing API integration for cost calculations
├── utils/          # Shared utilities (formatting, time, tags, regions)
└── version/        # Version info embedded via ldflags
```

### Key Patterns

**Resource Scanning Flow**:
1. Parse CLI flags (regions, services)
2. Initialize AWS clients per region
3. Scan resources in parallel using goroutines
4. Calculate idle time and potential cost savings
5. Display results in table format with summary statistics

**AWS Integration**:
- Uses AWS SDK v2 with context-aware operations
- Parallel region scanning with sync.WaitGroup
- Real-time pricing data via AWS Pricing API
- Graceful error handling per region

**Output Formatting**:
- Kubernetes-style tables using tabwriter
- Sorting by idle time (longest first)
- Unicode-aware column width calculation
- Color coding for idle resources
- Summary statistics with cost breakdowns

## Service-Specific Implementation

### Adding New AWS Service

1. Create model in `pkg/models/{service}.go`
2. Implement scanner in `pkg/aws/{service}.go` with parallel region support
3. Add formatter in `pkg/formatter/{service}_table.go`
4. Integrate in `cmd/idled/main.go` with service flag
5. Add pricing support in `pkg/pricing/{service}.go` if applicable

### Idle Detection Criteria

- **EC2**: Stopped instances
- **EBS**: Unattached volumes
- **S3**: No objects or access for 90+ days
- **Lambda**: No invocations in 30+ days
- **EIP**: Unassociated addresses
- **IAM**: No activity in 90+ days
- **MSK**: ConnectionCount=0 or CPU<30% over 30 days
- **Config**: Disabled rules or no recent evaluations
- **ELB**: No targets or traffic
- **CloudWatch Logs**: No ingestion in 90+ days
- **ECR**: No pulls/pushes in 90+ days
- **Secrets Manager**: No access in 90+ days

## Testing

Currently no test files exist. When adding tests:
- Place unit tests alongside code files (*_test.go)
- Use table-driven tests for multiple scenarios
- Mock AWS API calls using interfaces

## Release Process

GitHub Actions workflow (`release-idled.yml`) triggers on tags:
- Pattern: `idled/[0-9]+.[0-9]+.[0-9]+`
- Builds for linux/darwin and amd64/arm64
- Version info embedded via ldflags
- Creates GitHub release with artifacts

## Environment Configuration

```bash
export AWS_PROFILE=your-profile    # Required for AWS credentials
export AWS_REGION=us-east-1        # Default region (optional)
```

## Performance Considerations

- Spinner feedback during long-running scans
- Parallel region processing to reduce latency
- Pricing data caching to minimize API calls
- Early termination on context cancellation