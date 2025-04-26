[![Go version](https://img.shields.io/badge/Go-1.24-black?style=flat-square)](./go.mod) ![License](https://img.shields.io/badge/License-MIT-black?style=flat-square)

# ol

> *ol stands for one-liner.*

ol is a simple command-line tool to quickly jot down one-liner logs or notes.

## Features

*   Appends timestamped messages to a yearly log file.
*   Simple configuration via YAML.
*   Cross-platform.

## Installation

In the root of the repository, build binary for your platform:

```bash
make build
```

## Basic Usage

To record a message:

```bash
ol "Your message for the day"
```

This command appends a timestamped log entry to `ol-YYYY.txt` (where `YYYY` is the current year). The log format is:

```bash
YYYY-MM-DD HH:MM:SS TZ | Message
```

By default, log files are stored in:

*   `$XDG_DATA_HOME/ol/` (if `$XDG_DATA_HOME` is set)
*   `$HOME/.local/share/ol/` (otherwise)

This location can be changed via configuration.

## Configuration

`ol` uses a configuration file located at:

*   `$XDG_CONFIG_HOME/ol/config.yaml` (if `$XDG_CONFIG_HOME` is set)
*   `$HOME/.config/ol/config.yaml` (otherwise)

**Available Keys:**

*   `timezone`: (String, Optional) Sets the timezone for log timestamps. Uses IANA Time Zone names (e.g., `America/New_York`, `Asia/Seoul`). Defaults to `UTC`. See [List of tz database time zones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).
*   `dataDirectory`: (String, Optional) Specifies the directory to store log files (like `ol-YYYY.txt`). Tilde (~) and environment variables (e.g., `$HOME`) are expanded. Defaults to the standard user data directory mentioned above.

**Example `config.yaml`:**

```yaml
timezone: "Asia/Seoul"
# dataDirectory: "~/MyLogs/ol" # Example using tilde expansion
dataDirectory: "$HOME/MyLogs/ol" # Example using environment variable
```

**Managing Configuration:**

*   Initialize a default config file:
    ```bash
    ol config init
    ```
*   Show current configuration:
    ```bash
    ol config get
    ```
*   Set a configuration value:
    ```bash
    ol config set timezone Asia/Tokyo
    ol config set dataDirectory "/path/to/your/logs"
    ol config set dataDirectory "~/Documents/OlLogs"
    ```

## Development

```bash
# Navigate to the tool's directory if needed
# cd ./tools/ol 
go build ./cmd/ol
```

## License

MIT License
