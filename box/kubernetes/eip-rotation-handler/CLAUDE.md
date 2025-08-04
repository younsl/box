# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EIP Rotation Handler is a Kubernetes DaemonSet that automatically rotates AWS Elastic IPs on worker nodes to bypass IP-based rate limits for forward proxy servers. It uses AWS EC2 Instance Metadata Service (IMDS) for automatic instance discovery and AWS SDK v2 for EIP management.

## Development Commands

```bash
# Build binary
make build

# Run all development tasks (deps, fmt, vet, test, build)
make all

# Development workflow
make deps     # Go mod tidy and download
make fmt      # Format code with go fmt
make vet      # Static analysis with go vet
make test     # Run tests
make clean    # Clean build artifacts

# Local development with debug logging
make dev      # Runs with LOG_LEVEL=debug and 1-minute rotation

# Code quality (requires golangci-lint)
make lint

# Container operations
make docker-build    # Build Docker image
make docker-push     # Push to ECR (requires AWS credentials)

# Deployment
make deploy          # Deploy to Kubernetes
```

## Architecture

### Core Components

- **`cmd/eip-rotation-handler/main.go`**: Application lifecycle management with graceful shutdown
- **`pkg/rotation/handler.go`**: Core EIP rotation logic with 6-step rotation process
- **`pkg/ec2/`**: AWS EC2 client and IMDS metadata client
- **`pkg/configs/config.go`**: Environment-based configuration with validation
- **`pkg/health/health.go`**: Health check server on port 8080
- **`pkg/logger/logger.go`**: Structured logging with logrus

### AWS Integration Pattern

The application follows this pattern for AWS resource management:
1. **IMDS Discovery**: Auto-detect instance ID, region, and current public IP
2. **EIP Allocation**: Request new Elastic IP from AWS
3. **EIP Association**: Attach new EIP to current instance
4. **Cleanup**: Release old EIP (with automatic rollback on failure)

### Key Environment Variables

- `ROTATION_INTERVAL_MINUTES`: Rotation frequency (1-1440 minutes, default: 10)
- `LOG_LEVEL`: Logging level (default: info)
- `METADATA_URL`: IMDS endpoint (default: http://169.254.169.254/latest/meta-data)
- `IMDS_VERSION`: IMDS version preference (auto, v1, v2, default: auto)

## Deployment Architecture

### Kubernetes Resources
- **DaemonSet**: Deploys one pod per node with `node-type: public` label
- **ServiceAccount**: Minimal RBAC permissions for Kubernetes API access
- **ClusterRole**: Required permissions for node operations

### Security Considerations
- **Host Network**: Required for direct access to IMDS at 169.254.169.254
- **IAM Integration**: Uses EC2 instance IAM role (not IRSA) for AWS API access
- **Non-root Container**: Runs as user 1000 with read-only filesystem
- **Required IAM Permissions**: `ec2:AllocateAddress`, `ec2:AssociateAddress`, `ec2:DescribeAddresses`, `ec2:ReleaseAddress`

### Helm Chart Configuration
- **Node Targeting**: Uses node affinity to target nodes with `node-type: public` label
- **Health Checks**: Liveness and readiness probes on `/healthz` endpoint
- **Resource Limits**: Conservative CPU/memory limits (10m CPU request, 40Mi memory limit)

## Testing and Quality

Run tests with proper AWS credentials configured (for integration tests that may contact AWS services):

```bash
# Unit tests
make test

# Code formatting and static analysis
make fmt vet

# Full quality check
make all
```