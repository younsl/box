# Installation Guide

This guide covers different methods to install Promdrop.

## Requirements

### For Pre-built Binaries
- Linux, macOS, or Windows operating system
- `curl` or `wget` for downloading
- `tar` for extracting archives

### For Building from Source
- Go 1.25 or higher
- Make utility installed on your system

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

Build promdrop binary file from source code:

```bash
# Navigate to the project root (younsl/promdrop)
make build
```

This command will create a `promdrop` binary file in your current directory.

## Verifying Installation

After installation, you can verify that Promdrop is working correctly:

```bash
promdrop --help
```

This should display the help message with available commands and flags.