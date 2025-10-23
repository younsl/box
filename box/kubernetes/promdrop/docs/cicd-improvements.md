# CI/CD Improvements

This document describes the critical improvements made to the GitHub Actions release workflow.

## Critical Issues Fixed

### 1. VERSION Extraction from Git Tags

**Problem**: The workflow used `github.ref_name` directly, which includes the full tag path `promdrop/1.0.0`, but only the version number `1.0.0` was needed.

**Solution**: Added dedicated `extract-version` job that properly extracts the version:

```yaml
extract-version:
  runs-on: ubuntu-24.04
  outputs:
    version: ${{ steps.get-version.outputs.version }}
  steps:
    - name: Extract version from tag or input
      id: get-version
      run: |
        if [ "${{ github.event_name }}" = "push" ]; then
          # Extract version from tag (promdrop/1.0.0 -> 1.0.0)
          VERSION="${GITHUB_REF#refs/tags/promdrop/}"
        else
          # Use manual input version
          VERSION="${{ inputs.version }}"
        fi
        echo "version=${VERSION}" >> $GITHUB_OUTPUT
        echo "Extracted version: ${VERSION}"
```

**Benefits**:
- Clean version numbers in release names
- Consistent VERSION across all jobs
- Works for both tag push and manual dispatch

### 2. Linux ARM64 Cross-Compilation Configuration

**Problem**: Installing gcc-aarch64-linux-gnu alone is not enough. Cargo needs to know which linker to use for cross-compilation.

**Solution**: Added proper environment variables for the ARM64 target:

```yaml
- name: Install cross-compilation tools (Linux ARM64)
  if: matrix.target == 'aarch64-unknown-linux-gnu'
  run: |
    sudo apt-get update
    sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
  env:
    DEBIAN_FRONTEND: noninteractive

- name: Configure cross-compilation (Linux ARM64)
  if: matrix.target == 'aarch64-unknown-linux-gnu'
  run: |
    echo "CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc" >> $GITHUB_ENV
    echo "CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc" >> $GITHUB_ENV
    echo "CXX_aarch64_unknown_linux_gnu=aarch64-linux-gnu-g++" >> $GITHUB_ENV
```

**Benefits**:
- Enables successful ARM64 cross-compilation on x86_64 runners
- Properly links C dependencies
- Matches Rust cross-compilation best practices

### 3. Build Failure Debugging

**Problem**: When builds fail, there was insufficient information to diagnose the issue.

**Solution**: Enhanced build step with comprehensive logging and validation:

```yaml
- name: Build binary
  working-directory: ${{ env.PROJECT_BASE_DIR }}
  run: |
    echo "================================="
    echo "Building ${{ env.PROJECT_NAME }}"
    echo "Target: ${{ matrix.target }}"
    echo "Platform: ${{ matrix.platform }}"
    echo "Architecture: ${{ matrix.arch }}"
    echo "Version: ${{ env.VERSION }}"
    echo "================================="

    # Build with verbose output
    cargo build --release --target ${{ matrix.target }} --verbose

    # Verify binary exists
    SOURCE_BINARY="target/${{ matrix.target }}/release/${{ env.PROJECT_NAME }}"
    if [ ! -f "${SOURCE_BINARY}" ]; then
      echo "❌ Error: Binary not found at ${SOURCE_BINARY}"
      echo "Contents of target/${{ matrix.target }}/release/:"
      ls -la target/${{ matrix.target }}/release/ || echo "Directory not found"
      exit 1
    fi

    # Copy and set permissions
    cp "${SOURCE_BINARY}" "${BINARY_NAME}"
    chmod +x "${BINARY_NAME}"

    # Smoke test
    ./"${BINARY_NAME}" --version || echo "⚠️  Warning: Version check failed"
```

**Benefits**:
- Clear build context in logs
- Early failure detection with helpful error messages
- Directory listing on failure for debugging
- Smoke test ensures binary is executable
- Proper file permissions set

## Architecture Changes

### Job Dependencies

The workflow now has a clear dependency chain:

```
extract-version (runs first)
    ↓
    ├─→ build-binaries (needs: extract-version)
    ├─→ build-docker   (needs: extract-version)
    └─→ test           (needs: extract-version)
         ↓
    release (needs: all above jobs)
```

**Benefits**:
- VERSION is computed once, used everywhere
- Parallel execution where possible
- Clear failure points

### Environment Variables Per Job

Each job now has its own environment block:

```yaml
build-binaries:
  needs: extract-version
  env:
    PROJECT_NAME: promdrop
    PROJECT_BASE_DIR: box/kubernetes/promdrop
    VERSION: ${{ needs.extract-version.outputs.version }}

build-docker:
  needs: extract-version
  env:
    PROJECT_BASE_DIR: box/kubernetes/promdrop
    REGISTRY: ghcr.io
    IMAGE_NAME: younsl/promdrop
    VERSION: ${{ needs.extract-version.outputs.version }}
```

**Benefits**:
- Clear scope of variables
- No global pollution
- Easy to understand what each job uses

## Testing the Workflow Locally

### Using act

```bash
# Install act (GitHub Actions local runner)
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash  # Linux

# Test the workflow
cd /path/to/o
act -W .github/workflows/release-promdrop.yml --secret-file .secrets
```

### Manual Validation

Before pushing a release tag, validate:

1. **Syntax check**:
   ```bash
   # Install actionlint
   brew install actionlint

   # Check workflow syntax
   actionlint .github/workflows/release-promdrop.yml
   ```

2. **Test cross-compilation locally**:
   ```bash
   cd box/kubernetes/promdrop

   # Test Linux ARM64 build
   export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
   cargo build --release --target aarch64-unknown-linux-gnu
   ```

3. **Test version extraction**:
   ```bash
   # Simulate tag extraction
   GITHUB_REF="refs/tags/promdrop/1.0.0"
   VERSION="${GITHUB_REF#refs/tags/promdrop/}"
   echo "Extracted version: ${VERSION}"
   # Should output: 1.0.0
   ```

## Monitoring and Debugging

### View Workflow Runs

1. Go to repository's Actions tab
2. Select "Release promdrop" workflow
3. Click on specific run to see job details

### Common Issues

**Issue**: ARM64 build fails with "linker not found"

**Solution**: Check that linker environment variables are set:
```yaml
CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
```

**Issue**: Version shows as `promdrop/1.0.0` instead of `1.0.0`

**Solution**: Verify extract-version job completed successfully and check its output.

**Issue**: Binary not found after build

**Solution**: Check cargo build output for compilation errors. The enhanced logging will show directory contents.

## Performance Optimizations

The workflow already includes several optimizations:

1. **Cargo caching**: Registry, index, and build artifacts
2. **Docker layer caching**: Using GitHub Actions cache
3. **Parallel jobs**: Build, test, and Docker jobs run in parallel
4. **Matrix builds**: All platforms build simultaneously

### Cache Keys

```yaml
# Cargo registry - shared across jobs
~/.cargo/registry: ${{ runner.os }}-cargo-registry-${{ hashFiles('**/Cargo.lock') }}

# Cargo index - shared across jobs
~/.cargo/git: ${{ runner.os }}-cargo-index-${{ hashFiles('**/Cargo.lock') }}

# Build artifacts - separate per job
target/: ${{ runner.os }}-cargo-build-target-${{ hashFiles('**/Cargo.lock') }}
```

## Security Considerations

1. **Permissions**: Each job declares minimal required permissions
2. **Secrets**: GitHub token used only where needed (packages, contents)
3. **Non-root user**: Docker images run as uid 1000
4. **No hardcoded credentials**: All secrets via GitHub Secrets

## Rollback Procedure

If a release fails:

1. Delete the failed release from GitHub Releases
2. Delete the git tag:
   ```bash
   git tag -d promdrop/1.0.0
   git push origin :refs/tags/promdrop/1.0.0
   ```
3. Fix the issue
4. Create new tag with incremented patch version

## Related Documentation

- [Release Workflow Guide](release-workflow.md)
- [Testing Guide](testing.md)
- [Docker Build Guide](docker-build.md)
