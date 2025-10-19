# Docker Build Guide

This guide explains how to build and run promdrop using Docker/Podman.

## Prerequisites

- Docker 20.10+ or Podman 4.0+
- At least 2GB of available disk space

## Building the Image

### Using Docker

```bash
# Build the image
docker build -t promdrop:local .

# Build with specific version tag
docker build -t promdrop:1.0.0 .
```

### Using Podman

```bash
# Build the image
podman build -t promdrop:local .

# Build with specific version tag
podman build -t promdrop:1.0.0 .
```

## Multi-stage Build Process

The Dockerfile uses a two-stage build:

### Stage 1: Builder (rust:1.83-slim)

```dockerfile
FROM rust:1.83-slim as builder
WORKDIR /build
COPY Cargo.toml ./
COPY Cargo.lock* ./
COPY src ./src
RUN cargo build --release
```

This stage:
- Uses official Rust image for compilation
- Copies dependency manifests first (better caching)
- Builds optimized release binary
- Total image size: ~1.5GB (temporary)

### Stage 2: Runtime (debian:bookworm-slim)

```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /build/target/release/promdrop /usr/local/bin/promdrop
RUN useradd -m -u 1000 promdrop
USER promdrop
WORKDIR /data
ENTRYPOINT ["promdrop"]
CMD ["--help"]
```

This stage:
- Uses minimal Debian base (~80MB)
- Includes CA certificates for HTTPS
- Runs as non-root user (uid 1000)
- Final image size: ~150MB

## Running the Container

### Basic Usage

```bash
# Show help
docker run --rm promdrop:local --help

# Process metrics file
docker run --rm \
  -v $(pwd):/data \
  promdrop:local \
  --file /data/prometheus-metrics.json
```

### With Custom Output Directory

```bash
# Create output directory
mkdir -p unused

# Run with volume mounts
docker run --rm \
  -v $(pwd):/data \
  -v $(pwd)/unused:/unused \
  promdrop:local \
  --file /data/prometheus-metrics.json \
  --txt-output-dir /unused \
  --output /data/combined_relabel_configs.yaml
```

### Interactive Mode

```bash
# Enter container shell for debugging
docker run --rm -it \
  -v $(pwd):/data \
  --entrypoint /bin/bash \
  promdrop:local
```

## Build Optimizations

### Dependency Caching

The Dockerfile is structured to leverage Docker layer caching:

1. **Layer 1**: Copy Cargo.toml and Cargo.lock
2. **Layer 2**: Copy source code
3. **Layer 3**: Build release binary

This means dependency downloads are cached separately from source changes.

### .dockerignore

The `.dockerignore` file excludes unnecessary files:

```
target/          # Build artifacts
docs/            # Documentation
tests/           # Test files
*.md             # Markdown files
.github/         # CI/CD configs
```

This reduces build context size and speeds up builds.

### Multi-architecture Builds

For ARM64 support:

```bash
# Using Docker Buildx
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t promdrop:multi-arch \
  .

# Using Podman
podman build \
  --platform linux/amd64,linux/arm64 \
  --manifest promdrop:multi-arch \
  .
```

## Troubleshooting

### Build Fails: "Cargo.lock not found"

**Solution**: Ensure Cargo.lock exists and is not in .gitignore

```bash
# Generate Cargo.lock if missing
cargo generate-lockfile

# Verify it's not ignored
git check-ignore Cargo.lock
```

### Build Fails: "Out of memory"

**Solution**: Increase Docker memory limit

```bash
# Docker Desktop: Settings > Resources > Memory
# Or use docker run with memory limit
docker build --memory 4g -t promdrop:local .
```

### Runtime Error: "Permission denied"

**Solution**: The container runs as non-root user (uid 1000). Ensure mounted volumes have correct permissions:

```bash
# Fix volume permissions
chmod 755 $(pwd)
chmod 644 prometheus-metrics.json
```

### Slow Builds

**Solutions**:

1. **Enable BuildKit** (Docker):
   ```bash
   DOCKER_BUILDKIT=1 docker build -t promdrop:local .
   ```

2. **Use build cache**:
   ```bash
   docker build --cache-from promdrop:local -t promdrop:local .
   ```

3. **Prune build cache periodically**:
   ```bash
   docker builder prune -a
   ```

## Security Considerations

### Non-root User

The container runs as user `promdrop` (uid 1000):

```dockerfile
RUN useradd -m -u 1000 promdrop
USER promdrop
```

This follows the principle of least privilege.

### Minimal Base Image

Using `debian:bookworm-slim`:
- Small attack surface
- Only essential packages
- Regular security updates

### No Secrets in Image

The Dockerfile:
- Does not embed credentials
- Does not copy sensitive files
- Uses build-time arguments only

### CA Certificates

HTTPS support requires CA certificates:

```dockerfile
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*
```

This allows secure communication with external services.

## Image Size Comparison

| Stage | Image | Size |
|-------|-------|------|
| Builder | rust:1.83-slim | ~1.5GB |
| Runtime | debian:bookworm-slim | ~80MB |
| Final | promdrop:local | ~150MB |

The multi-stage build reduces final image size by 90%.

## Best Practices

1. **Use specific tags**: Avoid `:latest` in production
2. **Scan for vulnerabilities**: Use `docker scan` or Trivy
3. **Keep images updated**: Rebuild regularly for security patches
4. **Minimize layers**: Combine RUN commands where possible
5. **Use .dockerignore**: Reduce build context size
6. **Pin dependencies**: Lock Cargo.toml versions
7. **Test locally**: Before pushing to registry

## CI/CD Integration

The GitHub Actions workflow automatically:
- Builds multi-arch images
- Pushes to ghcr.io
- Tags with version and latest
- Includes metadata and labels

See [release-workflow.md](release-workflow.md) for details.

## Related Documentation

- [Installation Guide](installation.md)
- [Release Workflow](release-workflow.md)
- [Testing Guide](testing.md)
