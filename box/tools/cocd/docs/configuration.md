# Configuration

## Auto-Generated Configuration

🆕 **New in v0.3.0**: cocd automatically creates a default configuration file on first run!

When you run cocd for the first time, it will automatically create a configuration file at `$HOME/.config/cocd/config.yaml` (or `$XDG_CONFIG_HOME/cocd/config.yaml`) with helpful comments and sensible defaults.

The skeleton configuration includes:
- Detailed comments explaining each option
- Examples for common use cases
- Environment variable alternatives
- Authentication method options

## Setup

cocd uses YAML configuration files and environment variables. Configuration files are loaded from the following locations in order (first found is used):

1. XDG config directory (`$HOME/.config/cocd/config.yaml` or `$XDG_CONFIG_HOME/cocd/config.yaml`)
2. Legacy location (`~/.cocd/config.yaml`)
3. System directory (`/etc/cocd/config.yaml`)

For more details on the configuration loading implementation, see [internal/config/config.go](../internal/config/config.go) and [internal/config/skeleton.go](../internal/config/skeleton.go).

## Configuration Format

The auto-generated skeleton config provides a complete template with detailed comments. For additional examples, see [config-example.yaml](../config-example.yaml).

```yaml
# COCD Configuration File
# 
# This file was automatically generated. Please update it with your settings.
# You can also use environment variables to override these settings.
#
# For more information, see: https://github.com/younsl/box/tree/main/box/tools/cocd

# GitHub configuration
github:
  # GitHub token (can also be set via COCD_GITHUB_TOKEN or GITHUB_TOKEN env var)
  # You can also authenticate using 'gh auth login' command
  token: ""
  # GitHub API base URL (default: api.github.com)
  # For GitHub Enterprise Server, use: github.example.com/api/v3
  base_url: api.github.com
  # GitHub organization name (required)
  # Can also be set via COCD_GITHUB_ORG env var
  org: ""
  # GitHub repository name (optional)
  # If not specified, monitors all repositories in the organization
  # Can also be set via COCD_GITHUB_REPO env var
  repo: ""

# Monitor configuration
monitor:
  # Refresh interval in seconds (default: 5)
  interval: 5
  # Timezone for displaying timestamps (default: UTC)
  # Examples: UTC, Asia/Seoul, America/New_York, Europe/London, Asia/Tokyo
  timezone: UTC
```

## Environment Variables

All configuration options can be overridden with `COCD_` prefixed environment variables:

```bash
export COCD_GITHUB_TOKEN="your-token"
export COCD_GITHUB_ORG="your-org"
export COCD_GITHUB_REPO="your-repo"
export COCD_MONITOR_INTERVAL=10
export COCD_MONITOR_TIMEZONE="Asia/Seoul"
```

## Authentication

GitHub token can be provided in three ways (in order of precedence):

1. Configuration file: `github.token`
2. Environment variable: `GITHUB_TOKEN` or `COCD_GITHUB_TOKEN`
3. GitHub CLI: `gh auth token` (requires `gh auth login`)

If both `token` and `GITHUB_TOKEN` environment variable are omitted, cocd will automatically attempt to obtain a Personal Access Token (PAT) through the local GitHub CLI's `gh auth token` command.

## Skeleton Configuration Features

The auto-generated skeleton configuration includes several enhancements:

- **XDG Base Directory Specification**: Follows standard Linux config directory conventions
- **Detailed Comments**: Each field includes explanatory comments with examples
- **Environment Variable Hints**: Shows corresponding environment variable names
- **Multiple Authentication Options**: Documents token, env var, and GitHub CLI methods
- **Smart Defaults**: Provides sensible default values for immediate use
- **Path Resolution**: Automatically handles config directory creation with proper permissions

### Implementation Details

The skeleton configuration is implemented in [internal/config/skeleton.go](../internal/config/skeleton.go) with:

- `GetDefaultConfig()`: Returns default configuration values
- `CreateSkeletonConfig()`: Creates config file with comments and header
- `GetConfigDir()`: Resolves config directory following XDG specification
- `TryCreateDefaultConfig()`: Main function for auto-config generation

## Examples

### Basic Setup

```yaml
github:
  token: ghp_xxxxxxxxxxxx
  org: my_org

monitor:
  interval: 30
  timezone: Asia/Seoul
```

### First Run Experience

When running cocd for the first time, if config.yaml is not found in any of the predefined paths:

```bash
$ cocd
✅ Created default config file at: /home/user/.config/cocd/config.yaml

📝 Next steps:
   1. Edit the config file and set your GitHub organization
   2. Set your GitHub token using one of these methods:
      - Edit config file (github.token field)
      - Set environment variable: export COCD_GITHUB_TOKEN=<token>
      - Use GitHub CLI: gh auth login
```

### Enterprise GitHub Server

If github.repo is omitted, cocd scans all repositories in the specified organization:

```yaml
github:
  token: ghp_xxxxxxxxxxxx
  base_url: "github.company.com/api/v3"
  org: engineering

monitor:
  interval: 30
  timezone: Asia/Seoul
```
