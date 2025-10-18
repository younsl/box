# Java Version Scanner (jvs)

A Kubernetes tool to scan and report Java versions across pods in your cluster.

This tool connects to your Kubernetes cluster via `kubectl` and executes `java -version` inside each pod to detect installed Java versions. It processes multiple pods concurrently with real-time progress tracking, parses the Java version output using regex patterns, and presents the results in a clean kubectl-style format. Perfect for auditing Java versions across microservices, identifying legacy JDK installations, or ensuring compliance with security policies.

## Features

- **Real-time Progress Tracking** - Multi-level progress bars showing namespace and pod-level scanning progress
- **Concurrent Scanning** - Parallel processing across multiple namespaces with configurable concurrency
- **Namespace Statistics** - Per-namespace summary with total pods, JDK usage count, and adoption ratio
- **CSV Export** - Export results to CSV for analysis in Excel or Google Sheets
- **DaemonSet Filtering** - Optionally skip DaemonSet pods
- **Async/Parallel Processing** - Built with Tokio for efficient resource utilization
- **Configurable Timeouts** - Adjust kubectl command timeouts for slow-responding pods
- **Graceful Shutdown** - Handles Ctrl+C for clean cancellation
- **Structured Logging** - Verbose mode with tracing for debugging

## Prerequisites

- Rust 1.75+ (for building from source)
- `kubectl` installed and configured
- Access to a Kubernetes cluster
- Pods must have Java installed (`java -version` should work)

## Installation

### From Source

```bash
# Build release binary
cargo build --release

# Binary will be at target/release/jvs
```

### Using Cargo

```bash
cargo install --path .
```

## Quick Start

### Basic Usage

Scan the default namespace:

```bash
jvs --namespaces default
```

**Output:**
```
INDEX  NAMESPACE  POD                         JAVA_VERSION
1      default    spring-boot-app-7d8f9c-xyz  17.0.8
2      default    legacy-service-5b6c7-abc    1.8.0_372

NAMESPACE  TOTAL PODS  JDK PODS  RATIO   TIME
default    45          2         4.4%    0m 5s
```

### Scan Multiple Namespaces

```bash
jvs --namespaces production,staging,development
```

**Output with Real-time Progress:**
```
 ✓ [████████████████████████████████████████] 3/3 namespaces | Scanning...

INDEX  NAMESPACE    POD                        JAVA_VERSION
1      production   api-gateway-6f7d8-qwer     11.0.19
2      production   payment-service-9c4b-zxcv  17.0.8
3      staging      api-gateway-5e6c7-asdf     11.0.19
4      development  test-app-3a2b1-hjkl        1.8.0_372

NAMESPACE    TOTAL PODS  JDK PODS  RATIO   TIME
development  12          1         8.3%    0m 8s
production   78          2         2.6%    0m 8s
staging      34          1         2.9%    0m 8s
```

### Export to CSV

```bash
jvs --namespaces production --output results.csv
```

**CSV Output:**
```csv
INDEX,NAMESPACE,POD,JAVA_VERSION
1,production,api-gateway-6f7d8-qwer,11.0.19
2,production,payment-service-9c4b-zxcv,17.0.8

# Scan Summary
# Total pods scanned: 78
# Pods using JDK: 2
# Time taken: 0m 8s
```

### Verbose Mode

```bash
jvs --verbose -n default
```

**Output:**
```
2025-10-18T04:15:32Z DEBUG Checking pods in namespace: default
2025-10-18T04:15:32Z DEBUG Scanning pod 1 of 5: app-deployment-abc123
2025-10-18T04:15:33Z DEBUG Scanning pod 2 of 5: worker-deployment-def456
...
INDEX  NAMESPACE  POD                    JAVA_VERSION
1      default    app-deployment-abc123  11.0.16
2      default    api-service-def456     17.0.2

NAMESPACE  TOTAL PODS  JDK PODS  RATIO   TIME
default    5           2         40.0%   0m 3s
```

## Usage Examples

### Include DaemonSet Pods

By default, DaemonSet pods are skipped. To include them:

```bash
jvs --skip-daemonset=false
```

### Increase Concurrency

Process more pods in parallel:

```bash
jvs --max-concurrent 50
```

### Custom Timeout

Increase timeout for slow-responding pods:

```bash
jvs --timeout 60
```

### Complete Example

```bash
jvs \
  --namespaces production,staging \
  --max-concurrent 30 \
  --timeout 45 \
  --skip-daemonset=false \
  --verbose \
  --output production-java-audit.csv
```

## Command Options

All available command-line options for `jvs` (also viewable with `jvs --help`):

| Option | Short | Type | Required | Default | Description |
|--------|-------|------|----------|---------|-------------|
| `--namespaces` | `-n` | String | **Yes** | - | Comma-separated list of namespaces to scan |
| `--max-concurrent` | `-c` | Integer | No | `20` | Maximum number of concurrent tasks |
| `--timeout` | `-t` | Integer | No | `30` | Timeout for kubectl commands (seconds) |
| `--skip-daemonset` | `-s` | Flag | No | `true` | Skip DaemonSet pods |
| `--verbose` | `-v` | Flag | No | `false` | Enable verbose logging |
| `--output` | `-o` | String | No | - | CSV filename to export results (e.g., `results.csv`) |
| `--help` | `-h` | Flag | No | - | Print help message |
| `--version` | `-V` | Flag | No | - | Print version information |

## Architecture

Built with modern Rust ecosystem:

- **Tokio** - Async runtime for concurrent operations, handles process spawning for kubectl commands, manages timeouts and graceful shutdown
- **Clap v4** - CLI parsing with derive macros, automatic help generation, type-safe argument validation
- **Serde + serde_json** - Zero-copy JSON parsing for Kubernetes API responses, type-safe struct mapping for Pod metadata
- **Tracing** - Structured logging with multiple verbosity levels, environment-based filtering for better debugging
- **Regex** - Parses Java version strings using pattern `version "([^"]+)"`, handles various JDK formats (OpenJDK, Oracle, Corretto, etc.)
- **Futures** - Stream combinators (`buffer_unordered`) for controlled parallelism across pod scanning
- **Anyhow** - Simplified error handling with context propagation and human-readable messages
- **Indicatif** - Multi-level progress bars with real-time feedback for namespace and pod scanning

## How It Works

1. **Fetch pods** from specified namespaces using `kubectl get pods -o json`
2. **Filter** DaemonSet pods (if enabled)
3. **Display progress** with multi-level progress bars (namespace + per-pod)
4. **Execute** `kubectl exec -- java -version` concurrently for each pod
5. **Parse** Java version from stderr output using regex
6. **Display** results in a kubectl-style formatted table
7. **Calculate statistics** - Per-namespace totals, JDK pod counts, and adoption ratios
8. **Export CSV** (optional) - Save results for further analysis

## Output Format

### Main Table
- **INDEX** - Sequential number for each Java pod found
- **NAMESPACE** - Kubernetes namespace
- **POD** - Pod name
- **JAVA_VERSION** - Detected Java version (e.g., "11.0.19", "17.0.8")

### Namespace Summary
- **NAMESPACE** - Kubernetes namespace (alphabetically sorted)
- **TOTAL PODS** - Total number of pods scanned in the namespace
- **JDK PODS** - Number of pods with Java detected
- **RATIO** - Percentage of pods using Java (JDK PODS / TOTAL PODS × 100)
- **TIME** - Total scan duration in minutes and seconds

## Troubleshooting

### "Unknown" Version

If a pod shows "Unknown" version:
- Pod may not have Java installed
- Java binary may be in a non-standard location (try: `which java` in the pod)
- Timeout may be too short for pod to respond
- Pod may not be in Running state
- Container may not have shell access

Try increasing timeout or enabling verbose mode:

```bash
jvs --timeout 60 --verbose
```

### No Progress Bars Visible

Progress bars are sent to stderr. If you're redirecting output:

```bash
jvs -n production 2>/dev/null  # Hides progress bars
jvs -n production > output.txt  # Progress bars still visible
```

### Slow Scanning

To improve performance:
- Increase concurrency: `--max-concurrent 50`
- Decrease timeout for faster failure: `--timeout 10`
- Skip DaemonSets: `--skip-daemonset` (default)

## License

MIT
