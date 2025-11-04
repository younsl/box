# kk

[![GitHub release](https://img.shields.io/github/v/release/younsl/o?filter=kk*&style=flat-square&color=black)](https://github.com/younsl/o/releases?q=kk&expanded=true)
[![Rust](https://img.shields.io/badge/rust-1.91-black?style=flat-square&logo=rust&logoColor=white)](https://www.rust-lang.org/)
[![GitHub license](https://img.shields.io/github/license/younsl/o?style=flat-square&color=black)](https://github.com/younsl/o/blob/main/LICENSE)

**kk** (knock-knock) checks domain connectivity from a YAML config. Fast, concurrent, and reliable.

## Features

- Concurrent domain checks with automatic retries
- Auto-adds HTTPS prefix to bare domains
- Clean table output with response times
- Verbose logging for debugging

## Quick Start

```bash
# Run with example config
make run

# Or build and run manually
make release
./target/release/kk --config configs/domain-example.yaml
```

## Installation

**Prerequisites**: Rust 1.91+ (Edition 2024)

```bash
# Build release binary
make release

# Install to ~/.cargo/bin/
make install
```

## Usage

```bash
kk --config configs/domain-example.yaml

# Enable verbose logging
kk --config configs/domain-example.yaml --verbose
```

### Output

```console
┌─────────────────────────────────┬──────┬─────────────────┬──────┬──────────────┐
│ URL                             │ TIME │ STATUS          │ CODE │ ATTEMPTS     │
├─────────────────────────────────┼──────┼─────────────────┼──────┼──────────────┤
│ https://www.github.com          │ 205ms│ OK              │ 200  │ 1            │
│ https://registry.k8s.io/v2/     │ 237ms│ OK              │ 200  │ 1            │
│ https://www.google.com          │ 401ms│ OK              │ 200  │ 1            │
│ https://reddit.com              │ 324ms│ UNEXPECTED_CODE │ 403  │ 3 (failed)   │
└─────────────────────────────────┴──────┴─────────────────┴──────┴──────────────┘

Summary: 3/4 successful checks in 2.1s
```

## Configuration

```yaml
# configs/domain-example.yaml
domains:
  - https://www.google.com       # Full URL
  - https://registry.k8s.io/v2/
  - reddit.com                   # Auto-prefixes with https://
  - www.github.com
```

## Development

```bash
make build      # Debug build
make release    # Release build
make test       # Run tests
make fmt        # Format code
make lint       # Run clippy
make clean      # Clean artifacts
```

## Built With

- **Rust 1.91+ Edition 2024** - Memory safety, zero-cost abstractions
- **Tokio** - Async runtime for concurrent checks
- **Clap** - CLI argument parsing
- **Reqwest** - HTTP client
- **Tabled** - Table formatting

## License

MIT
