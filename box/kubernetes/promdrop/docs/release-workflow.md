# Release Workflow

This document describes the automated release process for promdrop.

## Release Triggers

The release workflow is triggered by:

1. **Git Tag Push** - Tags matching pattern `promdrop/x.y.z`
2. **Manual Workflow Dispatch** - Via GitHub Actions UI

## Workflow Structure

The release pipeline consists of four jobs:

### 1. Build Binaries

Builds cross-platform binaries for multiple architectures.

**Platforms supported**:
- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)

**Build process**:
1. Setup Rust toolchain with target support
2. Cache cargo dependencies for faster builds
3. Compile release binary with optimizations
4. Create compressed tar.gz archives
5. Upload artifacts for release job

**Output**: `promdrop-{platform}-{arch}.tar.gz`

### 2. Build Docker Image

Builds and pushes multi-architecture container images.

**Platforms supported**:
- linux/amd64
- linux/arm64

**Build process**:
1. Setup QEMU for cross-platform builds
2. Configure Docker Buildx
3. Login to GitHub Container Registry
4. Build multi-arch image using Dockerfile
5. Push to ghcr.io/younsl/promdrop

**Image tags**:
- `ghcr.io/younsl/promdrop:{version}` - Specific version
- `ghcr.io/younsl/promdrop:{major}.{minor}` - Major.minor version
- Auto-generated from tag

### 3. Test

Validates code quality before release.

**Test steps**:
1. Run all unit tests
2. Run integration tests
3. Run end-to-end tests
4. Run clippy linter with strict checks
5. Verify code formatting

**Exit conditions**:
- All tests must pass
- No clippy warnings allowed
- Code must be properly formatted

### 4. Release

Creates GitHub release with all artifacts.

**Prerequisites**:
- All previous jobs must succeed
- Test job must pass

**Release contents**:
- Binary archives for all platforms
- SHA256 checksums file
- Auto-generated release notes
- Installation instructions
- Documentation links

## Creating a Release

### Option 1: Tag-based Release (Recommended)

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Create and push release tag
git tag promdrop/1.0.0
git push origin promdrop/1.0.0
```

The workflow will automatically:
1. Build binaries for all platforms
2. Build and push Docker images
3. Run full test suite
4. Create GitHub release

### Option 2: Manual Workflow Dispatch

1. Go to GitHub Actions
2. Select "Release promdrop" workflow
3. Click "Run workflow"
4. Enter version (e.g., "1.0.0")
5. Click "Run workflow"

## Version Numbering

Follow semantic versioning (SemVer):

- **Major** (1.0.0): Breaking changes
- **Minor** (0.1.0): New features, backward compatible
- **Patch** (0.0.1): Bug fixes, backward compatible

Examples:
- `promdrop/1.0.0` - First stable release
- `promdrop/1.1.0` - Added new features
- `promdrop/1.1.1` - Bug fixes

## Workflow Environment Variables

```yaml
PROJECT_NAME: promdrop
PROJECT_BASE_DIR: box/kubernetes/promdrop
REGISTRY: ghcr.io
IMAGE_NAME: younsl/promdrop
VERSION: {github.ref_name or input.version}
```

## Build Matrix

### Binary Builds

| OS | Target | Platform | Architecture |
|----|--------|----------|--------------|
| ubuntu-24.04 | x86_64-unknown-linux-gnu | linux | amd64 |
| ubuntu-24.04 | aarch64-unknown-linux-gnu | linux | arm64 |
| macos-latest | x86_64-apple-darwin | darwin | amd64 |
| macos-latest | aarch64-apple-darwin | darwin | arm64 |

### Docker Builds

- Platform: linux/amd64, linux/arm64
- Base image: rust:1.83-slim (builder), debian:bookworm-slim (runtime)
- Registry: GitHub Container Registry (ghcr.io)

## Caching Strategy

The workflow uses GitHub Actions cache for:

1. **Cargo registry**: `~/.cargo/registry`
2. **Cargo index**: `~/.cargo/git`
3. **Build artifacts**: `target/`
4. **Docker layers**: GitHub Actions cache (gha)

This significantly reduces build times for subsequent releases.

## Artifact Structure

After a successful release, the following artifacts are available:

```
Release Assets:
├── promdrop-linux-amd64.tar.gz
├── promdrop-linux-arm64.tar.gz
├── promdrop-darwin-amd64.tar.gz
├── promdrop-darwin-arm64.tar.gz
└── checksums.txt
```

Container Images:
```
ghcr.io/younsl/promdrop:{version}
ghcr.io/younsl/promdrop:{major}.{minor}
```

## Troubleshooting

### Build Fails on Specific Platform

Check the build logs for the specific platform matrix job:
1. Go to Actions tab
2. Click on the failed workflow run
3. Expand the failed job
4. Review build output

Common issues:
- Missing Rust toolchain target
- Cross-compilation tools not installed
- Dependency version conflicts

### Docker Build Fails

Common causes:
- Dockerfile syntax errors
- Missing dependencies in base image
- Build context issues

Fix:
1. Test locally: `docker build -t promdrop:test .`
2. Check .dockerignore file
3. Verify Dockerfile syntax

### Tests Fail

Before release:
```bash
# Run tests locally
cd box/kubernetes/promdrop
cargo test --verbose
cargo clippy -- -D warnings
cargo fmt --check
```

### Release Creation Fails

Check permissions:
- Workflow needs `contents: write` permission
- Ensure GITHUB_TOKEN has correct scopes

## Post-Release Checklist

After successful release:

1. Verify release on GitHub Releases page
2. Test binary downloads:
   ```bash
   curl -LO https://github.com/younsl/o/releases/download/promdrop/1.0.0/promdrop-linux-amd64.tar.gz
   tar -xzf promdrop-linux-amd64.tar.gz
   ./promdrop-linux-amd64 --version
   ```

3. Test Docker image:
   ```bash
   docker pull ghcr.io/younsl/promdrop:1.0.0
   docker run --rm ghcr.io/younsl/promdrop:1.0.0 --version
   ```

4. Update documentation if needed
5. Announce release (if applicable)

## Best Practices

1. **Always tag from main branch** - Ensure stable codebase
2. **Test locally first** - Run `cargo test` before tagging
3. **Use semantic versioning** - Follow SemVer strictly
4. **Write release notes** - Document changes clearly
5. **Verify artifacts** - Download and test before announcing

## Security Considerations

- Container images run as non-root user (uid 1000)
- Minimal runtime dependencies (Debian slim)
- CA certificates included for HTTPS support
- No secrets or credentials in images
- GitHub token has minimal required permissions

## Related Documentation

- [Installation Guide](installation.md)
- [Testing Guide](testing.md)
- [Usage Guide](usage.md)
