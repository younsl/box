# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a monorepo called "box" that serves as a comprehensive DevOps toolbox containing Kubernetes utilities, automation scripts, infrastructure code, and engineering documentation.

## Development Commands

### Go Projects

Most Go projects follow standard patterns with Makefiles:

```bash
# Build Go binaries
make build
make all  # build + test + fmt + vet

# Development
make test
make clean
make deps  # go mod tidy + download

# Code quality
make fmt   # go fmt
make vet   # go vet
make lint  # golangci-lint (if available)
```

Key Go projects:
- `box/kubernetes/eip-rotation-handler/` - Kubernetes DaemonSet for EIP rotation
- `box/kubernetes/jdk-version-scanner/` - JDK version scanning tool
- `box/tools/kk/` - Domain connectivity checker
- `box/tools/qg/` - QR code generator
- `box/tools/cryptopunk-generator/` - NFT-style image generator

### Container Images

```bash
# Build Docker images
make docker-build
make docker-push  # Push to ECR (requires AWS credentials)
```

### Resume Generation

```bash
# From box/resume/ directory
make open    # Open HTML resume with language toggle in Chrome
make pdf     # Generate both English and Korean PDFs
make pdf-en  # Generate English PDF (resume-en.pdf)
make pdf-ko  # Generate Korean PDF (resume-ko.pdf)
make clean   # Remove generated PDF files
make help    # Show all available commands
```

The resume includes bilingual support (EN/KO) with language toggle functionality. PDF generation uses headless Chrome with automatic language-specific styling and no headers/footers.

## Architecture Overview

### Kubernetes Ecosystem (`box/kubernetes/`)

- **Controllers**: DaemonSets and controllers for AWS resource management
- **Policies**: Kyverno and CEL admission policies for governance
- **Helm Charts**: Production-ready deployments with RBAC
- **Architecture Docs**: Mermaid diagrams explaining component relationships

### Tool Structure (`box/tools/`)

All CLI tools follow Go Standard Project Layout:
- `cmd/` - Main applications
- `internal/` or `pkg/` - Library code
- `configs/` - Configuration examples
- Standard Go module structure with versioned builds

### Scripts Collection (`box/scripts/`)

Organized by service/platform:
- `aws/` - AWS resource management (EBS, EIP)
- `github/` - Repository automation
- `k8s-registry-io-stat/` - Kubernetes connectivity testing

### Infrastructure (`box/terraform/`)

- Vault integration with AWS KMS
- IRSA (IAM Roles for Service Accounts) configurations

## Key Patterns

### Kubernetes Deployments
- All K8s applications use DaemonSet pattern for node-level operations
- Comprehensive RBAC with minimal required permissions
- Health checks and graceful shutdowns implemented
- AWS IAM integration via IRSA

### Go Applications
- Version information embedded via ldflags during build
- Environment-based configuration
- Structured logging throughout
- AWS SDK v2 integration for cloud services

### CI/CD Integration
- GitHub Actions workflows in `box/actions/`
- Terraform automation with scheduled applies
- Multi-architecture container builds

## AWS Integration

Many tools integrate with AWS services:
- EIP rotation for worker nodes
- ECR for container registry
- IAM roles via IRSA for Kubernetes workloads
- KMS for Vault auto-unseal

Set AWS credentials via environment or IAM roles before working with AWS-integrated tools.

## Performance Optimization Guidelines

When working on performance optimizations, especially for external API interactions:

### Anti-Patterns to Avoid
- **Premature Optimization**: Don't add complex performance logic without measuring actual problems
- **Aggressive Backpressure**: Avoid large delay multipliers (>1.5x) that hurt user experience  
- **Low Thresholds**: Use realistic thresholds based on actual server characteristics
- **Unpredictable Behavior**: Prefer fixed, predictable delays over dynamic adaptive delays

### Recommended Approach
1. **Measure First**: Profile and measure before optimizing
2. **Start Simple**: Use fixed delays and simple rate limiting
3. **Test Thoroughly**: Test under various load conditions including server slowdowns
4. **Monitor User Impact**: Always consider user experience over theoretical optimization

### Reference
See `box/tools/cocd/docs/performance-optimization-lessons.md` for detailed case study on PerformanceOptimizer anti-patterns.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
