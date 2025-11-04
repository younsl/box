# s3vget

[![Rust](https://img.shields.io/badge/rust-1.91-black?style=flat-square&logo=rust&logoColor=white)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&color=black)](https://opensource.org/licenses/MIT)

S3 Object Version Downloader with interactive prompts and configurable timezone support.

## Features

- Interactive prompts for bucket, key, and date range
- Configurable timezone support (default: Asia/Seoul)
- Multiple date format support (YYYY-MM-DD, YYYY/MM/DD, etc.)
- Support for 'now' keyword
- Structured logging with tracing
- Pagination for large version lists

## Quick Start

```bash
# Interactive mode
make run

# Or build and run manually
make release
./target/release/s3vget
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

### Interactive Mode

```bash
s3vget

# With debug logging
make dev
```

### Command-line Arguments

```bash
# All parameters via CLI
s3vget \
  --bucket my-bucket \
  --key path/to/file.json \
  --start 2025-10-21 \
  --end 2025-10-22 \
  --timezone America/New_York

# Download all versions without filtering
s3vget --bucket my-bucket --key path/to/file.json --no-interactive

# Use 'now' as end date
s3vget -b my-bucket -k path/to/file.json -s 2025-10-01 -e now

# Different timezone
s3vget -b my-bucket -k path/to/file.json -z UTC
```

## Timezone Support

The `--timezone` (or `-z`) option accepts IANA timezone names:

```bash
# Asia
-z Asia/Seoul          # KST (default)
-z Asia/Tokyo          # JST
-z Asia/Shanghai       # CST

# America
-z America/New_York     # EST/EDT
-z America/Los_Angeles  # PST/PDT

# Europe
-z Europe/London        # GMT/BST
-z Europe/Paris         # CET/CEST

# UTC
-z UTC
```

## Date Formats

Supported formats:
- `YYYY-MM-DD` (e.g., `2025-10-21`)
- `YYYY/MM/DD` (e.g., `2025/10/21`)
- `YYYY.MM.DD` (e.g., `2025.10.21`)
- `YYYYMMDD` (e.g., `20251021`)
- `now` (current time)

## Output Format

Downloaded files are named as:

```
{version_number}_{timestamp}_{original_filename}.{ext}
```

Example:
```
001_20251021_143022_config.json
002_20251022_091544_config.json
003_20251023_165511_config.json
```

## CLI Options

All options are optional. If not provided, s3vget will prompt interactively or use default values.

| Option | Type | Description | Example |
|--------|------|-------------|---------|
| `-b, --bucket` | String | S3 bucket name | `--bucket my-bucket` |
| `-k, --key` | String | S3 object key (path) | `--key path/to/file.json` |
| `-o, --output-dir` | Path | Output directory (default: `versions`) | `--output-dir ./downloads` |
| `-s, --start` | Date/String | Start date for filtering versions | `--start 2025-10-21` or `--start now` |
| `-e, --end` | Date/String | End date for filtering versions | `--end 2025-10-22` or `--end now` |
| `-z, --timezone` | String | Timezone for date interpretation and output display (default: `Asia/Seoul`). Required because S3 stores timestamps in UTC, but users need to filter and view versions in their local timezone. | `--timezone UTC` or `-z America/New_York` |
| `--no-interactive` | Flag | Skip interactive prompts | `--no-interactive` |
| `--log-level` | String | Log level: trace, debug, info, warn, error (default: `info`) | `--log-level debug` |
| `-h, --help` | Flag | Print help information | `--help` |
| `-V, --version` | Flag | Print version information | `--version` |

## AWS Authentication

s3vget uses AWS SDK authentication methods:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. AWS credentials file (`~/.aws/credentials`)
3. IAM instance profile (EC2)
4. ECS task role

Required permissions:
- `s3:GetObject`
- `s3:GetObjectVersion`
- `s3:ListBucket`
- `s3:ListBucketVersions`

## Development

```bash
make build      # Debug build
make release    # Release build
make test       # Run tests
make fmt        # Format code
make lint       # Run clippy
make check      # Check without building
make deps       # Update dependencies
make clean      # Clean artifacts
```

## Built With

- **Rust 1.91+ Edition 2024** - Memory safety, performance
- **Tokio** - Async runtime
- **AWS SDK** - S3 operations
- **Clap** - CLI parsing
- **Tracing** - Structured logging
- **Chrono-tz** - Timezone handling

## License

MIT
