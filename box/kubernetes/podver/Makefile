.PHONY: build clean test run help release install fmt docker-build docker-push

BINARY_NAME=podver
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

help: ## 도움말 출력
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## 바이너리 빌드 (debug)
	@echo "Building $(BINARY_NAME) (debug)..."
	@cargo build
	@echo "Binary built: target/debug/$(BINARY_NAME)"

release: ## 바이너리 빌드 (release)
	@echo "Building $(BINARY_NAME) (release)..."
	@cargo build --release
	@echo "Binary built: target/release/$(BINARY_NAME)"

clean: ## 빌드 아티팩트 정리
	@echo "Cleaning up..."
	@cargo clean

test: ## 테스트 실행
	@echo "Running tests..."
	@cargo test

run: build ## 빌드 후 실행
	@echo "Running $(BINARY_NAME)..."
	@cargo run -- --help

dev: ## 개발 모드로 실행 (verbose)
	@echo "Running $(BINARY_NAME) in dev mode..."
	@cargo run -- --verbose

install: release ## 시스템에 설치
	@echo "Installing $(BINARY_NAME) to ~/.cargo/bin/..."
	@cargo install --path .
	@echo "Installation complete"

fmt: ## 코드 포맷팅
	@echo "Formatting code..."
	@cargo fmt

lint: ## 코드 린팅
	@echo "Running linter..."
	@cargo clippy -- -D warnings

docker-build: ## Docker 이미지 빌드
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-push: ## Docker 이미지 푸시 (ECR)
	@echo "Pushing Docker image to ECR..."
	@echo "Configure AWS credentials and ECR repository first"
	@# aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com
	@# docker tag $(BINARY_NAME):$(VERSION) <account-id>.dkr.ecr.<region>.amazonaws.com/$(BINARY_NAME):$(VERSION)
	@# docker push <account-id>.dkr.ecr.<region>.amazonaws.com/$(BINARY_NAME):$(VERSION) 