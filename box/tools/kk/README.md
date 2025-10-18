# kk

[![Rust Version](https://img.shields.io/badge/Rust-1.75+-orange?style=flat-square&logo=rust&color=black&logoColor=white)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&color=black)](https://opensource.org/licenses/MIT)

**kk** stands for **knock-knock**.

A CLI tool that checks the status of domains specified in a YAML configuration file. Written in Rust for performance and reliability.

## Features

- ✅ Concurrent domain connectivity checks
- ✅ Automatic HTTPS prefix addition
- ✅ Configurable retry logic (3 attempts with 2s interval)
- ✅ Beautiful progress bar and table output
- ✅ Support for both domains and full URLs
- ✅ Verbose logging mode

## Project Structure

Following Rust best practices:

```
kk/
├── src/
│   ├── main.rs       # CLI entry point
│   ├── config.rs     # YAML configuration parsing
│   └── checker.rs    # HTTP connectivity checking logic
├── configs/
│   └── domain-example.yaml
├── Cargo.toml        # Rust dependencies and metadata
├── Makefile          # Build automation
└── README.md
```

## Installation

### Prerequisites

- Rust 1.75 or later
- Cargo (comes with Rust)

### Build from source

```bash
# Clone the repository (if not already cloned)
git clone https://github.com/younsl/o.git
cd o/box/tools/kk

# Build the project
make build

# Or build optimized release version
make release
```

### Install to system

```bash
# Install to ~/.cargo/bin/
make install

# Or use cargo directly
cargo install --path .
```

## Usage

### Basic usage

```bash
# Using the binary directly
./target/release/kk --config configs/domain-example.yaml

# Or if installed
kk --config configs/domain-example.yaml
```

### Example output

```console
$ kk --config configs/domain-example.yaml
Loaded domain list from 'configs/domain-example.yaml'.
┌─────────────────────────────────┬──────┬─────────────────┬──────┬──────────────┐
│ URL                             │ TIME │ STATUS          │ CODE │ ATTEMPTS     │
├─────────────────────────────────┼──────┼─────────────────┼──────┼──────────────┤
│ https://www.github.com          │ 205ms│ OK              │ 200  │ 1            │
│ https://registry.k8s.io/v2/     │ 237ms│ OK              │ 200  │ 1            │
│ https://www.google.com          │ 401ms│ OK              │ 200  │ 1            │
│ https://stackoverflow.com       │ 591ms│ OK              │ 200  │ 1            │
│ https://www.stackoverflow.com   │ 830ms│ OK              │ 200  │ 1            │
│ https://reddit.com              │ 324ms│ UNEXPECTED_CODE │ 403  │ 3 (failed)   │
└─────────────────────────────────┴──────┴─────────────────┴──────┴──────────────┘

Summary: 5/6 successful checks in 5.3s.
```

### Verbose mode

Enable debug logging:

```bash
kk --config configs/domain-example.yaml --verbose
```

### Makefile targets

```bash
make build      # Build debug version
make release    # Build optimized release version
make run        # Build and run with example config
make dev        # Run with verbose logging
make test       # Run tests
make fmt        # Format code with rustfmt
make lint       # Run clippy linter
make clean      # Remove build artifacts
make install    # Install to ~/.cargo/bin/
```

## Configuration

The configuration file should be in YAML format:

```yaml
# configs/domain-example.yaml
domains:
  # Full URLs with scheme
  - https://www.google.com
  - https://registry.k8s.io/v2/

  # Domain only (https:// will be added automatically)
  - reddit.com
  - www.github.com
```

## Technical Details

Built with Rust for:
- **Memory safety**: Compile-time borrow checker prevents memory bugs
- **Concurrency**: Tokio async/await for efficient parallel checks
- **Type safety**: Strong type system with Result<T, E> error handling
- **Performance**: Zero-cost abstractions and small binary size (~5MB stripped)

## Dependencies

Major crates used:
- `clap` - Command-line argument parsing
- `tokio` - Async runtime
- `reqwest` - HTTP client
- `serde` / `serde_yaml` - YAML parsing
- `indicatif` - Progress bars
- `tabled` - Table formatting
- `anyhow` - Error handling

## Testing

```bash
# Run all tests
make test

# Or use cargo directly
cargo test --verbose
```

## License

MIT License - See [LICENSE](../../LICENSE) for details
