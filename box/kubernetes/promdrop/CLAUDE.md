# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Promdrop is a Go CLI tool that generates Prometheus `metric_relabel_configs` to drop unused metrics, helping reduce monitoring costs. It processes JSON output from Grafana Mimirtool and creates optimized YAML configurations for Prometheus jobs.

## Development Commands

### Build & Run
```bash
# Build binary for current platform
make build

# Run the built binary
./bin/promdrop generate -f prometheus-metrics.json

# Cross-platform compilation (Linux/macOS, amd64/arm64)
make build-all

# Clean build artifacts
make clean
```

### Code Quality
```bash
# Format Go code
make fmt

# Run tests (when implemented)
make test
```

### Docker
```bash
# Build container image
docker build -t promdrop:latest .

# Run in container
docker run --rm -v $(pwd):/data promdrop:latest generate -f /data/prometheus-metrics.json
```

## Architecture

### Core Structure
- **cmd/promdrop/**: CLI entry point using Cobra framework
  - `main.go`: Application bootstrap
  - `generate.go`: Main command implementation
  
- **internal/**: Business logic (unexported packages)
  - `parser/`: JSON parsing for Mimirtool output
  - `report/`: Report generation and metric grouping
  - `utils/`: File I/O and utility functions

### Key Design Patterns
- Command pattern via Cobra for CLI interface
- Modular package design with clear separation of concerns
- Prefix-based metric grouping for efficient regex patterns
- YAML generation for Prometheus configuration

### Data Flow
1. Read `prometheus-metrics.json` from Mimirtool
2. Parse JSON structure: `data.metricsUsage.additional[].metrics[]`
3. Group metrics by Prometheus job
4. Generate prefix-based regex patterns for efficiency
5. Output YAML configs and text files by job

## Critical Implementation Details

### Metric Processing
- Metrics are grouped by common prefixes to create efficient regex patterns
- Single metrics use exact match, multiple use regex with alternation
- Output files are named by job: `<job>_drop_relabel_config.yaml`

### Input Format
Expects JSON from Mimirtool with structure:
```json
{
  "data": {
    "metricsUsage": {
      "additional": [{
        "job": "job_name",
        "metrics": ["metric1", "metric2"]
      }]
    }
  }
}
```

### Output Format
Generates Prometheus relabel configs:
```yaml
metric_relabel_configs:
  - source_labels: [__name__]
    regex: 'prefix1_(metric1|metric2|metric3)'
    action: drop
```

## Development Workflow

### Adding Features
1. Implement logic in appropriate `internal/` package
2. Update command in `cmd/promdrop/`
3. Add tests for new functionality
4. Update documentation in `/docs/`

### Testing Changes
Currently no test files exist. When adding tests:
- Place unit tests alongside source files (*_test.go)
- Use Go's standard testing package
- Test core logic in `internal/` packages

### Release Process
- GitHub Actions automates releases on tag push
- Builds cross-platform binaries and container images
- Publishes to GitHub Container Registry

## Important Notes

- Go 1.25+ required for development
- Grafana Mimirtool required for generating input data
- Output optimization through prefix grouping is critical for performance
- YAML output must be valid Prometheus configuration