# filesystem-cleaner

[![GitHub Container Registry](https://img.shields.io/badge/ghcr.io-younsl%2Ffilesystem--cleaner-000000?style=flat-square&logo=github&logoColor=white)](https://github.com/younsl/box/pkgs/container/filesystem-cleaner)
[![Go Version](https://img.shields.io/badge/go-1.25-000000?style=flat-square&logo=go&logoColor=white)](./go.mod)

A lightweight Go-based container image for automatic filesystem cleanup in Kubernetes environments. Designed as a sidecar container or init container, it monitors disk usage and intelligently removes files to prevent storage exhaustion. Particularly useful for GitHub Actions self-hosted runners, CI/CD pipelines, and any workloads that generate temporary files requiring periodic cleanup.

## Features

- **Automatic disk usage monitoring** - Triggers cleanup when usage exceeds threshold
- **Two operation modes**:
  - `once` - Single cleanup run (for initContainers)
  - `interval` - Periodic cleanup (for sidecar containers)
- **Configurable cleanup patterns** - Include/exclude file patterns
- **Dry-run mode** - Preview what would be deleted
- **Non-root execution** - Runs as unprivileged user

## Installation

filesystem-cleaner supports multiple deployment methods: standalone binary execution, container image, and Kubernetes sidecar/init container patterns. **The Kubernetes sidecar/init container pattern is the recommended approach and was the original design purpose** - specifically created to periodically free up caching disk space in build Actions Runner pods.

### Binary

```bash
make build
make install
```

### Docker

```bash
make docker-build
```

## Usage

### Command Line

```bash
filesystem-cleaner \
  --target-paths /home/runner/_work \
  --usage-threshold-percent 80 \
  --cleanup-mode interval \
  --check-interval-minutes 10 \
  --dry-run
```

### Kubernetes

**Important Requirements:**

- **Volume Mounting**: The filesystem-cleaner container must mount the same volume as the target pod (e.g., actions-runner) to access and clean the files. Ensure both containers share the same volume mount path.
- **Security Context**: Set `runAsUser: 1001` and `runAsGroup: 1001` to match the GitHub Actions runner user. This ensures the cleaner has proper permissions to delete files created by the runner without requiring elevated privileges.

#### Sidecar (Interval Mode)

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: actions-runner
    image: actions/runner:latest
    volumeMounts:
    - name: workspace
      mountPath: /home/runner/_work
  - name: filesystem-cleaner
    image: ghcr.io/younsl/filesystem-cleaner:0.1.0
    args:
    - "--target-paths=/home/runner/_work"
    - "--usage-threshold-percent=80"
    - "--cleanup-mode=interval"
    - "--check-interval-minutes=10"
    securityContext:
      runAsUser: 1001  # Same as runner user
      runAsGroup: 1001
      runAsNonRoot: true
    volumeMounts:
    - name: workspace
      mountPath: /home/runner/_work
  volumes:
  - name: workspace
    emptyDir: {}
```

### Kubernetes

#### Init Container (Once Mode)

```yaml
apiVersion: v1
kind: Pod
spec:
  initContainers:
  - name: filesystem-cleaner
    image: ghcr.io/younsl/filesystem-cleaner:0.1.0
    args:
    - "--target-paths=/home/runner/_work"
    - "--usage-threshold-percent=70"
    - "--cleanup-mode=once"
    - "--include-patterns=*"
    - "--exclude-patterns=.git,*.log"
    securityContext:
      runAsUser: 1001  # Same as runner user
      runAsGroup: 1001
      runAsNonRoot: true
    volumeMounts:
    - name: workspace
      mountPath: /home/runner/_work
  containers:
  - name: actions-runner
    image: actions/runner:latest
    volumeMounts:
    - name: workspace
      mountPath: /home/runner/_work
  volumes:
  - name: workspace
    emptyDir: {}
```

## Configuration

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--target-paths` | `string` | `/home/runner/_work` | Paths to monitor and clean (comma-separated) |
| `--usage-threshold-percent` | `int` | `80` | Disk usage percentage to trigger cleanup (0-100) |
| `--cleanup-mode` | `string` | `interval` | Cleanup mode: `once` or `interval` |
| `--check-interval-minutes` | `int` | `10` | Check interval in minutes (only used when `--cleanup-mode=interval`) |
| `--include-patterns` | `string` | `*` | File patterns to include (comma-separated) |
| `--exclude-patterns` | `string` | `.git,node_modules,*.log` | Patterns to exclude (comma-separated) |
| `--dry-run` | `bool` | `false` | Preview mode without deletion |
| `--log-level` | `string` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--version` | `bool` | `false` | Show version information |

## Building

```bash
# Local build
make build

# Multi-platform build
make build-all

# Docker image
make docker-build

# Push to ECR
make docker-push
```

## Development

```bash
# Run with debug logging
make dev

# Run tests
make test

# Format and lint
make fmt vet lint
```
