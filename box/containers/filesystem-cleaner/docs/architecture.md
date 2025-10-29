# Architecture

This document explains how filesystem-cleaner is built and organized. You'll learn:

- **What each component does** - Clear responsibilities for every module
- **How they work together** - Data flow from CLI input to file deletion
- **Why this design** - Single responsibility principle in action

If you're contributing code, debugging, or just curious about the internals, start here.

## Design Philosophy

filesystem-cleaner follows the Unix philosophy: **"Do one thing and do it well"**. Each component has a single, well-defined responsibility.

## Component Overview

```
┌─────────────┐
│   main.rs   │  Entry point - CLI initialization & signal handling
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  cleaner.rs │  Orchestrator - Schedules cleanup & monitors disk usage
└──────┬──────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼
┌─────────────┐  ┌─────────────┐
│ matcher.rs  │  │ scanner.rs  │
│ Pattern     │  │ File system │
│ matching    │  │ traversal   │
└─────────────┘  └─────────────┘
```

## Components

### main.rs (91 lines)
**Responsibility**: Application entry point

- Parse CLI arguments
- Setup logging
- Handle shutdown signals (SIGTERM, SIGINT)
- Start the cleaner

**Dependencies**: `config`, `cleaner`

### config.rs (156 lines)
**Responsibility**: Configuration management

- Define CLI arguments with Clap
- Parse environment variables
- Validate configuration values
- Define `CleanupMode` enum (Once/Interval)

**Dependencies**: None

### matcher.rs (97 lines)
**Responsibility**: Pattern matching logic

Answers one question: *"Does this file path match the configured patterns?"*

**Key Methods**:
- `should_exclude(path) -> bool` - Check if path matches exclude patterns
- `should_include(path) -> bool` - Check if path matches include patterns

**Example**:
```rust
let matcher = PatternMatcher::new(
    &["*".to_string()],
    &["**/.git/**".to_string()]
)?;

matcher.should_exclude("project/.git/config")  // true
matcher.should_exclude("project/src/main.rs")  // false
```

**Dependencies**: `globset`

### scanner.rs (193 lines)
**Responsibility**: File system traversal

Walks directory trees and collects files based on pattern rules.

**Key Methods**:
- `scan(base_path) -> Vec<FileInfo>` - Collect all matching files
- `walk_directory()` - Recursively traverse directories

**How it works**:
1. Start from `base_path`
2. For each file/directory:
   - Calculate relative path
   - Ask `matcher` if it should be excluded
   - If directory and not excluded → recurse
   - If file and passes filters → collect
3. Return list of files to delete

**Example**:
```rust
let scanner = FileScanner::new(&matcher);
let files = scanner.scan("/home/runner/_work");
// Returns: Vec<FileInfo> with paths and sizes
```

**Dependencies**: `matcher`

### cleaner.rs (251 lines)
**Responsibility**: Cleanup orchestration

Coordinates all components to perform the actual cleanup operation.

**Key Responsibilities**:
- **Scheduling**: Run once or periodically based on `CleanupMode`
- **Disk monitoring**: Check if usage exceeds threshold
- **Coordination**: Use `scanner` to find files, then delete them
- **Logging**: Report cleanup progress and results

**Workflow**:
```
1. Check disk usage
   ↓
2. If > threshold:
   ├─> Create scanner with matcher
   ├─> Collect files to delete
   ├─> Delete files (or dry-run)
   └─> Log results (freed space, file count)
3. If interval mode:
   └─> Sleep and repeat
```

**Example**:
```rust
let cleaner = Cleaner::new(args)?;
cleaner.run().await?;  // Runs cleanup cycle(s)
```

**Dependencies**: `config`, `matcher`, `scanner`, `sysinfo`, `tokio`

## Data Flow

```
User → CLI Args → Config → Cleaner
                              ↓
                    ┌─────────┴─────────┐
                    ↓                   ↓
            Disk Monitor          Matcher + Scanner
                    ↓                   ↓
            Threshold Check       File Collection
                    ↓                   ↓
                    └─────────┬─────────┘
                              ↓
                        File Deletion
                              ↓
                      Logging & Results
```

## Design Principles

### 1. Single Responsibility Principle
Each component does **one thing only**:
- `matcher` → Pattern matching
- `scanner` → File traversal
- `cleaner` → Orchestration

### 2. Dependency Direction
```
cleaner → scanner → matcher
   ↓
config
```

Dependencies flow in one direction. Lower-level components (`matcher`, `scanner`) don't know about higher-level ones (`cleaner`).

### 3. Testability
Each component has its own unit tests:
- `matcher`: 5 tests for pattern matching edge cases
- `scanner`: 2 tests for file collection scenarios
- All tests run independently

### 4. Unix Philosophy
> "Write programs that do one thing and do it well. Write programs to work together."

- Small, focused modules (< 200 lines each)
- Clear interfaces between components
- Easy to understand, test, and modify

## Adding New Features

**Want to add a new pattern type?**
→ Modify `matcher.rs` only

**Want to change directory traversal logic?**
→ Modify `scanner.rs` only

**Want to add a new scheduling mode?**
→ Modify `cleaner.rs` only

Each change is **isolated to one component**, making the codebase easy to maintain and extend.

## Testing Strategy

**Unit Tests**: Each module (`matcher`, `scanner`) tests its own logic independently.

**Integration**: `cleaner` tests that components work together correctly.

**Example**: Testing glob patterns
```rust
// matcher_tests.rs
#[test]
fn test_nested_glob_patterns() {
    let matcher = PatternMatcher::new(
        &["*".to_string()],
        &["**/groovy-dsl/**".to_string()]
    ).unwrap();

    assert!(matcher.should_exclude("build/groovy-dsl/cache.jar"));
}
```

## Performance Considerations

- **Scanner**: Traverses directories only once per cleanup cycle
- **Matcher**: Pre-compiled glob patterns (via `GlobSet`) for fast matching
- **Memory**: Files collected in memory before deletion (acceptable for typical workspace sizes)

## Future Improvements

Potential enhancements that maintain single responsibility:

1. **Add `disk.rs`** - Extract disk monitoring logic from `cleaner.rs`
2. **Add `reporter.rs`** - Separate logging/metrics from cleanup logic
3. **Add `filter.rs`** - Advanced file filtering (size, age, etc.)

Each improvement would be a **new, focused component** rather than adding complexity to existing ones.
