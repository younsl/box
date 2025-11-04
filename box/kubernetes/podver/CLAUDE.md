# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`podver` (Pod Version Scanner) is a Kubernetes CLI tool that scans and reports Java and Node.js runtime versions across pods in a cluster. Built in Rust, it uses async/concurrent processing with Tokio to scan hundreds of pods efficiently.

## Development Commands

### Building and Testing

```bash
# Standard Rust workflow
cargo build              # Debug build → target/debug/podver
cargo build --release    # Optimized build → target/release/podver
cargo test --verbose     # Run all tests
cargo fmt                # Format code
cargo clippy -- -D warnings  # Lint

# Makefile shortcuts (Korean help text)
make build              # Debug build
make release            # Release build
make test               # Run tests
make install            # Install to ~/.cargo/bin/
make run                # Build and show --help
make dev                # Run with --verbose
```

### Running Tests

```bash
# Run all tests
cargo test --verbose

# Run specific test
cargo test test_version_parse

# Run tests in specific module
cargo test types::tests::
cargo test scanner::tests::
```

## Architecture

### Module Structure

- **`main.rs`** - Entry point, CLI initialization, signal handling
- **`config.rs`** - Clap-based CLI configuration with version filtering options
- **`types.rs`** - Core data structures and version comparison logic
- **`scanner.rs`** - Main scanning orchestration, progress bars, kubectl execution

### Key Components

#### Version Filtering System (`types.rs`)

The `Version` struct handles semantic versioning with special support for Java 8 build numbers:

```rust
Version {
    major: u32,    // 1, 11, 17, 20
    minor: u32,    // 8, 0, 3
    patch: u32,    // 0, 16
    build: u32,    // 232, 292, 342 (Java 8 build numbers like 1.8.0_292)
}
```

**Comparison order**: major → minor → patch → build

**Critical**: The `build` field is essential for Java 8 version comparison. When parsing versions like `1.8.0_292`, split by underscore first, then by dot.

#### Kubernetes Pod Structure

**Important**: `ownerReferences` is located in `metadata`, not at the Pod level:

```rust
Pod {
    metadata: PodMetadata {
        name: String,
        namespace: String,
        owner_references: Vec<OwnerReference>,  // ← Inside metadata!
    }
}
```

This matches the actual Kubernetes API structure. Incorrect placement causes DaemonSet filtering to fail silently.

#### Concurrent Scanning Architecture

**Race condition prevention**: The progress bar length must be set BEFORE starting any async scans:

```rust
// CORRECT order:
let pods_to_scan = filter_pods();
progress_bar.set_length(pods_to_scan.len());  // ← Set first
progress_bar.tick();                           // ← Force update
start_scanning(pods_to_scan);                  // ← Then scan

// WRONG: Setting length after scans start causes jumps (e.g., 348/376)
```

All namespaces are scanned concurrently via `buffer_unordered()`. Each namespace has its own progress bar. The scanner uses `Arc<Mutex<ScanResult>>` for thread-safe result aggregation.

#### Version Detection

- **Java**: Execute `kubectl exec -- java -version`, parse stderr using regex `version "([^"]+)"`
- **Node.js**: Execute `kubectl exec -- node --version`, parse stdout using regex `v(\d+\.\d+\.\d+)`

Both commands run **concurrently** for each pod with configurable timeout (default 30s).

#### Filtering Logic

When version filters are applied (e.g., `--min-java-version 15`):
- Only pods with versions **below** the threshold are included
- Pods without the runtime are excluded (not counted)
- Empty namespaces still appear in summary table with 0 counts

**Error handling**: Even if scanning fails, namespace stats are recorded with zero counts to prevent namespace omission.

### CSV Export Format

The CSV export has three sections:

1. **Pod List**: `INDEX,NAMESPACE,POD,JAVA_VERSION,NODE_VERSION`
2. **Namespace Summary**: Table with per-namespace statistics (new section added after pod list)
3. **Overall Summary**: Comments with total counts

All three sections are written to the same CSV file, separated by blank lines.

## Testing Patterns

When adding tests:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_parse() {
        // Test all version formats: "17", "20.3", "11.0.16", "1.8.0_292"
        // Java 8 build numbers are critical edge cases
    }
}
```

**Test placement**: Unit tests go in the same file using `#[cfg(test)]` module.

## Common Pitfalls

1. **Progress bar initialization**: Always initialize with correct length BEFORE async work starts
2. **ownerReferences location**: Must be in `metadata`, not at Pod root level
3. **Java 8 parsing**: Remember to split by underscore (`_`) before splitting by dot (`.`)
4. **Version filtering**: Filter criteria is "less than" (below threshold), not "greater than"
5. **Error handling**: Always record namespace stats even on errors to prevent omission

## Binary Naming

- **Binary name**: `podver` (defined in Cargo.toml `[[bin]]` section and Makefile)
- **Package name**: `podver` (for cargo metadata)
- All build outputs should reference `podver`
