BINARY_NAME := filesystem-cleaner
# Use environment variables if set, otherwise fallback to git/system values
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# Docker/ECR settings
DOCKER_IMAGE := filesystem-cleaner
DOCKER_TAG := $(VERSION)
ECR_REGISTRY := your-account.dkr.ecr.us-east-1.amazonaws.com
ECR_IMAGE := $(ECR_REGISTRY)/$(DOCKER_IMAGE)

# Rust build targets
RUST_TARGET_DIR := target
DEBUG_BIN := $(RUST_TARGET_DIR)/debug/$(BINARY_NAME)
RELEASE_BIN := $(RUST_TARGET_DIR)/release/$(BINARY_NAME)

.PHONY: all
all: fmt lint test build

.PHONY: build
build:
	cargo build

.PHONY: release
release:
	cargo build --release --locked

.PHONY: build-all
build-all:
	@echo "Building for all platforms (requires cross)..."
	@command -v cross >/dev/null 2>&1 || { echo "cross not installed. Install with: cargo install cross"; exit 1; }
	cross build --release --target x86_64-unknown-linux-gnu
	cross build --release --target aarch64-unknown-linux-gnu
	cross build --release --target x86_64-apple-darwin
	cross build --release --target aarch64-apple-darwin

.PHONY: run
run: build
	$(DEBUG_BIN) --dry-run --usage-threshold-percent 70 --target-paths /tmp

.PHONY: dev
dev: build
	$(DEBUG_BIN) --log-level debug --dry-run --usage-threshold-percent 70 --target-paths /tmp

.PHONY: test
test:
	cargo test --verbose

.PHONY: fmt
fmt:
	cargo fmt

.PHONY: check
check:
	cargo check

.PHONY: lint
lint:
	cargo clippy -- -D warnings

.PHONY: deps
deps:
	cargo update

.PHONY: clean
clean:
	cargo clean
	rm -f Cargo.lock

.PHONY: docker-build
docker-build:
	docker build \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE="$(DATE)" \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

.PHONY: docker-push
docker-push: docker-build
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(ECR_REGISTRY)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(ECR_IMAGE):$(DOCKER_TAG)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(ECR_IMAGE):latest
	docker push $(ECR_IMAGE):$(DOCKER_TAG)
	docker push $(ECR_IMAGE):latest

.PHONY: install
install: release
	cargo install --path . --locked

.PHONY: uninstall
uninstall:
	cargo uninstall $(BINARY_NAME)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Run fmt, lint, test, and build"
	@echo "  build       - Build debug binary (target/debug/)"
	@echo "  release     - Build optimized release binary (target/release/)"
	@echo "  build-all   - Build for all platforms (requires cross)"
	@echo "  run         - Build and run with default settings"
	@echo "  dev         - Run with debug logging"
	@echo "  test        - Run tests (cargo test --verbose)"
	@echo "  fmt         - Format code (cargo fmt)"
	@echo "  check       - Check code without building"
	@echo "  lint        - Run clippy (cargo clippy)"
	@echo "  deps        - Update dependencies (cargo update)"
	@echo "  clean       - Remove build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push - Push to ECR"
	@echo "  install     - Install to ~/.cargo/bin/"
	@echo "  uninstall   - Uninstall from ~/.cargo/bin/"