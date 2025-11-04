# Testing Guide

This document describes the testing strategy and structure for promdrop.

## Test Organization

```
tests/
├── fixtures/
│   └── sample-metrics.json          # Mock input data (12 metrics, 3 jobs)
├── integration_test.rs              # Basic integration tests (3 tests)
└── e2e_conversion_test.rs           # End-to-end conversion tests (4 tests)
```

## Running Tests

```bash
# Run all tests
cargo test

# Run specific test file
cargo test --test e2e_conversion_test

# Run specific test function
cargo test test_full_conversion_pipeline

# Run with output
cargo test -- --nocapture

# Run unit tests only
cargo test --lib
```

## Test Categories

### Unit Tests

Located in source files using `#[cfg(test)]` modules:

- `src/parser.rs`: JSON parsing logic
- `src/grouper.rs`: Metric grouping and prefix extraction
- `src/output.rs`: File generation utilities

**Example**:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_group_metrics_by_prefix() {
        // Test implementation
    }
}
```

### Integration Tests

**File**: `tests/integration_test.rs`

Validates the structure and content of mock data:

1. **test_end_to_end_conversion**
   - Verifies JSON file exists and is readable
   - Validates JSON structure (additional_metric_counts field)
   - Confirms 12 metrics are present

2. **test_sample_data_has_multiple_jobs**
   - Checks for 3 unique jobs (api-server, web-server, db-proxy)
   - Validates job name extraction

3. **test_sample_data_metric_distribution**
   - Verifies metrics are distributed across jobs
   - Ensures api-server has at least 7 metrics

### End-to-End Tests

**File**: `tests/e2e_conversion_test.rs`

Tests the complete conversion pipeline from input to output:

1. **test_full_conversion_pipeline**
   - Parse JSON input file
   - Group metrics by prefix
   - Generate relabel rules
   - Create summary.txt file
   - Generate per-job .txt files
   - Generate combined YAML configuration
   - Validate all output files

2. **test_metric_prefix_grouping**
   - Verify metrics are grouped by prefix (http_, grpc_, database_, etc.)
   - Check "other" group for metrics without prefix
   - Validate prefix extraction logic

3. **test_yaml_parsing_validity**
   - Ensure generated YAML is syntactically valid
   - Parse YAML back into JobConfig structures
   - Verify summary comments are included

4. **test_empty_metrics_handling**
   - Test behavior with empty input file
   - Verify graceful handling of edge cases

## Mock Data Structure

The test fixture `sample-metrics.json` contains:

- **Total metrics**: 12 unique metrics
- **Jobs**: 3 jobs (api-server, web-server, db-proxy)
- **Metric distribution**:
  - api-server: 10 metrics (http_*, grpc_*, database_*, memory_*, cpu_*, orphan_*)
  - web-server: 5 metrics (cache_*, http_*, memory_*, cpu_*)
  - db-proxy: 2 metrics (database_*, memory_*)

## Conversion Pipeline Flow

```
Input: sample-metrics.json
  ↓
[Parser] Parse JSON and deduplicate metrics
  ↓
ParsedMetrics {
  job_metrics_map: HashMap<String, Vec<String>>
  summary_data: Vec<JobMetricSummary>
  total_unique_metrics: usize
}
  ↓
[Grouper] Group by prefix (text before first underscore)
  ↓
Groups: HashMap<String, Vec<String>>
  ↓
[Generator] Create regex patterns and relabel rules
  ↓
RelabelRules: Vec<RelabelRule>
  ↓
[Output] Generate files
  ↓
Output files:
├── unused/summary.txt
├── unused/api-server_unused_metrics.txt
├── unused/web-server_unused_metrics.txt
├── unused/db-proxy_unused_metrics.txt
└── combined_relabel_configs.yaml
```

## Validation Checklist

The test suite validates:

- JSON parsing accuracy
- Duplicate metric removal
- Job-based metric classification
- Prefix-based grouping
- Regex pattern generation
- YAML syntax validity
- File creation completion
- Empty data handling
- Edge case scenarios

## Test Coverage

Current test coverage:

```
Unit tests (src/):        5 tests
Integration tests:        3 tests
End-to-end tests:         4 tests
Binary tests:             5 tests
──────────────────────────────────
Total:                   17 tests
```

## Adding New Tests

### For new features

1. Add unit tests in the relevant source file:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_feature() {
        // Test implementation
    }
}
```

2. Add integration tests if testing multiple modules:
```rust
// tests/integration_test.rs or new test file
#[test]
fn test_feature_integration() {
    // Test implementation
}
```

### For bug fixes

1. Write a failing test that reproduces the bug
2. Fix the bug
3. Verify the test passes

## Continuous Integration

For CI/CD pipelines:

```bash
# Fail on warnings
RUSTFLAGS="-D warnings" cargo test

# Run clippy with strict checks
cargo clippy -- -D warnings

# Check formatting
cargo fmt --check
```

## Troubleshooting Tests

### Test fails with "file not found"

Ensure you're running from the project root:
```bash
cd /path/to/promdrop
cargo test
```

### Tests pass but warnings appear

Fix warnings with:
```bash
# Auto-fix some warnings
cargo fix --tests

# Check what clippy suggests
cargo clippy --tests
```

### Need to see test output

Run with nocapture flag:
```bash
cargo test -- --nocapture
```

## Best Practices

1. Keep tests focused and independent
2. Use descriptive test names
3. Clean up temporary files (use `tempfile` crate)
4. Mock external dependencies
5. Test both success and failure paths
6. Document complex test scenarios
7. Maintain test fixtures separately
