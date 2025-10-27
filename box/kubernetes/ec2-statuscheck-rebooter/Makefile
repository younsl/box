.PHONY: build release run dev test fmt lint check clean install docker-build docker-push deploy help

# Binary name
BINARY_NAME := ec2-statuscheck-rebooter

# Docker configuration
IMAGE_NAME := ec2-statuscheck-rebooter
IMAGE_TAG := latest
ECR_REGISTRY := YOUR_ECR_REGISTRY
DOCKER_PLATFORMS := linux/amd64,linux/arm64

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build debug binary
	cargo build

release: ## Build optimized release binary
	cargo build --release

run: ## Build and run with example configuration
	@echo "Running with dry-run mode..."
	cargo build
	LOG_FORMAT=pretty ./target/debug/$(BINARY_NAME) \
		--check-interval-seconds 60 \
		--failure-threshold 2 \
		--dry-run

dev: ## Run with verbose debug logging
	@echo "Running with debug logging..."
	cargo build
	LOG_FORMAT=pretty LOG_LEVEL=debug ./target/debug/$(BINARY_NAME) \
		--check-interval-seconds 60 \
		--failure-threshold 2 \
		--dry-run

test: ## Run tests
	cargo test --verbose

fmt: ## Format code
	cargo fmt

lint: ## Run clippy linter
	cargo clippy -- -D warnings

check: ## Check code without building
	cargo check

deps: ## Update dependencies
	cargo update

clean: ## Remove build artifacts
	cargo clean

install: ## Install to ~/.cargo/bin/
	cargo install --path .

docker-build: ## Build Docker image
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):$(VERSION)

docker-push: ## Push Docker image to ECR
	@if [ "$(ECR_REGISTRY)" = "YOUR_ECR_REGISTRY" ]; then \
		echo "Error: Please update ECR_REGISTRY in Makefile"; \
		exit 1; \
	fi
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(ECR_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(ECR_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(ECR_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(ECR_REGISTRY)/$(IMAGE_NAME):$(VERSION)

docker-build-multi: ## Build multi-arch Docker image
	docker buildx build \
		--platform $(DOCKER_PLATFORMS) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):$(VERSION) \
		.

deploy: ## Deploy to Kubernetes via Helm
	helm upgrade --install ec2-statuscheck-rebooter ./charts/ec2-statuscheck-rebooter \
		--namespace monitoring \
		--create-namespace

all: fmt lint test build ## Run fmt, lint, test, and build
