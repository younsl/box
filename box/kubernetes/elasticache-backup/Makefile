.PHONY: help build release run dev test fmt lint check clean install docker-build docker-push build-all deps

# Variables
BINARY_NAME := elasticache-backup
IMAGE_NAME := elasticache-backup
ECR_REGISTRY := <YOUR_ECR_REGISTRY>
VERSION := $(shell grep '^version' Cargo.toml | sed 's/version = "\(.*\)"/\1/')
PLATFORMS := linux/amd64,linux/arm64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build debug binary
	cargo build --verbose

release: ## Build optimized release binary
	cargo build --release --verbose

run: build ## Build and run with example (human-readable logs)
	LOG_FORMAT=pretty ./target/debug/$(BINARY_NAME) \
		--cache-cluster-id "example-redis-001" \
		--s3-bucket-name "example-elasticache-backups"

dev: build ## Run with verbose logging (human-readable)
	LOG_LEVEL=debug LOG_FORMAT=pretty ./target/debug/$(BINARY_NAME) \
		--cache-cluster-id "example-redis-001" \
		--s3-bucket-name "example-elasticache-backups"

run-json: build ## Run with JSON logs (for production)
	LOG_FORMAT=json ./target/debug/$(BINARY_NAME) \
		--cache-cluster-id "example-redis-001" \
		--s3-bucket-name "example-elasticache-backups"

test: ## Run tests
	cargo test --verbose

fmt: ## Format code
	cargo fmt

lint: ## Run clippy
	cargo clippy -- -D warnings

check: ## Check code without building
	cargo check

deps: ## Update dependencies
	cargo update

clean: ## Remove build artifacts
	cargo clean

install: release ## Install to ~/.cargo/bin/
	cargo install --path .

docker-build: ## Build Docker image for current platform
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

docker-build-multiarch: ## Build multi-arch Docker image (amd64, arm64)
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		--load \
		.

docker-push: ## Push to ECR (single arch)
	docker tag $(IMAGE_NAME):$(VERSION) $(ECR_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker tag $(IMAGE_NAME):latest $(ECR_REGISTRY)/$(IMAGE_NAME):latest
	docker push $(ECR_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(ECR_REGISTRY)/$(IMAGE_NAME):latest

docker-push-multiarch: ## Push multi-arch image to ECR
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(ECR_REGISTRY)/$(IMAGE_NAME):$(VERSION) \
		-t $(ECR_REGISTRY)/$(IMAGE_NAME):latest \
		--push \
		.

build-all: ## Build for all platforms (requires cross)
	@echo "Building for linux/amd64..."
	cargo build --release --target x86_64-unknown-linux-gnu
	@echo "Building for linux/arm64..."
	cargo build --release --target aarch64-unknown-linux-gnu
	@echo "Building for darwin/amd64..."
	cargo build --release --target x86_64-apple-darwin
	@echo "Building for darwin/arm64..."
	cargo build --release --target aarch64-apple-darwin
