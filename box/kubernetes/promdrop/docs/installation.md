# Installation Guide

This guide covers different methods to install Promdrop (Rust version).

## Requirements

### For Pre-built Binaries
- Linux or macOS operating system
- `curl` or `wget` for downloading
- `tar` for extracting archives

### For Building from Source
- Rust 1.70 or higher (install via [rustup](https://rustup.rs/))
- Cargo (comes with Rust installation)
- Make utility installed on your system (optional, for convenience)

## Installation Methods

1. **Pre-built Binaries** - Download and install ready-to-use binaries for your platform
2. **Building from Source** - Compile Promdrop from source code

## Method 1: Pre-built Binaries

Download and install a pre-built binary for your platform:

```bash
# Set the version you want to install
VERSION="0.1.0"  # Change this to your desired version

# Get arch and os currently running on the machine
ARCH=$(arch)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Download the release
curl -LO https://github.com/younsl/promdrop/releases/download/${VERSION}/promdrop-${OS}-${ARCH}.tar.gz

# Extract the binary
tar -xzf promdrop-${OS}-${ARCH}.tar.gz

# Make it executable
chmod +x promdrop-${OS}-${ARCH}

# Move to system path
sudo mv promdrop-${OS}-${ARCH} /usr/local/bin/promdrop

# Clean up
rm promdrop-${OS}-${ARCH}.tar.gz
```

Check the [releases page](https://github.com/younsl/promdrop/releases) for available versions and platforms.

## Method 2: Building from Source

### Using Cargo (recommended)

```bash
# Clone the repository
git clone https://github.com/younsl/o.git
cd o/box/kubernetes/promdrop

# Build release version
cargo build --release

# Binary will be at target/release/promdrop
./target/release/promdrop --version
```

### Using Make

```bash
# Navigate to promdrop directory
cd o/box/kubernetes/promdrop

# Build debug version
make build

# Or build optimized release version
make release

# Binary will be at target/release/promdrop
./target/release/promdrop --version
```

### Installing to System

```bash
# Install to ~/.cargo/bin/ (make sure it's in your PATH)
cargo install --path .

# Or use make
make install
```

## Verifying Installation

After installation, you can verify that Promdrop is working correctly:

```bash
promdrop --help
```

This should display the help message with available commands and flags.