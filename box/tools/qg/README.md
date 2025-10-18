# qg

[![Rust Version](https://img.shields.io/badge/Rust-1.75+-orange?style=flat-square&logo=rust&color=black&logoColor=white)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&color=black)](https://opensource.org/licenses/MIT)

A simple QR code generator that creates a QR code from a given URL. Written in Rust for performance and reliability.

## Features

- Generate QR codes from URLs
- Customizable size (width and height)
- PNG output format
- Quiet mode for scripting
- URL validation (http:// or https:// required)

## Installation

### Prerequisites

- Rust 1.75 or later
- Cargo (comes with Rust)

### Build from source

```bash
# Clone the repository (if not already cloned)
git clone https://github.com/younsl/o.git
cd o/box/tools/qg

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
# Generate QR code from URL
./target/release/qg https://github.com/

# Or if installed
qg https://github.com/
```

### Custom options

```bash
# Custom filename
qg --filename my-qr.png https://example.com

# Custom size
qg --width 200 --height 200 https://example.com

# Quiet mode (no output messages)
qg --quiet https://example.com

# All options combined
qg --quiet --width 300 --height 300 --filename custom.png https://github.com/
```

### Example output

```console
$ qg https://github.com/
QR code saved as qrcode.png.
Address: https://github.com/. Size: 100x100
```

### Help

```bash
qg --help
```

### Makefile targets

```bash
make build      # Build debug version
make release    # Build optimized release version
make run        # Run with example URL
make test       # Run tests
make fmt        # Format code with rustfmt
make lint       # Run clippy linter
make clean      # Remove build artifacts
make install    # Install to ~/.cargo/bin/
```

## Dependencies

Major crates used:
- `clap` - Command-line argument parsing
- `qrcode` - QR code generation
- `image` - Image encoding
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
