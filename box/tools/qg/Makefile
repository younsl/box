# Binary name
BINARY_NAME=qg

# Default target: build
all: build

# Build the application (debug mode)
build:
	@echo "Building $(BINARY_NAME) (debug)..."
	@cargo build
	@echo "$(BINARY_NAME) built successfully (target/debug/$(BINARY_NAME))"

# Optimized release build
release:
	@echo "Building $(BINARY_NAME) for release..."
	@cargo build --release
	@echo "$(BINARY_NAME) built successfully (target/release/$(BINARY_NAME))"

# Build for all platforms (requires cross)
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@cargo build --release --target x86_64-unknown-linux-gnu
	@cargo build --release --target x86_64-apple-darwin
	@cargo build --release --target aarch64-apple-darwin
	@cargo build --release --target x86_64-pc-windows-gnu
	@echo "Multi-platform build completed"

# Run the application with example
run:
	@echo "Running $(BINARY_NAME) with example..."
	@cargo run -- https://github.com/

# Remove built artifacts
clean:
	@echo "Cleaning $(BINARY_NAME)..."
	@cargo clean
	@rm -f qrcode.png
	@echo "$(BINARY_NAME) cleaned."

# Run tests
test:
	@echo "Running tests..."
	@cargo test --verbose

# Format code
fmt:
	@echo "Formatting code..."
	@cargo fmt

# Run clippy (linter)
lint:
	@echo "Running clippy..."
	@cargo clippy -- -D warnings

# Check code without building
check:
	@echo "Checking code..."
	@cargo check

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@cargo update
	@echo "Dependencies updated."

# Install binary to cargo bin directory
install:
	@echo "Installing $(BINARY_NAME)..."
	@cargo install --path .
	@echo "$(BINARY_NAME) installed to ~/.cargo/bin/"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build      - Build the binary (debug mode)"
	@echo "  make release    - Build optimized release binary"
	@echo "  make build-all  - Build for all platforms (requires cross)"
	@echo "  make run        - Run with example URL"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run clippy linter"
	@echo "  make check      - Check code without building"
	@echo "  make deps       - Update dependencies"
	@echo "  make install    - Install to ~/.cargo/bin/"
	@echo "  make help       - Show this help"

# Define phony targets
.PHONY: all build release build-all run clean test fmt lint check deps install help
