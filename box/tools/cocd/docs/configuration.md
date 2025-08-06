# Configuration

## Setup

cocd uses YAML configuration files and environment variables. Configuration files are loaded from the following locations in order (first found is used):

1. Current directory (`./config.yaml`)
2. Home directory (`~/.cocd/config.yaml`) 
3. System directory (`/etc/cocd/config.yaml`)

For more details on the configuration loading implementation, see [internal/config/config.go](../internal/config/config.go).

## Configuration Format

For more config examples, see [config-example.yaml](../config-example.yaml).

```yaml
# GitHub API configuration
github:
  # GitHub personal access token for API authentication
  # Can also be set via GITHUB_TOKEN or COCD_GITHUB_TOKEN environment variables
  # If not set, will attempt to use 'gh auth token' from GitHub CLI
  token: ""
  
  # GitHub API base URL
  # Use default for GitHub.com or custom URL for GitHub Enterprise Server
  # Default: https://api.github.com
  base_url: "https://api.github.com"
  
  # GitHub organization name to monitor
  # Required: specify the organization containing your repositories
  org: "your-organization"
  
  # GitHub repository name to monitor  
  # Optional: if not specified, will monitor organization-wide deployments
  # repo: "your-repository"

# Monitoring behavior configuration
monitor:
  # Refresh interval for scanning deployments
  # How often to check for new deployment approvals (in seconds)
  # Default: 5 seconds
  interval: 5
  
  # Target deployment environment to monitor
  # Filter deployments by environment name (e.g., prod, staging, dev)
  # Default: prod
  environment: "prod"
  
  # Timezone for displaying approval timestamps
  # Used for converting UTC timestamps to local time
  # Default: UTC
  # Examples: UTC, Asia/Seoul, America/New_York, Europe/London, Asia/Tokyo
  timezone: "UTC"
```

## Environment Variables

All configuration options can be overridden with `COCD_` prefixed environment variables:

```bash
export COCD_GITHUB_TOKEN="your-token"
export COCD_GITHUB_ORG="your-org"
export COCD_GITHUB_REPO="your-repo"
export COCD_MONITOR_INTERVAL=10
export COCD_MONITOR_ENVIRONMENT="staging"
export COCD_MONITOR_TIMEZONE="Asia/Seoul"
```

## Authentication

GitHub token can be provided in three ways (in order of precedence):

1. Configuration file: `github.token`
2. Environment variable: `GITHUB_TOKEN` or `COCD_GITHUB_TOKEN`
3. GitHub CLI: `gh auth token` (requires `gh auth login`)

If both `token` and `GITHUB_TOKEN` environment variable are omitted, cocd will automatically attempt to obtain a Personal Access Token (PAT) through the local GitHub CLI's `gh auth token` command.

## Examples

### Basic Setup

```yaml
github:
  token: ghp_xxxxxxxxxxxx
  org: my_org

monitor:
  interval: 30
  environment: production
  timezone: Asia/Seoul
```

### Enterprise Server GitHub

If omitted github.repo value in config file, cocd scans organization-wide scope:

```yaml
github:
  token: ghp_xxxxxxxxxxxx
  base_url: "https://github.company.com/api/v3"
  org: engineering

monitor:
  interval: 30
  environment: production
  timezone: Asia/Seoul
```
