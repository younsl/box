.PHONY: build release run dev install test fmt lint check deps clean help

# Default target
.DEFAULT_GOAL := help

# Build debug binary
build:
	@echo "Building debug binary..."
	cargo build

# Build optimized release binary
release:
	@echo "Building release binary..."
	cargo build --release

# Build and run with interactive prompts
run: build
	@echo "Running s3vget (interactive mode)..."
	./target/debug/s3vget

# Run with debug logging
dev: build
	@echo "Running with debug logging..."
	./target/debug/s3vget --log-level debug

# Install to ~/.cargo/bin/
install:
	@echo "Installing s3vget..."
	cargo install --path .

# Run tests
test:
	@echo "Running tests..."
	cargo test --verbose

# Format code
fmt:
	@echo "Formatting code..."
	cargo fmt

# Run clippy
lint:
	@echo "Running clippy..."
	cargo clippy -- -D warnings

# Check code without building
check:
	@echo "Checking code..."
	cargo check

# Update dependencies
deps:
	@echo "Updating dependencies..."
	cargo update

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cargo clean

# Show help
help:
	@echo "s3vget - S3 Object Version Downloader"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build debug binary (target/debug/)"
	@echo "  release        Build optimized release binary (target/release/)"
	@echo "  run            Build and run with interactive prompts"
	@echo "  dev            Run with debug logging"
	@echo "  install        Install to ~/.cargo/bin/"
	@echo "  test           Run tests"
	@echo "  fmt            Format code"
	@echo "  lint           Run clippy"
	@echo "  check          Check code without building"
	@echo "  deps           Update dependencies"
	@echo "  clean          Remove build artifacts"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run                                    # Interactive mode"
	@echo "  make dev                                    # Debug logging"
	@echo "  ./target/release/s3vget --help              # Show CLI help"
	@echo "  ./target/release/s3vget -b my-bucket -k path/to/file"
